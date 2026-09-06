package subscriber

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// stubRepo is a hand-written in-memory SubscriberRepository stub, matching the
// repo's dominant test pattern.
type stubRepo struct {
	mu    sync.Mutex
	items map[primitive.ObjectID]*models.Subscriber
}

func newStubRepo() *stubRepo {
	return &stubRepo{items: map[primitive.ObjectID]*models.Subscriber{}}
}

func (r *stubRepo) FindByEmail(_ context.Context, email string) (*models.Subscriber, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, sub := range r.items {
		if sub.Email == email {
			cp := *sub
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *stubRepo) FindByVerificationToken(_ context.Context, token string) (*models.Subscriber, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, sub := range r.items {
		if sub.VerificationToken != "" && sub.VerificationToken == token {
			cp := *sub
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *stubRepo) FindByUnsubscribeToken(_ context.Context, token string) (*models.Subscriber, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, sub := range r.items {
		if sub.UnsubscribeToken != "" && sub.UnsubscribeToken == token {
			cp := *sub
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *stubRepo) FindByID(_ context.Context, id primitive.ObjectID) (*models.Subscriber, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if sub, ok := r.items[id]; ok {
		cp := *sub
		return &cp, nil
	}
	return nil, nil
}

func (r *stubRepo) Insert(_ context.Context, sub models.Subscriber) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := sub
	r.items[sub.ID] = &cp
	return nil
}

func (r *stubRepo) Update(_ context.Context, sub models.Subscriber) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[sub.ID]; !ok {
		return assert.AnError
	}
	cp := sub
	r.items[sub.ID] = &cp
	return nil
}

func (r *stubRepo) List(_ context.Context, _, _ int) ([]models.Subscriber, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]models.Subscriber, 0, len(r.items))
	for _, sub := range r.items {
		out = append(out, *sub)
	}
	return out, int64(len(out)), nil
}

func (r *stubRepo) ListVerified(_ context.Context) ([]models.Subscriber, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []models.Subscriber
	for _, sub := range r.items {
		if sub.Verified && !sub.Unsubscribed {
			out = append(out, *sub)
		}
	}
	return out, nil
}

func (r *stubRepo) DeleteByID(_ context.Context, id primitive.ObjectID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; ok {
		delete(r.items, id)
		return true, nil
	}
	return false, nil
}

func (r *stubRepo) PruneUnverified(_ context.Context, olderThan time.Duration) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-olderThan)
	var removed int64
	for id, sub := range r.items {
		if !sub.Verified && sub.CreatedAt.Before(cutoff) {
			delete(r.items, id)
			removed++
		}
	}
	return removed, nil
}

func TestSubscribeNewEmailIsPendingWithTokens(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	sub, err := svc.Subscribe(context.Background(), "User@Example.COM ")
	require.NoError(t, err)

	assert.False(t, sub.Verified, "new subscription must be pending verification")
	assert.NotEmpty(t, sub.VerificationToken, "verification token required")
	assert.NotEmpty(t, sub.UnsubscribeToken, "unsubscribe token required")
	require.NotNil(t, sub.VerificationTokenExpiresAt)
	assert.WithinDuration(t, time.Now().Add(VerificationTokenTTL), *sub.VerificationTokenExpiresAt, time.Minute)
	assert.Equal(t, "user@example.com", sub.Email, "email must be normalized")

	stored, err := repo.FindByID(context.Background(), sub.ID)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, sub.VerificationToken, stored.VerificationToken)
}

func TestSubscribeActiveVerifiedConflicts(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	first, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	first.Verified = true
	now := time.Now()
	first.VerifiedAt = &now
	first.VerificationToken = ""
	require.NoError(t, repo.Update(context.Background(), first))

	_, err = svc.Subscribe(context.Background(), "user@example.com")
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrConflict)
}

func TestSubscribePendingRefreshesToken(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	first, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	oldToken := first.VerificationToken

	second, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.NotEqual(t, oldToken, second.VerificationToken, "pending re-subscribe must rotate token")
	assert.False(t, second.Verified)
}

func TestSubscribeUnsubscribedReactivates(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	first, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	first.Verified = true
	require.NoError(t, repo.Update(context.Background(), first))

	_, err = svc.Unsubscribe(context.Background(), first.UnsubscribeToken)
	require.NoError(t, err)

	reSub, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, first.ID, reSub.ID)
	assert.False(t, reSub.Verified, "reactivation requires fresh verification")
	assert.False(t, reSub.Unsubscribed)
	assert.NotEmpty(t, reSub.VerificationToken)
}

