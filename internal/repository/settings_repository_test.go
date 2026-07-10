package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoSettingsRepositoryImplementsSettingsRepository(t *testing.T) {
	var _ SettingsRepository = (*MongoSettingsRepository)(nil)
}

func TestNewMongoSettingsRepositoryReturnsSettingsRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoSettingsRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "SettingsRepository", constructorType.Out(0).Name())
}
