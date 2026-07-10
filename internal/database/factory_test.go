package database

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNewRepositoryFactoryPopulatesMongoRepositories(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://127.0.0.1:27017/?serverSelectionTimeoutMS=1"))
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	factory := NewRepositoryFactory(client.Database("statora_test"))

	if factory == nil {
		t.Fatal("NewRepositoryFactory() = nil")
	}

	tests := []struct {
		name string
		got  any
	}{
		{name: "user", got: factory.Users},
		{name: "component", got: factory.Components},
		{name: "incident", got: factory.Incidents},
		{name: "maintenance", got: factory.Maintenances},
		{name: "monitor", got: factory.Monitors},
		{name: "settings", got: factory.Settings},
		{name: "status", got: factory.Status},
		{name: "subcomponent", got: factory.SubComponents},
		{name: "subscriber", got: factory.Subscribers},
		{name: "webhook", got: factory.WebhookChannels},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got == nil {
				t.Fatalf("repository %s was nil", tt.name)
			}
		})
	}
}
