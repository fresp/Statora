package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditResourceType string

const (
	AuditResourceIncident       AuditResourceType = "incident"
	AuditResourceIncidentUpdate AuditResourceType = "incident_update"
	AuditResourceMaintenance    AuditResourceType = "maintenance"
)

type AuditAction string

const (
	AuditActionCreated       AuditAction = "created"
	AuditActionEdited        AuditAction = "edited"
	AuditActionStatusChanged AuditAction = "status_changed"
	AuditActionUpdateAdded   AuditAction = "update_added"
	AuditActionResolved      AuditAction = "resolved"
	AuditActionDeleted       AuditAction = "deleted"
	AuditActionPublished     AuditAction = "published"
	AuditActionDraftSaved    AuditAction = "draft_saved"
)

type AuditLog struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ResourceType  AuditResourceType   `bson:"resourceType" json:"resourceType"`
	ResourceID    primitive.ObjectID  `bson:"resourceId" json:"resourceId"`
	Action        AuditAction         `bson:"action" json:"action"`
	ActorID       *primitive.ObjectID `bson:"actorId,omitempty" json:"actorId,omitempty"`
	ActorUsername string              `bson:"actorUsername,omitempty" json:"actorUsername,omitempty"`
	At            time.Time           `bson:"at" json:"at"`
	Summary       string              `bson:"summary,omitempty" json:"summary,omitempty"`
	Changes       map[string]any      `bson:"changes,omitempty" json:"changes,omitempty"`
}