func TestVerifyRoundTrip(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	sub, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)

	verified, err := svc.Verify(context.Background(), sub.VerificationToken)
	require.NoError(t, err)
	assert.True(t, verified.Verified)
	require.NotNil(t, verified.VerifiedAt)
	assert.Empty(t, verified.VerificationToken, "token must be cleared after use")
	assert.Nil(t, verified.VerificationTokenExpiresAt)
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	sub, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	expired := time.Now().Add(-time.Hour)
	sub.VerificationTokenExpiresAt = &expired
	require.NoError(t, repo.Update(context.Background(), sub))

	_, err = svc.Verify(context.Background(), sub.VerificationToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
}

func TestVerifyRejectsUnknownToken(t *testing.T) {
	svc := NewService(newStubRepo())
	_, err := svc.Verify(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestUnsubscribeRoundTrip(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	sub, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	sub.Verified = true
	require.NoError(t, repo.Update(context.Background(), sub))

	unsubbed, err := svc.Unsubscribe(context.Background(), sub.UnsubscribeToken)
	require.NoError(t, err)
	assert.True(t, unsubbed.Unsubscribed)
	require.NotNil(t, unsubbed.UnsubscribedAt)

	// Unsubscribed subscribers must no longer be broadcast targets.
	targets, err := svc.ListVerified(context.Background())
	require.NoError(t, err)
	assert.Empty(t, targets)
}

func TestUnsubscribeRejectsUnknownToken(t *testing.T) {
	svc := NewService(newStubRepo())
	_, err := svc.Unsubscribe(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestResendVerification(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	sub, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	oldToken := sub.VerificationToken

	resent, err := svc.ResendVerification(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.NotEqual(t, oldToken, resent.VerificationToken)
	assert.False(t, resent.Verified)

	// Old token must no longer verify; the new one must.
	_, err = svc.Verify(context.Background(), oldToken)
	assert.Error(t, err, "old token should be invalidated by resend")
}

func TestResendVerificationRejectsVerified(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	sub, err := svc.Subscribe(context.Background(), "user@example.com")
	require.NoError(t, err)
	sub.Verified = true
	require.NoError(t, repo.Update(context.Background(), sub))

	_, err = svc.ResendVerification(context.Background(), sub.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrConflict)
}

func TestPruneUnverifiedKeepsVerified(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo)

	stale, err := svc.Subscribe(context.Background(), "stale@example.com")
	require.NoError(t, err)
	stale.CreatedAt = time.Now().Add(-72 * time.Hour)
	require.NoError(t, repo.Update(context.Background(), stale))

	fresh, err := svc.Subscribe(context.Background(), "fresh@example.com")
	require.NoError(t, err)
	require.False(t, fresh.Verified)

	verifiedOld, err := svc.Subscribe(context.Background(), "verified@example.com")
	require.NoError(t, err)
	verifiedOld.Verified = true
	verifiedOld.CreatedAt = time.Now().Add(-72 * time.Hour)
	require.NoError(t, repo.Update(context.Background(), verifiedOld))

	removed, err := svc.PruneUnverified(context.Background(), UnverifiedRetention)
	require.NoError(t, err)
	assert.Equal(t, int64(1), removed, "only stale unverified subscriber should be pruned")

	_, err = repo.FindByID(context.Background(), stale.ID)
	require.NoError(t, err)
	storedStale, _ := repo.FindByID(context.Background(), stale.ID)
	assert.Nil(t, storedStale, "stale unverified subscriber must be deleted")

	keptVerified, err := repo.FindByID(context.Background(), verifiedOld.ID)
	require.NoError(t, err)
	assert.NotNil(t, keptVerified, "verified subscribers must never be pruned")

	keptFresh, err := repo.FindByID(context.Background(), fresh.ID)
	require.NoError(t, err)
	assert.NotNil(t, keptFresh, "recent unverified subscribers must be kept")
}

func TestTokenLength(t *testing.T) {
	token := repository.NewVerificationToken()
	assert.Equal(t, 64, len(token), "token should be 32 bytes hex-encoded")
	assert.True(t, isHex(token))
}

func isHex(s string) bool {
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}
