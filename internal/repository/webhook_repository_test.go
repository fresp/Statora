package repository

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoWebhookChannelRepositoryImplementsWebhookChannelRepository(t *testing.T) {
	var _ WebhookChannelRepository = (*MongoWebhookChannelRepository)(nil)
}

func TestNewMongoWebhookChannelRepositoryReturnsWebhookChannelRepositoryInterface(t *testing.T) {
	constructorType := reflect.TypeOf(NewMongoWebhookChannelRepository)
	require.Equal(t, reflect.Interface, constructorType.Out(0).Kind())
	require.Equal(t, "WebhookChannelRepository", constructorType.Out(0).Name())
}
