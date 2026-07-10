package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscriber struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	EmailHash string             `bson:"emailHash" json:"-"`
	Verified  bool               `bson:"verified" json:"verified"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
