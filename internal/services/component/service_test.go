package component

import (
	"context"
	"errors"
	"testing"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubComponentRepo struct {
	listComponents   []models.Component
	listTotal        int64
	listErr          error
	inserted         []models.Component
	insertErr        error
	updateResult     models.Component
	updateErr        error
	updateID         primitive.ObjectID
	updateSetFields  bson.M
	findByIDResult   models.Component
	findByIDErr      error
	findByIDInput    primitive.ObjectID
	deleteByIDErr    error
	deleteByIDInput  primitive.ObjectID
	deleteByIDCalled bool
}

func (r *stubComponentRepo) List(_ context.Context, page, limit int) ([]models.Component, int64, error) {
	return r.listComponents, r.listTotal, r.listErr
}

func (r *stubComponentRepo) Insert(_ context.Context, component models.Component) error {
	r.inserted = append(r.inserted, component)
	return r.insertErr
}

func (r *stubComponentRepo) UpdateByID(_ context.Context, id primitive.ObjectID, setFields bson.M) (models.Component, error) {
	r.updateID = id
	r.updateSetFields = setFields
	return r.updateResult, r.updateErr
}

func (r *stubComponentRepo) FindByID(_ context.Context, id primitive.ObjectID) (models.Component, error) {
	r.findByIDInput = id
	return r.findByIDResult, r.findByIDErr
}

func (r *stubComponentRepo) DeleteByID(_ context.Context, id primitive.ObjectID) error {
	r.deleteByIDInput = id
	r.deleteByIDCalled = true
	return r.deleteByIDErr
}

func TestCreateDefaultsStatusAndPersistsComponent(t *testing.T) {
	repo := &stubComponentRepo{}
	svc := NewService(repo, nil)

	created, err := svc.Create(context.Background(), CreateInput{Name: "API", Description: "Primary API"})
	require.NoError(t, err)
	require.Len(t, repo.inserted, 1)
	assert.Equal(t, "API", created.Name)
	assert.Equal(t, models.StatusOperational, created.Status)
	assert.Equal(t, created, repo.inserted[0])
	assert.False(t, created.ID.IsZero())
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())
}

func TestCreateRejectsBlankName(t *testing.T) {
	repo := &stubComponentRepo{}
	svc := NewService(repo, nil)

	_, err := svc.Create(context.Background(), CreateInput{Name: "   "})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Empty(t, repo.inserted)
}

func TestUpdateBuildsSetFieldsAndMapsNotFound(t *testing.T) {
	id := primitive.NewObjectID()
	repo := &stubComponentRepo{updateErr: mongo.ErrNoDocuments}
	svc := NewService(repo, nil)

	_, err := svc.Update(context.Background(), id, UpdateInput{Name: "Gateway", Description: "Updated", Status: models.StatusMajorOutage})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrNotFound)
	assert.Equal(t, id, repo.updateID)
	assert.Equal(t, "Gateway", repo.updateSetFields["name"])
	assert.Equal(t, "Updated", repo.updateSetFields["description"])
	assert.Equal(t, models.StatusMajorOutage, repo.updateSetFields["status"])
	_, ok := repo.updateSetFields["updatedAt"]
	assert.True(t, ok)
}

func TestGetByIDMapsNotFound(t *testing.T) {
	id := primitive.NewObjectID()
	repo := &stubComponentRepo{findByIDErr: mongo.ErrNoDocuments}
	svc := NewService(repo, nil)

	_, err := svc.GetByID(context.Background(), id)
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrNotFound)
	assert.Equal(t, id, repo.findByIDInput)
}

func TestDeleteLoadsComponentBeforeDelete(t *testing.T) {
	id := primitive.NewObjectID()
	repo := &stubComponentRepo{findByIDResult: models.Component{ID: id}}
	svc := NewService(repo, nil)

	err := svc.Delete(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, repo.findByIDInput)
	assert.Equal(t, id, repo.deleteByIDInput)
	assert.True(t, repo.deleteByIDCalled)
}

func TestListDelegatesToRepository(t *testing.T) {
	repo := &stubComponentRepo{listComponents: []models.Component{{Name: "API"}}, listTotal: 1}
	svc := NewService(repo, nil)

	components, total, err := svc.List(context.Background(), 2, 25)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, components, 1)
	assert.Equal(t, "API", components[0].Name)
}

func TestCreatePropagatesRepositoryErrors(t *testing.T) {
	repo := &stubComponentRepo{insertErr: errors.New("boom")}
	svc := NewService(repo, nil)

	_, err := svc.Create(context.Background(), CreateInput{Name: "API"})
	require.EqualError(t, err, "boom")
}

func TestCreateRejectsOverlongName(t *testing.T) {
	repo := &stubComponentRepo{}
	svc := NewService(repo, nil)
	overlongName := string(make([]byte, 121))

	_, err := svc.Create(context.Background(), CreateInput{Name: overlongName})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Empty(t, repo.inserted)
}
