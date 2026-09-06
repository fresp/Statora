package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscriber struct {
	ID                         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email                      string             `bson:"email" json:"email"`
	Verified                   bool               `bson:"verified" json:"verified"`
	VerificationToken          string             `bson:"verificationToken,omitempty" json:"-"`
	VerificationTokenExpiresAt *time.Time         `bson:"verificationTokenExpiresAt,omitempty" json:"-"`
	VerifiedAt                 *time.Time         `bson:"verifiedAt,omitempty" json:"verifiedAt,omitempty"`
	UnsubscribeToken           string             `bson:"unsubscribeToken,omitempty" json:"-"`
	Unsubscribed               bool               `bson:"unsubscribed" json:"unsubscribed"`
	UnsubscribedAt             *time.Time         `bson:"unsubscribedAt,omitempty" json:"unsubscribedAt,omitempty"`
	CreatedAt                  time.Time          `bson:"createdAt" json:"createdAt"`
}
