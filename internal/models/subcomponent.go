package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubComponent struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ComponentID primitive.ObjectID `bson:"componentId" json:"componentId"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      ComponentStatus    `bson:"status" json:"status"`
	Order     	int       		   `bson:"order"`
	CreatedAt 	time.Time 		   `bson:"created_at"`
	UpdatedAt 	time.Time 		   `bson:"updated_at"`
}
