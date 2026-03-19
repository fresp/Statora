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
	"github.com/fresp/StatusForge/internal/models"
)

func GetSubComponents(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{}
		if cid := c.Param("id"); cid != "" {
			oid, err := primitive.ObjectIDFromHex(cid)
			if err == nil {
				filter["componentId"] = oid
			}
		}
		if cid := c.Query("componentId"); cid != "" {
			oid, err := primitive.ObjectIDFromHex(cid)
			if err == nil {
				filter["componentId"] = oid
			}
		}

		cursor, err := db.Collection("subcomponents").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var subs []models.SubComponent
		if err := cursor.All(ctx, &subs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if subs == nil {
			subs = []models.SubComponent{}
		}
		c.JSON(http.StatusOK, subs)
	}
}

func CreateSubComponent(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ComponentID string                 `json:"componentId" binding:"required"`
			Name        string                 `json:"name" binding:"required"`
			Description string                 `json:"description"`
			Status      models.ComponentStatus `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		compID, err := primitive.ObjectIDFromHex(req.ComponentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid componentId"})
			return
		}

		if req.Status == "" {
			req.Status = models.StatusOperational
		}

		sub := models.SubComponent{
			ID:          primitive.NewObjectID(),
			ComponentID: compID,
			Name:        req.Name,
			Description: req.Description,
			Status:      req.Status,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := db.Collection("subcomponents").InsertOne(ctx, sub); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, sub)
	}
}

func UpdateSubComponent(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			Status      models.ComponentStatus `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		setFields := bson.M{}
		if req.Name != "" {
			setFields["name"] = req.Name
		}
		if req.Description != "" {
			setFields["description"] = req.Description
		}
		if req.Status != "" {
			setFields["status"] = req.Status
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var sub models.SubComponent
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		err = db.Collection("subcomponents").FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": setFields}, opts).Decode(&sub)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "subcomponent not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}
