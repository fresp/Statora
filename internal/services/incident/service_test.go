package incident

import (
	"context"
	"testing"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubIncidentRepo struct {
	listResult                   []models.Incident
	listTotal                    int64
	listErr                      error
	findByIDResult               models.Incident
	findByIDErr                  error
	insertIncidentErr            error
	updateIncidentResult         models.Incident
	updateIncidentErr            error
	deleteIncidentErr            error
	insertUpdateErr              error
	applyIncidentStatusErr       error
	listUpdatesResult            []models.IncidentUpdate
	listUpdatesErr               error
	insertAuditLogErr            error
	listHistoryResult            []models.AuditLog
	listHistoryErr               error
	countComponentsResult        int64
	countComponentsErr           error
	countSubComponentsResult     int64
	countSubComponentsErr        error
	insertedIncident             models.Incident
	insertUpdateInput            models.IncidentUpdate
	insertAuditLogs              []models.AuditLog
	countComponentsInput         []primitive.ObjectID
	countSubComponentsComponent  primitive.ObjectID
	countSubComponentsIDs        []primitive.ObjectID
}

func (r *stubIncidentRepo) List(_ context.Context, _ bson.M, _ int, _ int) ([]models.Incident, int64, error) {
	return r.listResult, r.listTotal, r.listErr
}

func (r *stubIncidentRepo) FindByID(_ context.Context, _ primitive.ObjectID) (models.Incident, error) {
	return r.findByIDResult, r.findByIDErr
}

func (r *stubIncidentRepo) InsertIncident(_ context.Context, incident models.Incident) error {
	r.insertedIncident = incident
	return r.insertIncidentErr
}

func (r *stubIncidentRepo) UpdateIncidentByID(_ context.Context, _ primitive.ObjectID, _ bson.M) (models.Incident, error) {
	return r.updateIncidentResult, r.updateIncidentErr
}

func (r *stubIncidentRepo) DeleteIncidentByID(_ context.Context, _ primitive.ObjectID) error {
	return r.deleteIncidentErr
}

func (r *stubIncidentRepo) InsertUpdate(_ context.Context, update models.IncidentUpdate) error {
	r.insertUpdateInput = update
	return r.insertUpdateErr
}

func (r *stubIncidentRepo) ApplyIncidentStatus(_ context.Context, _ primitive.ObjectID, _ models.IncidentStatus) error {
	return r.applyIncidentStatusErr
}

func (r *stubIncidentRepo) ListUpdates(_ context.Context, _ primitive.ObjectID) ([]models.IncidentUpdate, error) {
	return r.listUpdatesResult, r.listUpdatesErr
}

func (r *stubIncidentRepo) InsertAuditLog(_ context.Context, audit models.AuditLog) error {
	r.insertAuditLogs = append(r.insertAuditLogs, audit)
	return r.insertAuditLogErr
}

func (r *stubIncidentRepo) ListHistory(_ context.Context, _ primitive.ObjectID) ([]models.AuditLog, error) {
	return r.listHistoryResult, r.listHistoryErr
}

func (r *stubIncidentRepo) CountComponents(_ context.Context, ids []primitive.ObjectID) (int64, error) {
	r.countComponentsInput = ids
	return r.countComponentsResult, r.countComponentsErr
}

func (r *stubIncidentRepo) CountSubComponentsByComponent(_ context.Context, componentID primitive.ObjectID, ids []primitive.ObjectID) (int64, error) {
	r.countSubComponentsComponent = componentID
	r.countSubComponentsIDs = ids
	return r.countSubComponentsResult, r.countSubComponentsErr
}

func TestCreateRejectsBlankTitle(t *testing.T) {
	repo := &stubIncidentRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{RequestBody: RequestBody{Title: "   "}, CreatorIDHex: primitive.NewObjectID().Hex(), CreatorUsername: "admin"})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Empty(t, repo.insertAuditLogs)
	assert.True(t, repo.insertedIncident.ID.IsZero())
}

func TestCreateRejectsInvalidImpact(t *testing.T) {
	repo := &stubIncidentRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{RequestBody: RequestBody{Title: "API incident", Impact: models.IncidentImpact("banana")}, CreatorIDHex: primitive.NewObjectID().Hex(), CreatorUsername: "admin"})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Empty(t, repo.insertAuditLogs)
	assert.True(t, repo.insertedIncident.ID.IsZero())
}

func TestAddUpdateRejectsBlankMessage(t *testing.T) {
	repo := &stubIncidentRepo{findByIDResult: models.Incident{ID: primitive.NewObjectID()}}
	svc := NewService(repo)

	_, err := svc.AddUpdate(context.Background(), primitive.NewObjectID(), "   ", models.IncidentInvestigating)
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.True(t, repo.insertUpdateInput.ID.IsZero())
	assert.Empty(t, repo.insertAuditLogs)
}
