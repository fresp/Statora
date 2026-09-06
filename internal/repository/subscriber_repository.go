package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/fresp/Statora/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubscriberRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.Subscriber, error)
	FindByVerificationToken(ctx context.Context, token string) (*models.Subscriber, error)
	FindByUnsubscribeToken(ctx context.Context, token string) (*models.Subscriber, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Subscriber, error)
	Insert(ctx context.Context, sub models.Subscriber) error
	Update(ctx context.Context, sub models.Subscriber) error
	List(ctx context.Context, page, limit int) ([]models.Subscriber, int64, error)
	ListVerified(ctx context.Context) ([]models.Subscriber, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error)
	PruneUnverified(ctx context.Context, olderThan time.Duration) (int64, error)
}

type MongoSubscriberRepository struct {
	collection *mongo.Collection
}

func NewMongoSubscriberRepository(db *mongo.Database) *MongoSubscriberRepository {
	return &MongoSubscriberRepository{collection: db.Collection("subscribers")}
}

func (r *MongoSubscriberRepository) FindByEmail(ctx context.Context, email string) (*models.Subscriber, error) {
	var existing models.Subscriber
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&existing)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &existing, nil
}

func (r *MongoSubscriberRepository) Insert(ctx context.Context, sub models.Subscriber) error {
	_, err := r.collection.InsertOne(ctx, sub)
	return err
}

func (r *MongoSubscriberRepository) List(ctx context.Context, page, limit int) ([]models.Subscriber, int64, error) {
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	cursor, err := r.collection.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip(skip).SetLimit(int64(limit)))
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var subs []models.Subscriber
	if err := cursor.All(ctx, &subs); err != nil {
		return nil, 0, err
	}
	if subs == nil {
		subs = []models.Subscriber{}
	}

	return subs, total, nil
}

func (r *MongoSubscriberRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

func (r *MongoSubscriberRepository) FindByVerificationToken(ctx context.Context, token string) (*models.Subscriber, error) {
	return r.findOneSubscriber(ctx, bson.M{"verificationToken": token})
}

func (r *MongoSubscriberRepository) FindByUnsubscribeToken(ctx context.Context, token string) (*models.Subscriber, error) {
	return r.findOneSubscriber(ctx, bson.M{"unsubscribeToken": token})
}

func (r *MongoSubscriberRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Subscriber, error) {
	return r.findOneSubscriber(ctx, bson.M{"_id": id})
}

func (r *MongoSubscriberRepository) findOneSubscriber(ctx context.Context, filter bson.M) (*models.Subscriber, error) {
	var sub models.Subscriber
	err := r.collection.FindOne(ctx, filter).Decode(&sub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *MongoSubscriberRepository) Update(ctx context.Context, sub models.Subscriber) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": sub.ID}, sub)
	return err
}

func (r *MongoSubscriberRepository) ListVerified(ctx context.Context) ([]models.Subscriber, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"verified": true, "unsubscribed": bson.M{"$ne": true}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subs []models.Subscriber
	if err := cursor.All(ctx, &subs); err != nil {
		return nil, err
	}
	if subs == nil {
		subs = []models.Subscriber{}
	}
	return subs, nil
}

func (r *MongoSubscriberRepository) PruneUnverified(ctx context.Context, olderThan time.Duration) (int64, error) {
	res, err := r.collection.DeleteMany(ctx, bson.M{
		"verified":  false,
		"createdAt": bson.M{"$lt": time.Now().Add(-olderThan)},
	})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func NewSubscriberWithTokens(email string) models.Subscriber {
	now := time.Now()
	expires := now.Add(48 * time.Hour)
	return models.Subscriber{
		ID:                         primitive.NewObjectID(),
		Email:                      email,
		Verified:                   false,
		VerificationToken:          newSubscriberToken(),
		VerificationTokenExpiresAt: &expires,
		UnsubscribeToken:           newSubscriberToken(),
		CreatedAt:                  now,
	}
}

// NewVerificationToken generates a cryptographically random 64-char hex token.
func NewVerificationToken() string {
	return newSubscriberToken()
}

func newSubscriberToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("generate subscriber token: %v", err)) // crypto/rand failure is unrecoverable
	}
	return hex.EncodeToString(buf)
}
