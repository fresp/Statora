package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fresp/Statora/internal/models"
)

type SettingsRepository interface {
	GetSSOSettings(ctx context.Context) (*models.StatusPageSSOSettings, error)
	UpdateSSOSettings(ctx context.Context, updates bson.M) (*models.StatusPageSSOSettings, error)
}

type MongoSettingsRepository struct {
	db *mongo.Database
}

func NewMongoSettingsRepository(db *mongo.Database) *MongoSettingsRepository {
	return &MongoSettingsRepository{db: db}
}

func (r *MongoSettingsRepository) GetSSOSettings(ctx context.Context) (*models.StatusPageSSOSettings, error) {
	settings, err := r.fetchOrCreateStatusPageSettings(ctx)
	if err != nil {
		return nil, err
	}

	return &settings.SSO, nil
}

func (r *MongoSettingsRepository) UpdateSSOSettings(ctx context.Context, updates bson.M) (*models.StatusPageSSOSettings, error) {
	current, err := r.fetchOrCreateStatusPageSettings(ctx)
	if err != nil {
		return nil, err
	}

	var updated models.StatusPageSettings
	err = r.settingsCollection().FindOneAndUpdate(
		ctx,
		bson.M{"key": models.StatusPageSettingsKey},
		bson.M{
			"$set":         updates,
			"$setOnInsert": bson.M{"createdAt": current.CreatedAt, "key": models.StatusPageSettingsKey},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated.SSO, nil
}

func (r *MongoSettingsRepository) settingsCollection() *mongo.Collection {
	return r.db.Collection("settings")
}

func (r *MongoSettingsRepository) fetchOrCreateStatusPageSettings(ctx context.Context) (models.StatusPageSettings, error) {
	var settings models.StatusPageSettings
	err := r.settingsCollection().FindOne(ctx, bson.M{"key": models.StatusPageSettingsKey}).Decode(&settings)
	if err == nil {
		return settings, nil
	}
	if err != mongo.ErrNoDocuments {
		return models.StatusPageSettings{}, err
	}

	defaultSettings := models.DefaultStatusPageSettings()
	if _, insertErr := r.settingsCollection().InsertOne(ctx, defaultSettings); insertErr != nil {
		return models.StatusPageSettings{}, insertErr
	}

	return defaultSettings, nil
}
