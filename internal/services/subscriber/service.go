package subscriber

import (
	"context"
	"fmt"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/fresp/Statora/internal/security/pii"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repo     repository.SubscriberRepository
	emailKey []byte
}

func NewService(repo repository.SubscriberRepository, emailEncryptionKey string) *Service {
	return &Service{repo: repo, emailKey: []byte(emailEncryptionKey)}
}

func (s *Service) Create(ctx context.Context, email string) (models.Subscriber, error) {
	processedEmail, err := pii.Process(email, s.emailKey)
	if err != nil {
		return models.Subscriber{}, err
	}
	if processedEmail.Normalized == "" {
		return models.Subscriber{}, fmt.Errorf("%w: email is required", shared.ErrInvalidInput)
	}

	existing, err := s.repo.FindByEmailHash(ctx, processedEmail.Hash)
	if err != nil {
		return models.Subscriber{}, err
	}
	if existing != nil {
		return models.Subscriber{}, fmt.Errorf("%w: email already subscribed", shared.ErrConflict)
	}

	sub := repository.NewSubscriber(processedEmail.Encrypted, processedEmail.Hash)
	if err := s.repo.Insert(ctx, sub); err != nil {
		return models.Subscriber{}, err
	}

	sub.Email = processedEmail.Normalized
	return sub, nil
}

func (s *Service) List(ctx context.Context, page, limit int) ([]models.Subscriber, int64, error) {
	subs, total, err := s.repo.List(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}

	for i := range subs {
		decryptedEmail, err := pii.Decrypt(subs[i].Email, s.emailKey)
		if err != nil {
			return nil, 0, err
		}
		subs[i].Email = decryptedEmail
	}

	return subs, total, nil
}

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
