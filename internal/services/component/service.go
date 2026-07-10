package component

import (
	"context"
	"fmt"
	"strings"
	"time"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const maxComponentNameLength = 120

type BroadcastFunc func(eventType string, data any)

type Service struct {
	repo      repository.ComponentRepository
	broadcast BroadcastFunc
}

func NewService(repo repository.ComponentRepository, broadcast BroadcastFunc) *Service {
	return &Service{repo: repo, broadcast: broadcast}
}

type CreateInput struct {
	Name        string
	Description string
	Status      models.ComponentStatus
}

type UpdateInput struct {
	Name        string
	Description string
	Status      models.ComponentStatus
}

func (s *Service) List(ctx context.Context, page, limit int) ([]models.Component, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (models.Component, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return models.Component{}, fmt.Errorf("%w: component name is required", shared.ErrInvalidInput)
	}
	if len(name) > maxComponentNameLength {
		return models.Component{}, fmt.Errorf("%w: component name exceeds %d characters", shared.ErrInvalidInput, maxComponentNameLength)
	}

	status := input.Status
	if status == "" {
		status = models.StatusOperational
	}

	now := time.Now()
	component := models.Component{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: input.Description,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Insert(ctx, component); err != nil {
		return models.Component{}, err
	}

	s.broadcastEvent("component_created", component)

	return component, nil
}

func (s *Service) Update(ctx context.Context, id primitive.ObjectID, input UpdateInput) (models.Component, error) {
	if input.Name != "" && strings.TrimSpace(input.Name) == "" {
		return models.Component{}, fmt.Errorf("%w: component name cannot be blank", shared.ErrInvalidInput)
	}
	if len(strings.TrimSpace(input.Name)) > maxComponentNameLength {
		return models.Component{}, fmt.Errorf("%w: component name exceeds %d characters", shared.ErrInvalidInput, maxComponentNameLength)
	}

	setFields := bson.M{"updatedAt": time.Now()}
	if input.Name != "" {
		setFields["name"] = strings.TrimSpace(input.Name)
	}
	if input.Description != "" {
		setFields["description"] = input.Description
	}
	if input.Status != "" {
		setFields["status"] = input.Status
	}

	component, err := s.repo.UpdateByID(ctx, id, setFields)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Component{}, fmt.Errorf("%w: component not found", shared.ErrNotFound)
		}
		return models.Component{}, err
	}

	s.broadcastEvent("component_updated", component)

	return component, nil
}

func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (models.Component, error) {
	component, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Component{}, fmt.Errorf("%w: component not found", shared.ErrNotFound)
		}
		return models.Component{}, err
	}

	return component, nil
}

func (s *Service) Delete(ctx context.Context, id primitive.ObjectID) error {
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return err
	}

	s.broadcastEvent("component_deleted", map[string]string{"id": id.Hex()})

	return nil
}

func (s *Service) broadcastEvent(eventType string, data any) {
	if s.broadcast != nil {
		s.broadcast(eventType, data)
	}
}
