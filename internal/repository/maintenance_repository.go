package repository

import (
	"context"
	"time"

	"github.com/fresp/Statora/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MaintenanceRepository interface {
	List(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error)
	ListPublic(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (models.Maintenance, error)
	Insert(ctx context.Context, maintenance models.Maintenance) error
	UpdateByID(ctx context.Context, id primitive.ObjectID, setFields bson.M) (models.Maintenance, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	InsertAuditLog(ctx context.Context, audit models.AuditLog) error
	ListHistory(ctx context.Context, maintenanceID primitive.ObjectID) ([]models.AuditLog, error)
}

type MongoMaintenanceRepository struct {
	collection *mongo.Collection
	auditLogs  *mongo.Collection
}

var _ MaintenanceRepository = (*MongoMaintenanceRepository)(nil)

func NewMongoMaintenanceRepository(db *mongo.Database) MaintenanceRepository {
	return &MongoMaintenanceRepository{
		collection: db.Collection("maintenance"),
		auditLogs:  db.Collection("audit_logs"),
	}
}

func (r *MongoMaintenanceRepository) List(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error) {
	return r.listByFilter(ctx, bson.M{}, page, limit)
}

func (r *MongoMaintenanceRepository) ListPublic(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error) {
	return r.listByFilter(ctx, bson.M{
		"status": bson.M{"$in": []models.MaintenanceStatus{
			models.MaintenanceScheduled,
			models.MaintenanceInProgress,
			models.MaintenanceActive,
		}},
	}, page, limit)
}

func (r *MongoMaintenanceRepository) FindByID(ctx context.Context, id primitive.ObjectID) (models.Maintenance, error) {
	var maintenance models.Maintenance
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&maintenance)
	if err != nil {
		return models.Maintenance{}, err
	}

	return maintenance, nil
}

func (r *MongoMaintenanceRepository) listByFilter(ctx context.Context, filter bson.M, page, limit int) ([]models.Maintenance, int64, error) {
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "startTime", Value: -1}}).SetSkip(skip).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []models.Maintenance
	if err := cursor.All(ctx, &items); err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []models.Maintenance{}
	}

	return items, total, nil
}

func (r *MongoMaintenanceRepository) Insert(ctx context.Context, maintenance models.Maintenance) error {
	_, err := r.collection.InsertOne(ctx, maintenance)
	return err
}

func (r *MongoMaintenanceRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, setFields bson.M) (models.Maintenance, error) {
	setFields["updatedAt"] = time.Now()
	var result models.Maintenance
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": setFields}, opts).Decode(&result)
	if err != nil {
		return models.Maintenance{}, err
	}

	return result, nil
}

func (r *MongoMaintenanceRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *MongoMaintenanceRepository) InsertAuditLog(ctx context.Context, audit models.AuditLog) error {
	_, err := r.auditLogs.InsertOne(ctx, audit)
	return err
}

func (r *MongoMaintenanceRepository) ListHistory(ctx context.Context, maintenanceID primitive.ObjectID) ([]models.AuditLog, error) {
	cursor, err := r.auditLogs.Find(ctx,
		bson.M{"resourceId": maintenanceID},
		options.Find().SetSort(bson.D{{Key: "at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []models.AuditLog
	if err := cursor.All(ctx, &history); err != nil {
		return nil, err
	}
	if history == nil {
		history = []models.AuditLog{}
	}

	return history, nil
}
