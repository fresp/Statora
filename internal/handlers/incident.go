package handlers

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/fresp/StatusForge/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var errInvalidAffectedComponentPayload = errors.New("invalid affected component payload")

type incidentAffectedComponentInput struct {
	ComponentID     string   `json:"componentId"`
	SubComponentIDs []string `json:"subComponentIds"`
}

type incidentRequestBody struct {
	Title                    string                           `json:"title" binding:"required"`
	Description              string                           `json:"description"`
	Status                   models.IncidentStatus            `json:"status"`
	Impact                   models.IncidentImpact            `json:"impact"`
	AffectedComponents       []string                         `json:"affectedComponents"`
	Components               []string                         `json:"components"`
	AffectedComponentTargets []incidentAffectedComponentInput `json:"affectedComponentTargets"`
	AffectedComponentsNew    []incidentAffectedComponentInput `json:"affected_components"`
}

func uniqueObjectIDs(ids []primitive.ObjectID) []primitive.ObjectID {
	seen := make(map[primitive.ObjectID]struct{}, len(ids))
	result := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func normalizeIncidentTargets(
	targets []incidentAffectedComponentInput,
	legacy []string,
	legacyAlias []string,
	allowEmpty bool,
) ([]primitive.ObjectID, []models.IncidentAffectedComponent, error) {
	bucket := make(map[primitive.ObjectID]map[primitive.ObjectID]struct{})
	componentOrder := make([]primitive.ObjectID, 0)

	mergeTarget := func(componentID primitive.ObjectID, subIDs []primitive.ObjectID) {
		if _, ok := bucket[componentID]; !ok {
			bucket[componentID] = map[primitive.ObjectID]struct{}{}
			componentOrder = append(componentOrder, componentID)
		}
		for _, subID := range subIDs {
			bucket[componentID][subID] = struct{}{}
		}
	}

	for _, t := range targets {
		if t.ComponentID == "" {
			return nil, nil, errInvalidAffectedComponentPayload
		}
		componentID, err := primitive.ObjectIDFromHex(t.ComponentID)
		if err != nil {
			return nil, nil, err
		}

		subIDs := make([]primitive.ObjectID, 0, len(t.SubComponentIDs))
		for _, sid := range t.SubComponentIDs {
			subID, subErr := primitive.ObjectIDFromHex(sid)
			if subErr != nil {
				return nil, nil, subErr
			}
			subIDs = append(subIDs, subID)
		}
		mergeTarget(componentID, subIDs)
	}

	for _, rawID := range append(legacy, legacyAlias...) {
		componentID, err := primitive.ObjectIDFromHex(rawID)
		if err != nil {
			return nil, nil, err
		}
		mergeTarget(componentID, nil)
	}

	componentIDs := make([]primitive.ObjectID, 0, len(componentOrder))
	normalizedTargets := make([]models.IncidentAffectedComponent, 0, len(componentOrder))

	for _, componentID := range componentOrder {
		subMap := bucket[componentID]
		subIDs := make([]primitive.ObjectID, 0, len(subMap))
		for sid := range subMap {
			subIDs = append(subIDs, sid)
		}
		sort.Slice(subIDs, func(i, j int) bool { return subIDs[i].Hex() < subIDs[j].Hex() })

		componentIDs = append(componentIDs, componentID)
		normalizedTargets = append(normalizedTargets, models.IncidentAffectedComponent{
			ComponentID:     componentID,
			SubComponentIDs: subIDs,
		})
	}

	componentIDs = uniqueObjectIDs(componentIDs)

	if !allowEmpty && len(componentIDs) == 0 {
		return nil, nil, mongo.ErrNoDocuments
	}

	return componentIDs, normalizedTargets, nil
}

func validateIncidentTargets(ctx context.Context, db *mongo.Database, targets []models.IncidentAffectedComponent) error {
	if len(targets) == 0 {
		return nil
	}

	componentIDs := make([]primitive.ObjectID, 0, len(targets))
	for _, t := range targets {
		componentIDs = append(componentIDs, t.ComponentID)
	}

	componentCount, err := db.Collection("components").CountDocuments(ctx, bson.M{"_id": bson.M{"$in": componentIDs}})
	if err != nil {
		return err
	}
	if componentCount != int64(len(uniqueObjectIDs(componentIDs))) {
		return mongo.ErrNoDocuments
	}

	for _, t := range targets {
		if len(t.SubComponentIDs) == 0 {
			continue
		}
		subCount, countErr := db.Collection("subcomponents").CountDocuments(ctx, bson.M{
			"_id":         bson.M{"$in": t.SubComponentIDs},
			"componentId": t.ComponentID,
		})
		if countErr != nil {
			return countErr
		}
		if subCount != int64(len(uniqueObjectIDs(t.SubComponentIDs))) {
			return mongo.ErrNoDocuments
		}
	}

	return nil
}

func invalidAffectedComponentsError() gin.H {
	return gin.H{"error": "invalid affected components payload"}
}

func invalidAffectedComponentsReferenceError() gin.H {
	return gin.H{"error": "one or more affected components or subcomponents are invalid"}
}

func GetIncidents(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, limit, err := parsePaginationParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{}
		if status := c.Query("status"); status == "active" {
			filter["status"] = bson.M{"$ne": models.IncidentResolved}
		} else if status != "" {
			filter["status"] = status
		}

		startDate, endDate, err := parseDateRangeParams(c.Query("start_date"), c.Query("end_date"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if startDate != nil || endDate != nil {
			createdAtFilter := bson.M{}
			if startDate != nil {
				createdAtFilter["$gte"] = *startDate
			}
			if endDate != nil {
				createdAtFilter["$lt"] = *endDate
			}
			filter["createdAt"] = createdAtFilter
		}

		total, err := db.Collection("incidents").CountDocuments(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		skip := int64((page - 1) * limit)
		cursor, err := db.Collection("incidents").Find(ctx, filter,
			options.Find().
				SetSort(bson.D{{Key: "createdAt", Value: -1}}).
				SetSkip(skip).
				SetLimit(int64(limit)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var incidents []models.Incident
		if err := cursor.All(ctx, &incidents); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if incidents == nil {
			incidents = []models.Incident{}
		}
		writePaginatedResponse(c, incidents, int(total), page, limit)
	}
}

func CreateIncident(db *mongo.Database, hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req incidentRequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Status == "" {
			req.Status = models.IncidentInvestigating
		}
		if req.Impact == "" {
			req.Impact = models.ImpactMinor
		}

		rawUserID, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authenticated user context"})
			return
		}

		userIDHex, ok := rawUserID.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user context"})
			return
		}

		userID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user id"})
			return
		}

		creatorUsername, _ := c.Get("username")
		creatorName, _ := creatorUsername.(string)

		compIDs, targets, err := normalizeIncidentTargets(
			append(req.AffectedComponentTargets, req.AffectedComponentsNew...),
			req.AffectedComponents,
			req.Components,
			true,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, invalidAffectedComponentsError())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := validateIncidentTargets(ctx, db, targets); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, invalidAffectedComponentsReferenceError())
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		incident := models.Incident{
			ID:                       primitive.NewObjectID(),
			Title:                    req.Title,
			Description:              req.Description,
			Status:                   req.Status,
			Impact:                   req.Impact,
			CreatorID:                &userID,
			CreatorUsername:          creatorName,
			AffectedComponents:       compIDs,
			AffectedComponentTargets: targets,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if _, err := db.Collection("incidents").InsertOne(ctx, incident); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		DispatchWebhookEvent(db, "incident_created", incident)
		BroadcastEvent(hub, "incident_created", incident)
		c.JSON(http.StatusCreated, incident)
	}
}

func UpdateIncident(db *mongo.Database, hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req incidentRequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		setFields := bson.M{"updatedAt": time.Now()}
		if req.Title != "" {
			setFields["title"] = req.Title
		}
		if req.Description != "" {
			setFields["description"] = req.Description
		}
		if req.Status != "" {
			setFields["status"] = req.Status
			if req.Status == models.IncidentResolved {
				now := time.Now()
				setFields["resolvedAt"] = now
			}
		}
		if req.Impact != "" {
			setFields["impact"] = req.Impact
		}
		hasTargetPayload := len(req.AffectedComponentTargets) > 0 || len(req.AffectedComponentsNew) > 0 || len(req.AffectedComponents) > 0 || len(req.Components) > 0

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if hasTargetPayload {
			compIDs, targets, targetErr := normalizeIncidentTargets(
				append(req.AffectedComponentTargets, req.AffectedComponentsNew...),
				req.AffectedComponents,
				req.Components,
				true,
			)
			if targetErr != nil {
				c.JSON(http.StatusBadRequest, invalidAffectedComponentsError())
				return
			}

			if validationErr := validateIncidentTargets(ctx, db, targets); validationErr != nil {
				if validationErr == mongo.ErrNoDocuments {
					c.JSON(http.StatusBadRequest, invalidAffectedComponentsReferenceError())
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": validationErr.Error()})
				return
			}

			setFields["affectedComponents"] = compIDs
			setFields["affectedComponentTargets"] = targets
		}

		var incident models.Incident
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		err = db.Collection("incidents").FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": setFields}, opts).Decode(&incident)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		eventType := "incident_updated"
		if incident.Status == models.IncidentResolved {
			eventType = "incident_resolved"
		}
		DispatchWebhookEvent(db, eventType, incident)
		BroadcastEvent(hub, eventType, incident)
		c.JSON(http.StatusOK, incident)
	}
}

func AddIncidentUpdate(db *mongo.Database, hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		incidentID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident id"})
			return
		}

		var req struct {
			Message string                `json:"message" binding:"required"`
			Status  models.IncidentStatus `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := models.IncidentUpdate{
			ID:         primitive.NewObjectID(),
			IncidentID: incidentID,
			Message:    req.Message,
			Status:     req.Status,
			CreatedAt:  time.Now(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// insert update log
		if _, err := db.Collection("incident_updates").InsertOne(ctx, update); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// build update fields
		updateFields := bson.M{
			"status":    req.Status,
			"updatedAt": time.Now(),
		}

		// if resolved, set resolvedAt
		if req.Status == models.IncidentResolved {
			updateFields["resolvedAt"] = time.Now()
		}

		// update incident document
		_, err = db.Collection("incidents").UpdateOne(
			ctx,
			bson.M{"_id": incidentID},
			bson.M{"$set": updateFields},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		DispatchWebhookEvent(db, "incident_update_added", update)
		BroadcastEvent(hub, "incident_update_added", update)

		c.JSON(http.StatusCreated, update)
	}
}

func GetIncidentUpdates(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		incidentID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := db.Collection("incident_updates").Find(ctx,
			bson.M{"incidentId": incidentID},
			options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var updates []models.IncidentUpdate
		if err := cursor.All(ctx, &updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if updates == nil {
			updates = []models.IncidentUpdate{}
		}
		c.JSON(http.StatusOK, updates)
	}
}
