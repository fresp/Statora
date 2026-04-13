package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/fresp/Statora/internal/models"
)

type stubUserSearchService struct {
	user      *models.User
	err       error
	lastEmail string
}

func (s *stubUserSearchService) FindUserByEmail(_ context.Context, email string) (*models.User, error) {
	s.lastEmail = email
	if s.err != nil {
		return nil, s.err
	}

	return s.user, nil
}

func TestMapUserDefaults(t *testing.T) {
	user := models.User{
		ID:       primitive.NewObjectID(),
		Username: "admin",
		Email:    "admin@example.com",
	}

	mapped := mapUser(user)

	assert.Equal(t, user.ID.Hex(), mapped.ID)
	assert.Equal(t, "admin", mapped.Username)
	assert.Equal(t, "admin@example.com", mapped.Email)
	assert.Equal(t, "admin", mapped.Role)
	assert.Equal(t, "active", mapped.Status)
}

func TestPatchUserRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/api/users/:id", PatchUser(nil))

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req, _ := http.NewRequest(http.MethodPatch, "/api/users/not-an-object-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUserInvitationRejectsInvalidRoleBeforeDBAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/users/invitations", CreateUserInvitation(nil, nil))

	body, _ := json.Marshal(map[string]string{
		"email": "member@example.com",
		"role":  "superadmin",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/users/invitations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestActivateUserInvitationRejectsInvalidPayloadBeforeDBAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/users/invitations/activate", ActivateUserInvitation(nil, nil))

	body, _ := json.Marshal(map[string]string{
		"token":    "",
		"username": "member-user",
		"password": "short",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/users/invitations/activate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestActivateUserInvitationRejectsBlankTokenAfterTrim(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/users/invitations/activate", ActivateUserInvitation(nil, nil))

	body, _ := json.Marshal(map[string]string{
		"token":    "   ",
		"username": "member-user",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/users/invitations/activate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBuildActivationLinkUsesCurrentHostAndToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:18081/api/users/invitations", nil)
	req.Host = "localhost:18081"
	c.Request = req

	link := buildActivationLink(c, "abc123")
	assert.Equal(t, "http://localhost:18081/admin/activate?token=abc123", link)
}

func TestGetUserInvitationsMapsInvitationModel(t *testing.T) {
	now := time.Now()
	inv := models.UserInvitation{
		ID:        primitive.NewObjectID(),
		Email:     "invited@example.com",
		Role:      "operator",
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
	}

	resp := userInvitationResponse{
		ID:        inv.ID.Hex(),
		Email:     inv.Email,
		Role:      inv.Role,
		ExpiresAt: inv.ExpiresAt,
		CreatedAt: inv.CreatedAt,
		IsExpired: inv.ExpiresAt.Before(now),
	}

	assert.Equal(t, inv.ID.Hex(), resp.ID)
	assert.Equal(t, "invited@example.com", resp.Email)
	assert.Equal(t, "operator", resp.Role)
	assert.Equal(t, inv.ExpiresAt, resp.ExpiresAt)
	assert.Equal(t, inv.CreatedAt, resp.CreatedAt)
	assert.False(t, resp.IsExpired)
}

func TestDeleteUserRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/api/users/:id", DeleteUser(nil))

	req, _ := http.NewRequest(http.MethodDelete, "/api/users/not-an-object-id", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserRejectsSelfDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userId", "507f1f77bcf86cd799439011")
		c.Next()
	})
	r.DELETE("/api/users/:id", DeleteUser(nil))

	req, _ := http.NewRequest(http.MethodDelete, "/api/users/507f1f77bcf86cd799439011", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchUserByEmailRequiresEmailQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/users/search", searchUserByEmailWithService(&stubUserSearchService{}))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/search", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchUserByEmailReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/users/search", searchUserByEmailWithService(&stubUserSearchService{err: mongo.ErrNoDocuments}))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/search?email=missing@example.com", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSearchUserByEmailReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/users/search", searchUserByEmailWithService(&stubUserSearchService{err: errors.New("boom")}))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/search?email=alice@example.com", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchUserByEmailReturnsMappedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	stub := &stubUserSearchService{user: &models.User{
		ID:       primitive.NewObjectID(),
		Username: "alice",
		Email:    "alice@example.com",
		Role:     "operator",
		Status:   "active",
		SSO: models.UserSSO{
			Enabled: true,
		},
	}}

	r.GET("/api/v1/users/search", searchUserByEmailWithService(stub))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/search?email=%20Alice@Example.com%20", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Alice@Example.com", stub.lastEmail)

	var body userResponse
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, stub.user.ID.Hex(), body.ID)
	assert.Equal(t, "alice", body.Username)
	assert.Equal(t, "alice@example.com", body.Email)
	assert.Equal(t, "operator", body.Role)
	assert.Equal(t, "active", body.Status)
	assert.True(t, body.SSOEnabled)
}
