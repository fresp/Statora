package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoMaintenanceRepositoryImplementsMaintenanceRepository(t *testing.T) {
	var _ MaintenanceRepository = (*MongoMaintenanceRepository)(nil)
}

func TestNewMongoMaintenanceRepositoryReturnsMaintenanceRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoMaintenanceRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "MaintenanceRepository", constructorType.Out(0).Name())
}
