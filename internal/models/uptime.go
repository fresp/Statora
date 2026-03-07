package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DailyUptime struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MonitorID        primitive.ObjectID `bson:"monitorId" json:"monitorId"`
	Date             time.Time          `bson:"date" json:"date"`
	TotalChecks      int                `bson:"totalChecks" json:"totalChecks"`
	SuccessfulChecks int                `bson:"successfulChecks" json:"successfulChecks"`
	UptimePercent    float64            `bson:"uptimePercent" json:"uptimePercent"`
}
