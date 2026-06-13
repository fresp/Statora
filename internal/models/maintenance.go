package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MaintenanceStatus string

const (
	MaintenanceDraft      MaintenanceStatus = "draft"
	MaintenanceScheduled  MaintenanceStatus = "scheduled"
	MaintenanceInProgress MaintenanceStatus = "in_progress"
	MaintenanceActive     MaintenanceStatus = "active"
	MaintenanceCompleted  MaintenanceStatus = "completed"
)

type Maintenance struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title           string               `bson:"title" json:"title"`
	Description     string               `bson:"description" json:"description"`
	DescriptionJSON RichTextDocument     `bson:"descriptionJson,omitempty" json:"descriptionJson,omitempty"`
	CreatorID       *primitive.ObjectID  `bson:"creatorId,omitempty" json:"creatorId,omitempty"`
	CreatorUsername string               `bson:"creatorUsername,omitempty" json:"creatorUsername,omitempty"`
	Components      []primitive.ObjectID `bson:"components" json:"components"`
	StartTime       time.Time            `bson:"startTime" json:"startTime"`
	EndTime         time.Time            `bson:"endTime" json:"endTime"`
	Status          MaintenanceStatus    `bson:"status" json:"status"`
	UpdatedAt       time.Time            `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}
