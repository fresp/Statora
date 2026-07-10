package repository

import (
	"context"

	"github.com/fresp/Statora/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ComponentRepository interface {
	List(ctx context.Context, page, limit int) ([]models.Component, int64, error)
	Insert(ctx context.Context, component models.Component) error
	UpdateByID(ctx context.Context, id primitive.ObjectID, setFields bson.M) (models.Component, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (models.Component, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type MongoComponentRepository struct {
	collection *mongo.Collection
}

var _ ComponentRepository = (*MongoComponentRepository)(nil)

func NewMongoComponentRepository(db *mongo.Database) ComponentRepository {
	return &MongoComponentRepository{collection: db.Collection("components")}
}

func (r *MongoComponentRepository) List(ctx context.Context, page, limit int) ([]models.Component, int64, error) {
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetSkip(skip).SetLimit(int64(limit)))
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var components []models.Component
	if err := cursor.All(ctx, &components); err != nil {
		return nil, 0, err
	}
	if components == nil {
		components = []models.Component{}
	}

	return components, total, nil
}

func (r *MongoComponentRepository) Insert(ctx context.Context, component models.Component) error {
	_, err := r.collection.InsertOne(ctx, component)
	return err
}

func (r *MongoComponentRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, setFields bson.M) (models.Component, error) {
	var component models.Component
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": setFields}, opts).Decode(&component)
	if err != nil {
		return models.Component{}, err
	}

	return component, nil
}

func (r *MongoComponentRepository) FindByID(ctx context.Context, id primitive.ObjectID) (models.Component, error) {
	var component models.Component
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&component); err != nil {
		return models.Component{}, err
	}

	return component, nil
}

func (r *MongoComponentRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
