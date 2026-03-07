package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ComponentStatus string

const (
	StatusOperational      ComponentStatus = "operational"
	StatusDegradedPerf     ComponentStatus = "degraded_performance"
	StatusPartialOutage    ComponentStatus = "partial_outage"
	StatusMajorOutage      ComponentStatus = "major_outage"
	StatusMaintenance      ComponentStatus = "maintenance"
)

type Component struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      ComponentStatus    `bson:"status" json:"status"`
	Order     	int       		   `bson:"order"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
