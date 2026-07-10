package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoSubscriberRepositoryImplementsSubscriberRepository(t *testing.T) {
	var _ SubscriberRepository = (*MongoSubscriberRepository)(nil)
}

func TestNewMongoSubscriberRepositoryReturnsSubscriberRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoSubscriberRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "SubscriberRepository", constructorType.Out(0).Name())
}
