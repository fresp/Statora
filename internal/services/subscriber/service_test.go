package subscriber

import (
	"context"
	"testing"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/security/pii"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const subscriberTestKey = "01234567890123456789012345678901"

type stubSubscriberRepo struct {
	existingByHash      map[string]*models.Subscriber
	lastFindByEmailHash string
	inserted            []models.Subscriber
	insertErr           error
	listResult          []models.Subscriber
	listTotal           int64
	listErr             error
	deleteResult        bool
	deleteErr           error
}

func (r *stubSubscriberRepo) FindByEmailHash(_ context.Context, emailHash string) (*models.Subscriber, error) {
	r.lastFindByEmailHash = emailHash
	if existing, ok := r.existingByHash[emailHash]; ok {
		copy := *existing
		return &copy, nil
	}

	return nil, nil
}

func (r *stubSubscriberRepo) Insert(_ context.Context, sub models.Subscriber) error {
	r.inserted = append(r.inserted, sub)
	return r.insertErr
}

func (r *stubSubscriberRepo) List(_ context.Context, page, limit int) ([]models.Subscriber, int64, error) {
	result := make([]models.Subscriber, len(r.listResult))
	copy(result, r.listResult)
	return result, r.listTotal, r.listErr
}

func (r *stubSubscriberRepo) DeleteByID(_ context.Context, _ primitive.ObjectID) (bool, error) {
	return r.deleteResult, r.deleteErr
}

func TestCreateEncryptsAndStoresHashWhileReturningNormalizedPlaintextEmail(t *testing.T) {
	t.Parallel()

	repo := &stubSubscriberRepo{}
	svc := NewService(repo, subscriberTestKey)

	created, err := svc.Create(context.Background(), "  USER@Example.COM  ")
	require.NoError(t, err)
	require.Len(t, repo.inserted, 1)

	normalized := "user@example.com"
	expectedHash := pii.Hash(normalized)
	stored := repo.inserted[0]

	assert.Equal(t, normalized, created.Email)
	assert.Equal(t, expectedHash, created.EmailHash)
	assert.Equal(t, expectedHash, repo.lastFindByEmailHash)
	assert.Equal(t, expectedHash, stored.EmailHash)
	assert.NotEqual(t, normalized, stored.Email)

	decrypted, err := pii.Decrypt(stored.Email, []byte(subscriberTestKey))
	require.NoError(t, err)
	assert.Equal(t, normalized, decrypted)
}

func TestCreateDetectsDuplicatesByHashEvenWithCaseWhitespaceVariants(t *testing.T) {
	t.Parallel()

	normalized := "user@example.com"
	expectedHash := pii.Hash(normalized)
	repo := &stubSubscriberRepo{
		existingByHash: map[string]*models.Subscriber{
			expectedHash: {ID: primitive.NewObjectID(), EmailHash: expectedHash},
		},
	}
	svc := NewService(repo, subscriberTestKey)

	_, err := svc.Create(context.Background(), "  USER@example.COM  ")
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrConflict)
	assert.Equal(t, expectedHash, repo.lastFindByEmailHash)
	assert.Empty(t, repo.inserted)
}

func TestCreateRejectsBlankNormalizedEmail(t *testing.T) {
	t.Parallel()

	repo := &stubSubscriberRepo{}
	svc := NewService(repo, subscriberTestKey)

	_, err := svc.Create(context.Background(), "   ")
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Empty(t, repo.inserted)
}

func TestListDecryptsStoredEmailsForDisplay(t *testing.T) {
	t.Parallel()

	encrypted, err := pii.Encrypt("admin@example.com", []byte(subscriberTestKey))
	require.NoError(t, err)

	repo := &stubSubscriberRepo{
		listResult: []models.Subscriber{{
			ID:        primitive.NewObjectID(),
			Email:     encrypted,
			EmailHash: pii.Hash("admin@example.com"),
			Verified:  true,
		}},
		listTotal: 1,
	}
	svc := NewService(repo, subscriberTestKey)

	subs, total, err := svc.List(context.Background(), 1, 20)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "admin@example.com", subs[0].Email)
	assert.Equal(t, encrypted, repo.listResult[0].Email)
}

func TestDeleteByIDReturnsErrNotFoundWhenRepoReportsNoDeletion(t *testing.T) {
	t.Parallel()

	repo := &stubSubscriberRepo{}
	svc := NewService(repo, subscriberTestKey)

	err := svc.DeleteByID(context.Background(), primitive.NewObjectID())
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrNotFound)
}
