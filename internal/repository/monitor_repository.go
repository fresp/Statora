package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fresp/StatusForge/internal/models"
)

type MonitorRepository interface {
	Insert(ctx context.Context, monitor models.Monitor) error
	Update(ctx context.Context, id primitive.ObjectID, monitor models.Monitor) (bool, error)
	List(ctx context.Context, page, limit int) ([]models.Monitor, int64, error)
}

type MongoMonitorRepository struct {
	collection *mongo.Collection
}

func NewMongoMonitorRepository(db *mongo.Database) *MongoMonitorRepository {
	return &MongoMonitorRepository{collection: db.Collection("monitors")}
}

func (r *MongoMonitorRepository) Insert(ctx context.Context, monitor models.Monitor) error {
	_, err := r.collection.InsertOne(ctx, monitor)
	return err
}

func (r *MongoMonitorRepository) Update(ctx context.Context, id primitive.ObjectID, monitor models.Monitor) (bool, error) {
	update := bson.M{
		"$set": bson.M{
			"name":            monitor.Name,
			"type":            monitor.Type,
			"target":          monitor.Target,
			"monitoring":      monitor.Monitoring,
			"sslThresholds":   monitor.SSLThresholds,
			"intervalSeconds": monitor.IntervalSeconds,
			"timeoutSeconds":  monitor.TimeoutSeconds,
			"componentId":     monitor.ComponentID,
			"subComponentId":  monitor.SubComponentID,
			"updatedAt":       time.Now(),
		},
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return false, err
	}

	return res.MatchedCount > 0, nil
}

func (r *MongoMonitorRepository) List(ctx context.Context, page, limit int) ([]models.Monitor, int64, error) {
	filter := bson.M{}
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var monitors []models.Monitor
	if err := cursor.All(ctx, &monitors); err != nil {
		return nil, 0, err
	}
	if monitors == nil {
		monitors = []models.Monitor{}
	}

	return monitors, total, nil
}
