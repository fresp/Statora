package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoComponentRepositoryImplementsComponentRepository(t *testing.T) {
	var _ ComponentRepository = (*MongoComponentRepository)(nil)
}

func TestNewMongoComponentRepositoryReturnsComponentRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoComponentRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "ComponentRepository", constructorType.Out(0).Name())
}
