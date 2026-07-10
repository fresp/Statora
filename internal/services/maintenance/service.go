package maintenance

import (
	"context"
	"errors"
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

type Service struct {
	repo repository.MaintenanceRepository
}

func NewService(repo repository.MaintenanceRepository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Title           string
	Description     string
	DescriptionJSON models.RichTextDocument
	VisibilityState models.IncidentVisibilityState
	Components      []string
	StartTime       string
	EndTime         string
	CreatorIDHex    string
	CreatorUsername string
}

type UpdateInput struct {
	Title       string
	Description string
	DescriptionJSON models.RichTextDocument
	VisibilityState models.IncidentVisibilityState
	Status      models.MaintenanceStatus
	StartTime   string
	EndTime     string
}

func (s *Service) List(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *Service) ListPublic(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error) {
	return s.repo.ListPublic(ctx, page, limit)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (models.Maintenance, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return models.Maintenance{}, fmt.Errorf("%w: title is required", shared.ErrInvalidInput)
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		return models.Maintenance{}, fmt.Errorf("%w: invalid startTime format, use RFC3339", shared.ErrInvalidInput)
	}
	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		return models.Maintenance{}, fmt.Errorf("%w: invalid endTime format, use RFC3339", shared.ErrInvalidInput)
	}
	if !startTime.Before(endTime) {
		return models.Maintenance{}, fmt.Errorf("%w: startTime must be before endTime", shared.ErrInvalidInput)
	}

	creatorID, err := primitive.ObjectIDFromHex(input.CreatorIDHex)
	if err != nil {
		return models.Maintenance{}, fmt.Errorf("%w: invalid authenticated user id", shared.ErrUnauthorized)
	}

	componentIDs := make([]primitive.ObjectID, 0, len(input.Components))
	for _, raw := range input.Components {
		oid, parseErr := primitive.ObjectIDFromHex(raw)
		if parseErr != nil {
			return models.Maintenance{}, fmt.Errorf("%w: invalid component id", shared.ErrInvalidInput)
		}
		componentIDs = append(componentIDs, oid)
	}

	status := models.MaintenanceScheduled
	if time.Now().After(startTime) {
		status = models.MaintenanceActive
	}
	if input.VisibilityState == models.IncidentVisibilityDraft {
		status = models.MaintenanceDraft
	}

	plainDescription := input.Description
	if derived := derivePlainText(input.DescriptionJSON); derived != "" {
		plainDescription = derived
	}

	maintenance := models.Maintenance{
		ID:              primitive.NewObjectID(),
		Title:           title,
		Description:     plainDescription,
		DescriptionJSON: input.DescriptionJSON,
		CreatorID:       &creatorID,
		CreatorUsername: input.CreatorUsername,
		Components:      componentIDs,
		StartTime:       startTime,
		EndTime:         endTime,
		Status:          status,
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Insert(ctx, maintenance); err != nil {
		return models.Maintenance{}, err
	}

	if err := s.repo.InsertAuditLog(ctx, models.AuditLog{
		ID:            primitive.NewObjectID(),
		ResourceType:  models.AuditResourceMaintenance,
		ResourceID:    maintenance.ID,
		Action:        auditActionForMaintenance(status),
		ActorID:       &creatorID,
		ActorUsername: input.CreatorUsername,
		At:            time.Now(),
		Summary:       "Maintenance created",
	}); err != nil {
		return models.Maintenance{}, err
	}

	return maintenance, nil
}

func (s *Service) Update(ctx context.Context, id primitive.ObjectID, input UpdateInput) (models.Maintenance, error) {
	setFields := bson.M{}
	if input.Title != "" {
		setFields["title"] = input.Title
	}
	if input.Description != "" {
		setFields["description"] = input.Description
	}
	if input.DescriptionJSON != nil {
		setFields["descriptionJson"] = input.DescriptionJSON
		if derived := derivePlainText(input.DescriptionJSON); derived != "" {
			setFields["description"] = derived
		}
	}
	if input.Status != "" {
		if !isValidMaintenanceStatus(input.Status) {
			return models.Maintenance{}, fmt.Errorf("%w: invalid maintenance status", shared.ErrInvalidInput)
		}
		setFields["status"] = input.Status
	}
	if input.VisibilityState == models.IncidentVisibilityDraft {
		setFields["status"] = models.MaintenanceDraft
	}
	if input.VisibilityState == models.IncidentVisibilityPublished {
		if status, ok := setFields["status"].(models.MaintenanceStatus); !ok || status == models.MaintenanceDraft {
			setFields["status"] = models.MaintenanceScheduled
		}
	}
	if input.StartTime != "" {
		t, err := time.Parse(time.RFC3339, input.StartTime)
		if err != nil {
			return models.Maintenance{}, fmt.Errorf("%w: invalid startTime format, use RFC3339", shared.ErrInvalidInput)
		}
		setFields["startTime"] = t
	}
	if input.EndTime != "" {
		t, err := time.Parse(time.RFC3339, input.EndTime)
		if err != nil {
			return models.Maintenance{}, fmt.Errorf("%w: invalid endTime format, use RFC3339", shared.ErrInvalidInput)
		}
		setFields["endTime"] = t
	}

	updated, err := s.repo.UpdateByID(ctx, id, setFields)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Maintenance{}, fmt.Errorf("%w: maintenance not found", shared.ErrNotFound)
		}
		return models.Maintenance{}, err
	}

	if err := s.repo.InsertAuditLog(ctx, models.AuditLog{
		ID:           primitive.NewObjectID(),
		ResourceType: models.AuditResourceMaintenance,
		ResourceID:   id,
		Action:       auditActionForMaintenance(updated.Status),
		At:           time.Now(),
		Summary:      "Maintenance updated",
	}); err != nil {
		return models.Maintenance{}, err
	}

	return updated, nil
}

