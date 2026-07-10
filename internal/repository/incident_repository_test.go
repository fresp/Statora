package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoIncidentRepositoryImplementsIncidentRepository(t *testing.T) {
	var _ IncidentRepository = (*MongoIncidentRepository)(nil)
}

func TestNewMongoIncidentRepositoryReturnsIncidentRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoIncidentRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "IncidentRepository", constructorType.Out(0).Name())
}
