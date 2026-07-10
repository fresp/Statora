package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoStatusRepositoryImplementsStatusRepository(t *testing.T) {
	var _ StatusRepository = (*MongoStatusRepository)(nil)
}

func TestNewMongoStatusRepositoryReturnsStatusRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoStatusRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "StatusRepository", constructorType.Out(0).Name())
}
