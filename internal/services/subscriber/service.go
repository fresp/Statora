package subscriber

import (
	"context"
	"fmt"
	"time"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/fresp/Statora/internal/security/pii"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// VerificationTokenTTL is how long a verification link stays valid.
	VerificationTokenTTL = 48 * time.Hour
	// UnverifiedRetention is how long unverified subscribers are kept before pruning.
	UnverifiedRetention = 48 * time.Hour
)

// Service implements the double opt-in subscriber state machine.
type Service struct {
	repo repository.SubscriberRepository
}

func NewService(repo repository.SubscriberRepository) *Service {
	return &Service{repo: repo}
}

// Subscribe normalizes the email and applies the double opt-in state machine:
// verified+active -> conflict; pending -> refresh token; unsubscribed -> re-activate;
// new -> insert with fresh tokens. The returned subscriber carries the current
// verification token so the caller can send the email.
func (s *Service) Subscribe(ctx context.Context, email string) (models.Subscriber, error) {
	email = pii.Normalize(email)
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return models.Subscriber{}, err
	}

	if existing == nil {
		sub := repository.NewSubscriberWithTokens(email)
		if err := s.repo.Insert(ctx, sub); err != nil {
			return models.Subscriber{}, err
		}
		return sub, nil
	}

	if existing.Verified && !existing.Unsubscribed {
		return models.Subscriber{}, fmt.Errorf("%w: email already subscribed", shared.ErrConflict)
	}

	if !existing.Verified {
		// Pending verification: refresh the token so the newest link works.
		existing.VerificationToken = repository.NewVerificationToken()
		expires := time.Now().Add(VerificationTokenTTL)
		existing.VerificationTokenExpiresAt = &expires
	} else {
		// Previously unsubscribed: re-activate with a fresh opt-in.
		existing.Verified = false
		existing.VerifiedAt = nil
		existing.VerificationToken = repository.NewVerificationToken()
		expires := time.Now().Add(VerificationTokenTTL)
		existing.VerificationTokenExpiresAt = &expires
		existing.Unsubscribed = false
		existing.UnsubscribedAt = nil
	}

	if err := s.repo.Update(ctx, *existing); err != nil {
		return models.Subscriber{}, err
	}
	return *existing, nil
}

// Verify consumes a verification token: expired or unknown tokens are rejected,
// valid tokens mark the subscriber verified and clear the token.
func (s *Service) Verify(ctx context.Context, token string) (models.Subscriber, error) {
	sub, err := s.repo.FindByVerificationToken(ctx, token)
	if err != nil {
		return models.Subscriber{}, err
	}
	if sub == nil {
		return models.Subscriber{}, fmt.Errorf("%w: unknown verification token", shared.ErrNotFound)
	}
	if sub.VerificationTokenExpiresAt != nil && time.Now().After(*sub.VerificationTokenExpiresAt) {
		return models.Subscriber{}, fmt.Errorf("%w: verification token expired", shared.ErrInvalidInput)
	}

	sub.Verified = true
	now := time.Now()
	sub.VerifiedAt = &now
	sub.VerificationToken = ""
	sub.VerificationTokenExpiresAt = nil

	if err := s.repo.Update(ctx, *sub); err != nil {
		return models.Subscriber{}, err
	}
	return *sub, nil
}

// Unsubscribe consumes an unsubscribe token and marks the subscriber unsubscribed.
// The record is kept so the same email can re-subscribe later.
func (s *Service) Unsubscribe(ctx context.Context, token string) (models.Subscriber, error) {
	sub, err := s.repo.FindByUnsubscribeToken(ctx, token)
	if err != nil {
		return models.Subscriber{}, err
	}
	if sub == nil {
		return models.Subscriber{}, fmt.Errorf("%w: unknown unsubscribe token", shared.ErrNotFound)
	}

	sub.Unsubscribed = true
	now := time.Now()
	sub.UnsubscribedAt = &now

	if err := s.repo.Update(ctx, *sub); err != nil {
		return models.Subscriber{}, err
	}
	return *sub, nil
}

// ResendVerification issues a fresh verification token for a pending subscriber.
func (s *Service) ResendVerification(ctx context.Context, id primitive.ObjectID) (models.Subscriber, error) {
	sub, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return models.Subscriber{}, err
	}
	if sub == nil {
		return models.Subscriber{}, fmt.Errorf("%w: subscriber not found", shared.ErrNotFound)
	}
	if sub.Verified && !sub.Unsubscribed {
		return models.Subscriber{}, fmt.Errorf("%w: subscriber already verified", shared.ErrConflict)
	}

	sub.Verified = false
	sub.VerificationToken = repository.NewVerificationToken()
	expires := time.Now().Add(VerificationTokenTTL)
	sub.VerificationTokenExpiresAt = &expires

	if err := s.repo.Update(ctx, *sub); err != nil {
		return models.Subscriber{}, err
	}
	return *sub, nil
}

// List returns a paginated subscriber list (admin view).
func (s *Service) List(ctx context.Context, page, limit int) ([]models.Subscriber, int64, error) {
	return s.repo.List(ctx, page, limit)
}

// ListVerified returns all verified, still-subscribed recipients for a broadcast.
func (s *Service) ListVerified(ctx context.Context) ([]models.Subscriber, error) {
	return s.repo.ListVerified(ctx)
}

// DeleteByID removes a subscriber record (admin action).
func (s *Service) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	deleted, err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("%w: subscriber not found", shared.ErrNotFound)
	}
	return nil
}

// PruneUnverified deletes unverified subscribers older than the retention window.
func (s *Service) PruneUnverified(ctx context.Context, olderThan time.Duration) (int64, error) {
	return s.repo.PruneUnverified(ctx, olderThan)
}
