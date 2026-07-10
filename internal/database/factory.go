package database

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/fresp/Statora/internal/repository"
)

type RepositoryFactory struct {
	Users           repository.UserRepository
	Components      repository.ComponentRepository
	Incidents       repository.IncidentRepository
	Maintenances    repository.MaintenanceRepository
	Monitors        repository.MonitorRepository
	Settings        repository.SettingsRepository
	Status          repository.StatusRepository
	SubComponents   repository.SubComponentRepository
	Subscribers     repository.SubscriberRepository
	WebhookChannels repository.WebhookChannelRepository
}

func NewRepositoryFactory(db *mongo.Database) *RepositoryFactory {
	return &RepositoryFactory{
		Users:           repository.NewMongoUserRepository(db),
		Components:      repository.NewMongoComponentRepository(db),
		Incidents:       repository.NewMongoIncidentRepository(db),
		Maintenances:    repository.NewMongoMaintenanceRepository(db),
		Monitors:        repository.NewMongoMonitorRepository(db),
		Settings:        repository.NewMongoSettingsRepository(db),
		Status:          repository.NewMongoStatusRepository(db),
		SubComponents:   repository.NewMongoSubComponentRepository(db),
		Subscribers:     repository.NewMongoSubscriberRepository(db),
		WebhookChannels: repository.NewMongoWebhookChannelRepository(db),
	}
}
