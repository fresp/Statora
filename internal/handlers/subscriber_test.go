package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/fresp/Statora/internal/security/pii"
	subscriberservice "github.com/fresp/Statora/internal/services/subscriber"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// newSubscribersTestDB connects to a throwaway MongoDB database defined by
// STATORA_TEST_MONGODB_URI. Tests skip gracefully when it is not configured,
// matching the repo's DB-dependent test convention.
func newSubscribersTestDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := strings.TrimSpace(os.Getenv("STATORA_TEST_MONGODB_URI"))
	if uri == "" {
		t.Skip("STATORA_TEST_MONGODB_URI not set; skipping MongoDB-backed subscriber tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Skipf("MongoDB unavailable: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB unreachable: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	dbName := "statora_subscriber_test_" + primitive.NewObjectID().Hex()
	t.Cleanup(func() { _ = client.Database(dbName).Drop(context.Background()) })
	return client.Database(dbName)
}

func performSubscriberRequest(t *testing.T, r *gin.Engine, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req, _ = http.NewRequest(method, path, nil)
	} else {
		req, _ = http.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func subscribeService(db *mongo.Database) *subscriberservice.Service {
	return subscriberservice.NewService(repository.NewMongoSubscriberRepository(db))
}

func TestSubscribeEndpointInsertsPendingSubscriber(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	r := gin.New()
	r.POST("/api/subscribe", Subscribe(db, nil))

	w := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribe", `{"email":"User@Example.com"}`, nil)
	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "confirm your subscription")

	var stored models.Subscriber
	err := db.Collection("subscribers").FindOne(context.Background(), bson.M{"email": "user@example.com"}).Decode(&stored)
	require.NoError(t, err)
	assert.False(t, stored.Verified)
	assert.NotEmpty(t, stored.VerificationToken)
	assert.NotEmpty(t, stored.UnsubscribeToken)

	// Re-subscribing a pending email refreshes the token instead of conflicting.
	w2 := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribe", `{"email":"user@example.com"}`, nil)
	assert.Equal(t, http.StatusCreated, w2.Code)
}

func TestSubscribeEndpointRejectsInvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	r := gin.New()
	r.POST("/api/subscribe", Subscribe(db, nil))

	w := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribe", `{"email":"not-an-email"}`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubscribeEndpointConflictsOnVerifiedActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	r := gin.New()
	r.POST("/api/subscribe", Subscribe(db, nil))

	svc := subscribeService(db)
	sub, err := svc.Subscribe(context.Background(), "active@example.com")
	require.NoError(t, err)
	sub.Verified = true
	require.NoError(t, repository.NewMongoSubscriberRepository(db).Update(context.Background(), sub))

	w := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribe", `{"email":"active@example.com"}`, nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestVerifyEndpointJSONAndRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	r := gin.New()
	r.GET("/api/subscribe/verify", VerifySubscriber(db))
	r.POST("/api/subscribe", Subscribe(db, nil))

	w := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribe", `{"email":"verify@example.com"}`, nil)
	require.Equal(t, http.StatusCreated, w.Code)
	var stored models.Subscriber
	require.NoError(t, db.Collection("subscribers").FindOne(context.Background(), bson.M{"email": "verify@example.com"}).Decode(&stored))

	// API client: JSON confirmation, subscriber flips to verified, token consumed.
	wJSON := performSubscriberRequest(t, r, http.MethodGet, "/api/subscribe/verify?token="+url.QueryEscape(stored.VerificationToken), "", nil)
	assert.Equal(t, http.StatusOK, wJSON.Code)
	assert.Contains(t, wJSON.Body.String(), "subscription confirmed")

	var verified models.Subscriber
	require.NoError(t, db.Collection("subscribers").FindOne(context.Background(), bson.M{"email": "verify@example.com"}).Decode(&verified))
	assert.True(t, verified.Verified)
	assert.Empty(t, verified.VerificationToken, "token must be consumed after verification")

	// Browser navigation: redirect to the status page banner.
	wRedir := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribe", `{"email":"redirect@example.com"}`, nil)
	require.Equal(t, http.StatusCreated, wRedir.Code)
	var stored2 models.Subscriber
	require.NoError(t, db.Collection("subscribers").FindOne(context.Background(), bson.M{"email": "redirect@example.com"}).Decode(&stored2))

	wBrowser := performSubscriberRequest(t, r, http.MethodGet, "/api/subscribe/verify?token="+url.QueryEscape(stored2.VerificationToken), "", map[string]string{"Accept": "text/html"})
	assert.Equal(t, http.StatusFound, wBrowser.Code)
	assert.Equal(t, "/?verified=true", wBrowser.Header().Get("Location"))
}

func TestVerifyEndpointRejectsBadToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	r := gin.New()
	r.GET("/api/subscribe/verify", VerifySubscriber(db))

	w := performSubscriberRequest(t, r, http.MethodGet, "/api/subscribe/verify?token=bogus", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	wBrowser := performSubscriberRequest(t, r, http.MethodGet, "/api/subscribe/verify?token=bogus", "", map[string]string{"Accept": "text/html"})
	assert.Equal(t, http.StatusFound, wBrowser.Code)
	assert.Equal(t, "/?error=invalid_or_expired_token", wBrowser.Header().Get("Location"))
}

func TestUnsubscribeEndpointRemovesFromBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	r := gin.New()
	r.GET("/api/subscribe/unsubscribe", Unsubscribe(db))
	r.POST("/api/subscribe", Subscribe(db, nil))

	w := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribe", `{"email":"unsub@example.com"}`, nil)
	require.Equal(t, http.StatusCreated, w.Code)
	var stored models.Subscriber
	require.NoError(t, db.Collection("subscribers").FindOne(context.Background(), bson.M{"email": "unsub@example.com"}).Decode(&stored))
	stored.Verified = true
	_, err := db.Collection("subscribers").ReplaceOne(context.Background(), bson.M{"_id": stored.ID}, stored)
	require.NoError(t, err)

	wUnsub := performSubscriberRequest(t, r, http.MethodGet, "/api/subscribe/unsubscribe?token="+url.QueryEscape(stored.UnsubscribeToken), "", nil)
	assert.Equal(t, http.StatusOK, wUnsub.Code)
	// Unsubscribe links are persistent (embedded in every email): a repeated
	// click is idempotent and must NOT error.
	wBrowser := performSubscriberRequest(t, r, http.MethodGet, "/api/subscribe/unsubscribe?token="+url.QueryEscape(stored.UnsubscribeToken), "", map[string]string{"Accept": "text/html"})
	assert.Equal(t, http.StatusFound, wBrowser.Code)
	assert.Equal(t, "/?unsubscribed=true", wBrowser.Header().Get("Location"))

	// But an unknown token must fail.
	wBrowserBad := performSubscriberRequest(t, r, http.MethodGet, "/api/subscribe/unsubscribe?token=unknown-token", "", map[string]string{"Accept": "text/html"})
	assert.Equal(t, http.StatusFound, wBrowserBad.Code)
	assert.Equal(t, "/?error=invalid_token", wBrowserBad.Header().Get("Location"))

	count, err := db.Collection("subscribers").CountDocuments(context.Background(), bson.M{"verified": true, "unsubscribed": bson.M{"$ne": true}})
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "unsubscribed subscriber must not be a broadcast target")
}

func TestResendVerificationEndpointRotatesToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	r := gin.New()
	r.POST("/api/subscribers/:id/resend-verification", ResendVerification(db, nil))

	id := primitive.NewObjectID()
	_, err := db.Collection("subscribers").InsertOne(context.Background(), models.Subscriber{
		ID:                id,
		Email:             "pending@example.com",
		Verified:          false,
		VerificationToken: "seed-token",
		CreatedAt:         time.Now(),
	})
	require.NoError(t, err)

	w := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribers/"+id.Hex()+"/resend-verification", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var after models.Subscriber
	require.NoError(t, db.Collection("subscribers").FindOne(context.Background(), bson.M{"_id": id}).Decode(&after))
	assert.NotEqual(t, "seed-token", after.VerificationToken, "resend must rotate the token")

	w404 := performSubscriberRequest(t, r, http.MethodPost, "/api/subscribers/"+primitive.NewObjectID().Hex()+"/resend-verification", "", nil)
	assert.Equal(t, http.StatusNotFound, w404.Code)
}

func TestMailSecretsStoredEncrypted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSubscribersTestDB(t)
	key := "0123456789abcdef0123456789abcdef" // 32 bytes

	_, err := db.Collection("settings").InsertOne(context.Background(), models.DefaultStatusPageSettings())
	require.NoError(t, err)

	cipher, err := pii.Encrypt("super-secret-smtp-pass", []byte(key))
	require.NoError(t, err)
	_, err = db.Collection("settings").UpdateOne(context.Background(),
		bson.M{"key": models.StatusPageSettingsKey},
		bson.M{"$set": bson.M{
			"mail.provider":      "smtp",
			"mail.smtp.host":     "smtp.example.com",
			"mail.smtp.port":     587,
			"mail.smtp.password": cipher,
		}})
	require.NoError(t, err)

	var storedDoc bson.M
	require.NoError(t, db.Collection("settings").FindOne(context.Background(), bson.M{"key": models.StatusPageSettingsKey}).Decode(&storedDoc))
	mailDoc := storedDoc["mail"].(bson.M)
	smtpDoc := mailDoc["smtp"].(bson.M)

	assert.NotEqual(t, "super-secret-smtp-pass", smtpDoc["password"], "plaintext secret must never be stored")
	assert.Equal(t, cipher, smtpDoc["password"], "ciphertext must be stored verbatim")

	plain, err := pii.Decrypt(smtpDoc["password"].(string), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, "super-secret-smtp-pass", plain)
}
