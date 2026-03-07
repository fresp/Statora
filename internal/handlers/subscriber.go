package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"status-platform/internal/models"
)

func Subscribe(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Check if already subscribed
		var existing models.Subscriber
		err := db.Collection("subscribers").FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already subscribed"})
			return
		}

		sub := models.Subscriber{
			ID:        primitive.NewObjectID(),
			Email:     req.Email,
			Verified:  false,
			CreatedAt: time.Now(),
		}

		if _, err := db.Collection("subscribers").InsertOne(ctx, sub); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "subscribed successfully", "id": sub.ID})
	}
}

func GetSubscribers(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := db.Collection("subscribers").Find(ctx, bson.M{},
			options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var subs []models.Subscriber
		if err := cursor.All(ctx, &subs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if subs == nil {
			subs = []models.Subscriber{}
		}
		c.JSON(http.StatusOK, subs)
	}
}

func DeleteSubscriber(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := db.Collection("subscribers").DeleteOne(ctx, bson.M{"_id": id})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscriber not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "unsubscribed"})
	}
}
