// Package server provides the unified server functionality.
package server

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/fresp/StatusForge/configs"
	"github.com/fresp/StatusForge/internal/handlers"
	"github.com/fresp/StatusForge/internal/middleware"
	"github.com/fresp/StatusForge/internal/models"
)

// RegisterAPIRoutes registers all API routes on the given Gin engine
func RegisterAPIRoutes(r *gin.Engine, hub *handlers.Hub, cfg *configs.Config, db *mongo.Database) {
	// Apply CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	r.GET("/ws", handlers.ServeWs(hub))

	api.GET("/status/summary", handlers.GetStatusSummary(db))
	api.GET("/status/components", handlers.GetStatusComponents(db))
	api.GET("/status/incidents", handlers.GetStatusIncidents(db))
	api.GET("/status/settings", handlers.GetPublicStatusPageSettings(db))
	api.POST("/subscribe", handlers.Subscribe(db))

	api.POST("/auth/login", handlers.Login(db, cfg.JWTSecret))
	api.POST("/users/invitations/activate", handlers.ActivateUserInvitation(db))

	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	partialAuth := auth.Group("")

	partialAuth.GET("/auth/me", handlers.GetMe(db))
	partialAuth.PATCH("/auth/me", handlers.ProfileUpdate(db, cfg))
	partialAuth.POST("/auth/mfa/setup", handlers.MFASetup(db, cfg))
	partialAuth.POST("/auth/mfa/verify", handlers.MFAVerify(db, cfg))
	partialAuth.POST("/auth/mfa/recovery/verify", handlers.MFARecoveryVerify(db, cfg))
	partialAuth.POST("/auth/mfa/disable", handlers.MFADisable(db, cfg))

	verifiedAuth := auth.Group("")
	verifiedAuth.Use(middleware.RequireMFAVerified())

	adminOnly := verifiedAuth.Group("")
	adminOnly.Use(middleware.RequireRoles("admin"))

	incidentAndMaintenance := verifiedAuth.Group("")
	incidentAndMaintenance.Use(middleware.RequireRoles("admin", "operator"))

	incidentAndMaintenance.GET("/incidents", handlers.GetIncidents(db))
	incidentAndMaintenance.POST("/incidents", handlers.CreateIncident(db, hub))
	incidentAndMaintenance.PATCH("/incidents/:id", handlers.UpdateIncident(db, hub))
	incidentAndMaintenance.POST("/incidents/:id/update", handlers.AddIncidentUpdate(db, hub))
	incidentAndMaintenance.GET("/incidents/:id/updates", handlers.GetIncidentUpdates(db))

	incidentAndMaintenance.GET("/maintenance", handlers.GetMaintenance(db))
	incidentAndMaintenance.POST("/maintenance", handlers.CreateMaintenance(db))
	incidentAndMaintenance.PATCH("/maintenance/:id", handlers.UpdateMaintenance(db))

	incidentAndMaintenance.GET("/components", handlers.GetComponents(db))
	incidentAndMaintenance.GET("/components/:id/subcomponents", handlers.GetSubComponents(db))
	incidentAndMaintenance.GET("/subcomponents", handlers.GetSubComponents(db))

	adminOnly.POST("/components", handlers.CreateComponent(db, hub))
	adminOnly.PATCH("/components/:id", handlers.UpdateComponent(db, hub))
	adminOnly.DELETE("/components/:id", handlers.DeleteComponent(db))

	adminOnly.POST("/subcomponents", handlers.CreateSubComponent(db))
	adminOnly.PATCH("/subcomponents/:id", handlers.UpdateSubComponent(db))

	adminOnly.GET("/monitors", handlers.GetMonitors(db))
	adminOnly.POST("/monitors", handlers.CreateMonitor(db))
	adminOnly.POST("/monitors/test", handlers.TestMonitor())
	adminOnly.PUT("/monitors/:id", handlers.UpdateMonitor(db))
	adminOnly.DELETE("/monitors/:id", handlers.DeleteMonitor(db))
	adminOnly.GET("/monitors/:id/logs", handlers.GetMonitorLogs(db))
	adminOnly.GET("/monitors/:id/uptime", handlers.GetMonitorUptime(db))
	adminOnly.GET("/monitors/:id/history", handlers.GetMonitorHistory(db))
	adminOnly.GET("/monitors/outages", handlers.GetMonitorOutages(db))

	adminOnly.GET("/subscribers", handlers.GetSubscribers(db))
	adminOnly.DELETE("/subscribers/:id", handlers.DeleteSubscriber(db))
	adminOnly.GET("/settings/status-page", handlers.GetAdminStatusPageSettings(db))
	adminOnly.PATCH("/settings/status-page", handlers.UpdateStatusPageSettings(db, hub))
	adminOnly.GET("/webhook-channels", handlers.GetWebhookChannels(db))
	adminOnly.POST("/webhook-channels", handlers.CreateWebhookChannel(db))
	adminOnly.DELETE("/webhook-channels/:id", handlers.DeleteWebhookChannel(db))

	adminOnly.GET("/users", handlers.GetUsers(db))
	adminOnly.PATCH("/users/:id", handlers.PatchUser(db))
	adminOnly.DELETE("/users/:id", handlers.DeleteUser(db))
	adminOnly.POST("/users/invitations", handlers.CreateUserInvitation(db))
	adminOnly.GET("/users/invitations", handlers.GetUserInvitations(db))
	adminOnly.POST("/users/invitations/:id/refresh", handlers.RefreshUserInvitation(db))
	adminOnly.DELETE("/users/invitations/:id", handlers.RevokeUserInvitation(db))
}

func SeedAdmin(db *mongo.Database, cfg *configs.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing models.User
	if err := db.Collection("users").FindOne(ctx, bson.M{"email": cfg.AdminEmail}).Decode(&existing); err == nil {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[HTTP] Failed to hash admin password: %v", err)
		return
	}

	user := models.User{
		ID:           primitive.NewObjectID(),
		Username:     cfg.AdminUser,
		Email:        cfg.AdminEmail,
		Role:         "admin",
		Status:       "active",
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if _, err := db.Collection("users").InsertOne(ctx, user); err != nil {
		log.Printf("[HTTP] Failed to seed admin: %v", err)
		return
	}

	log.Printf("[HTTP] Admin seeded: %s / %s", cfg.AdminEmail, cfg.AdminUser)
}
