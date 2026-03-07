package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MaintenanceStatus string

const (
	MaintenanceScheduled  MaintenanceStatus = "scheduled"
	MaintenanceInProgress MaintenanceStatus = "in_progress"
	MaintenanceCompleted  MaintenanceStatus = "completed"
)

type Maintenance struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title       string               `bson:"title" json:"title"`
	Description string               `bson:"description" json:"description"`
	Components  []primitive.ObjectID `bson:"components" json:"components"`
	StartTime   time.Time            `bson:"startTime" json:"startTime"`
	EndTime     time.Time            `bson:"endTime" json:"endTime"`
	Status      MaintenanceStatus    `bson:"status" json:"status"`
}
