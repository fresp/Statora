package maintenance

import (
	"context"
	"testing"
	"time"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type stubMaintenanceRepo struct {
	listResult      []models.Maintenance
	listTotal       int64
	listErr         error
	listPublicResult []models.Maintenance
	listPublicTotal int64
	listPublicErr   error
	findByIDResult  models.Maintenance
	findByIDErr     error
	insertErr       error
	updateResult    models.Maintenance
	updateErr       error
	deleteErr       error
	auditErr        error
	historyResult   []models.AuditLog
	historyErr      error
	inserted        models.Maintenance
	updatedID       primitive.ObjectID
	updatedFields   bson.M
	auditLogs       []models.AuditLog
	deletedID       primitive.ObjectID
	historyID       primitive.ObjectID
}

func (r *stubMaintenanceRepo) List(_ context.Context, _ int, _ int) ([]models.Maintenance, int64, error) {
	return r.listResult, r.listTotal, r.listErr
}

func (r *stubMaintenanceRepo) ListPublic(_ context.Context, _ int, _ int) ([]models.Maintenance, int64, error) {
	return r.listPublicResult, r.listPublicTotal, r.listPublicErr
}

func (r *stubMaintenanceRepo) FindByID(_ context.Context, _ primitive.ObjectID) (models.Maintenance, error) {
	return r.findByIDResult, r.findByIDErr
}

func (r *stubMaintenanceRepo) Insert(_ context.Context, maintenance models.Maintenance) error {
	r.inserted = maintenance
	return r.insertErr
}

func (r *stubMaintenanceRepo) UpdateByID(_ context.Context, id primitive.ObjectID, setFields bson.M) (models.Maintenance, error) {
	r.updatedID = id
	r.updatedFields = setFields
	return r.updateResult, r.updateErr
}

func (r *stubMaintenanceRepo) DeleteByID(_ context.Context, id primitive.ObjectID) error {
	r.deletedID = id
	return r.deleteErr
}

func (r *stubMaintenanceRepo) InsertAuditLog(_ context.Context, audit models.AuditLog) error {
	r.auditLogs = append(r.auditLogs, audit)
	return r.auditErr
}

func (r *stubMaintenanceRepo) ListHistory(_ context.Context, maintenanceID primitive.ObjectID) ([]models.AuditLog, error) {
	r.historyID = maintenanceID
	return r.historyResult, r.historyErr
}

func TestCreateRejectsBlankTitle(t *testing.T) {
	repo := &stubMaintenanceRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{
		Title:        "   ",
		StartTime:    time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339),
		EndTime:      time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
		CreatorIDHex: primitive.NewObjectID().Hex(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.True(t, repo.inserted.ID.IsZero())
	assert.Empty(t, repo.auditLogs)
}

func TestCreateRejectsEndTimeBeforeStartTime(t *testing.T) {
	repo := &stubMaintenanceRepo{}
	svc := NewService(repo)

	start := time.Now().Add(2 * time.Hour).UTC()
	end := start.Add(-1 * time.Hour)

	_, err := svc.Create(context.Background(), CreateInput{
		Title:        "Database maintenance",
		StartTime:    start.Format(time.RFC3339),
		EndTime:      end.Format(time.RFC3339),
		CreatorIDHex: primitive.NewObjectID().Hex(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.True(t, repo.inserted.ID.IsZero())
	assert.Empty(t, repo.auditLogs)
}

func TestCreateRejectsInvalidComponentObjectID(t *testing.T) {
	repo := &stubMaintenanceRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{
		Title:        "Database maintenance",
		Components:   []string{"not-an-object-id"},
		StartTime:    time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339),
		EndTime:      time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
		CreatorIDHex: primitive.NewObjectID().Hex(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.True(t, repo.inserted.ID.IsZero())
	assert.Empty(t, repo.auditLogs)
}

func TestUpdateRejectsInvalidStatus(t *testing.T) {
	repo := &stubMaintenanceRepo{}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), primitive.NewObjectID(), UpdateInput{Status: models.MaintenanceStatus("banana")})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Nil(t, repo.updatedFields)
	assert.Empty(t, repo.auditLogs)
}