func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (models.Maintenance, error) {
	maintenance, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Maintenance{}, fmt.Errorf("%w: maintenance not found", shared.ErrNotFound)
		}
		return models.Maintenance{}, err
	}

	return maintenance, nil
}

func (s *Service) Delete(ctx context.Context, id primitive.ObjectID) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("%w: maintenance not found", shared.ErrNotFound)
		}
		return err
	}

	return s.repo.InsertAuditLog(ctx, models.AuditLog{
		ID:           primitive.NewObjectID(),
		ResourceType: models.AuditResourceMaintenance,
		ResourceID:   id,
		Action:       models.AuditActionDeleted,
		At:           time.Now(),
		Summary:      "Maintenance deleted",
	})
}

func (s *Service) ListHistory(ctx context.Context, id primitive.ObjectID) ([]models.AuditLog, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: maintenance not found", shared.ErrNotFound)
		}
		return nil, err
	}

	return s.repo.ListHistory(ctx, id)
}

func auditActionForMaintenance(status models.MaintenanceStatus) models.AuditAction {
	if status == models.MaintenanceDraft {
		return models.AuditActionDraftSaved
	}
	if status == models.MaintenanceScheduled || status == models.MaintenanceActive || status == models.MaintenanceInProgress {
		return models.AuditActionPublished
	}

	return models.AuditActionEdited
}

func isValidMaintenanceStatus(status models.MaintenanceStatus) bool {
	switch status {
	case models.MaintenanceDraft, models.MaintenanceScheduled, models.MaintenanceInProgress, models.MaintenanceActive, models.MaintenanceCompleted:
		return true
	default:
		return false
	}
}

func derivePlainText(document models.RichTextDocument) string {
	content, ok := document["content"].([]any)
	if !ok {
		return ""
	}

	parts := make([]string, 0)
	for _, node := range content {
		parts = append(parts, flattenRichTextNode(node)...)
	}

	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func flattenRichTextNode(node any) []string {
	asMap, ok := node.(map[string]any)
	if !ok {
		return nil
	}

	parts := make([]string, 0)
	if text, ok := asMap["text"].(string); ok {
		parts = append(parts, text)
	}

	children, ok := asMap["content"].([]any)
	if !ok {
		return parts
	}

	for _, child := range children {
		parts = append(parts, flattenRichTextNode(child)...)
	}

	return parts
}
