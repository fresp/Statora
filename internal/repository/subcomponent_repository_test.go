package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoSubComponentRepositoryImplementsSubComponentRepository(t *testing.T) {
	var _ SubComponentRepository = (*MongoSubComponentRepository)(nil)
}

func TestNewMongoSubComponentRepositoryReturnsSubComponentRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoSubComponentRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "SubComponentRepository", constructorType.Out(0).Name())
}
