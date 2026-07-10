package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMongoMonitorRepositoryImplementsMonitorRepository(t *testing.T) {
	var _ MonitorRepository = (*MongoMonitorRepository)(nil)
}

func TestNewMongoMonitorRepositoryReturnsMonitorRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoMonitorRepository)
	assert.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	assert.Equal(t, "MonitorRepository", constructorType.Out(0).Name())
}

func TestBuildMonitorLogsPaginationQueryBuildsExpectedFilterAndOptions(t *testing.T) {
	monitorID := primitive.NewObjectID()

	filter, findOptions := buildMonitorLogsPaginationQuery(monitorID, 3, 25)

	assert.Equal(t, bson.M{"monitorId": monitorID}, filter)
	if assert.NotNil(t, findOptions.Sort) {
		assert.Equal(t, bson.D{{Key: "checkedAt", Value: -1}}, findOptions.Sort)
	}
	if assert.NotNil(t, findOptions.Skip) {
		assert.Equal(t, int64(50), *findOptions.Skip)
	}
	if assert.NotNil(t, findOptions.Limit) {
		assert.Equal(t, int64(25), *findOptions.Limit)
	}
}
