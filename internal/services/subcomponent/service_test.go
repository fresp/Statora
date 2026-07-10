package subcomponent

import (
	"context"
	"testing"

	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type stubSubComponentRepo struct {
	listResult        []models.SubComponent
	listTotal         int64
	listErr           error
	listFilter        bson.M
	listPage          int
	listLimit         int
	insertErr         error
	inserted          []models.SubComponent
	updateResult      models.SubComponent
	updateErr         error
	findByIDResult    models.SubComponent
	findByIDErr       error
	deleteResult      int64
	deleteErr         error
	componentExists   bool
	componentExistsErr error
	componentExistsID primitive.ObjectID
	cleanupErr        error
}

func (r *stubSubComponentRepo) List(_ context.Context, filter bson.M, page, limit int) ([]models.SubComponent, int64, error) {
	r.listFilter = filter
	r.listPage = page
	r.listLimit = limit
	return r.listResult, r.listTotal, r.listErr
}

func (r *stubSubComponentRepo) Insert(_ context.Context, sub models.SubComponent) error {
	r.inserted = append(r.inserted, sub)
	return r.insertErr
}

func (r *stubSubComponentRepo) UpdateByID(_ context.Context, _ primitive.ObjectID, _ bson.M) (models.SubComponent, error) {
	return r.updateResult, r.updateErr
}

func (r *stubSubComponentRepo) FindByID(_ context.Context, _ primitive.ObjectID) (models.SubComponent, error) {
	return r.findByIDResult, r.findByIDErr
}

func (r *stubSubComponentRepo) DeleteByID(_ context.Context, _ primitive.ObjectID) (int64, error) {
	return r.deleteResult, r.deleteErr
}

func (r *stubSubComponentRepo) CountByComponentID(_ context.Context, _ primitive.ObjectID) (int64, error) {
	return 0, nil
}

func (r *stubSubComponentRepo) ComponentExists(_ context.Context, id primitive.ObjectID) (bool, error) {
	r.componentExistsID = id
	return r.componentExists, r.componentExistsErr
}

func (r *stubSubComponentRepo) CleanupReferencesForDeletedSubComponent(_ context.Context, _ primitive.ObjectID, _ primitive.ObjectID) error {
	return r.cleanupErr
}

func TestCreateRejectsBlankName(t *testing.T) {
	repo := &stubSubComponentRepo{componentExists: true}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{ComponentID: primitive.NewObjectID().Hex(), Name: "   "})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Empty(t, repo.inserted)
}

func TestCreateRejectsUnknownComponent(t *testing.T) {
	repo := &stubSubComponentRepo{componentExists: false}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{ComponentID: primitive.NewObjectID().Hex(), Name: "Worker"})
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
	assert.Empty(t, repo.inserted)
}

func TestCreateDefaultsStatusAndPersistsSubcomponent(t *testing.T) {
	componentID := primitive.NewObjectID()
	repo := &stubSubComponentRepo{componentExists: true}
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateInput{ComponentID: componentID.Hex(), Name: "Worker", Description: "Queue worker"})
	require.NoError(t, err)
	require.Len(t, repo.inserted, 1)
	assert.Equal(t, componentID, repo.componentExistsID)
	assert.Equal(t, "Worker", created.Name)
	assert.Equal(t, models.StatusOperational, created.Status)
	assert.Equal(t, created, repo.inserted[0])
	assert.False(t, created.ID.IsZero())
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())
}

func TestGetByComponentIDDelegatesToRepositoryList(t *testing.T) {
	componentID := primitive.NewObjectID()
	repo := &stubSubComponentRepo{listResult: []models.SubComponent{{Name: "Worker"}}, listTotal: 1}
	svc := NewService(repo)

	subs, total, err := svc.GetByComponentID(context.Background(), componentID, 2, 25)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, bson.M{"componentId": componentID}, repo.listFilter)
	assert.Equal(t, 2, repo.listPage)
	assert.Equal(t, 25, repo.listLimit)
}
