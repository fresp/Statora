package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"status-platform/configs"
	"status-platform/internal/database"
	"status-platform/internal/handlers"
	"status-platform/internal/middleware"
	"status-platform/internal/models"
)

func main() {
	// Load .env if present
	godotenv.Load()

	cfg := configs.Load()

	// Connect MongoDB
	if err := database.ConnectMongo(cfg.MongoURI, cfg.MongoDBName); err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	// Connect Redis (non-fatal if unavailable)
	if err := database.ConnectRedis(cfg.RedisAddr); err != nil {
		log.Printf("Redis connection warning: %v", err)
	}

	db := database.GetDB()

	// Seed admin user
	seedAdmin(db, cfg)

	// WebSocket hub
	hub := handlers.NewHub()
	go hub.Run()

	// Gin setup
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// WebSocket
	r.GET("/ws", handlers.ServeWs(hub))

	api := r.Group("/api")

	// Public routes
	api.GET("/status/summary", handlers.GetStatusSummary(db))
	api.GET("/status/components", handlers.GetStatusComponents(db))
	api.GET("/status/incidents", handlers.GetStatusIncidents(db))
	api.POST("/subscribe", handlers.Subscribe(db))

	// Auth
	api.POST("/auth/login", handlers.Login(db, cfg.JWTSecret))

	// Protected routes
	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	auth.GET("/auth/me", handlers.GetMe(db))

	// Components
	auth.GET("/components", handlers.GetComponents(db))
	auth.POST("/components", handlers.CreateComponent(db, hub))
	auth.PATCH("/components/:id", handlers.UpdateComponent(db, hub))
	auth.DELETE("/components/:id", handlers.DeleteComponent(db))

	// Subcomponents
	auth.GET("/components/:id/subcomponents", handlers.GetSubComponents(db))
	auth.GET("/subcomponents", handlers.GetSubComponents(db))
	auth.POST("/subcomponents", handlers.CreateSubComponent(db))
	auth.PATCH("/subcomponents/:id", handlers.UpdateSubComponent(db))

	// Monitors
	auth.GET("/monitors", handlers.GetMonitors(db))
	auth.POST("/monitors", handlers.CreateMonitor(db))
	auth.DELETE("/monitors/:id", handlers.DeleteMonitor(db))
	auth.GET("/monitors/:id/logs", handlers.GetMonitorLogs(db))
	auth.GET("/monitors/:id/uptime", handlers.GetMonitorUptime(db))
	auth.GET("/monitors/:id/history", handlers.GetMonitorHistory(db))
	auth.GET("/monitors/outages", handlers.GetMonitorOutages(db))

	// Incidents
	auth.GET("/incidents", handlers.GetIncidents(db))
	auth.POST("/incidents", handlers.CreateIncident(db, hub))
	auth.PATCH("/incidents/:id", handlers.UpdateIncident(db, hub))
	auth.POST("/incidents/:id/update", handlers.AddIncidentUpdate(db, hub))
	auth.GET("/incidents/:id/updates", handlers.GetIncidentUpdates(db))

	// Maintenance
	auth.GET("/maintenance", handlers.GetMaintenance(db))
	auth.POST("/maintenance", handlers.CreateMaintenance(db))
	auth.PATCH("/maintenance/:id", handlers.UpdateMaintenance(db))

	// Subscribers
	auth.GET("/subscribers", handlers.GetSubscribers(db))
	auth.DELETE("/subscribers/:id", handlers.DeleteSubscriber(db))

	log.Printf("API server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func seedAdmin(db *mongo.Database, cfg *configs.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing models.Admin
	if err := db.Collection("admins").FindOne(ctx, bson.M{"email": cfg.AdminEmail}).Decode(&existing); err == nil {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return
	}

	admin := models.Admin{
		ID:           primitive.NewObjectID(),
		Username:     cfg.AdminUser,
		Email:        cfg.AdminEmail,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if _, err := db.Collection("admins").InsertOne(ctx, admin); err != nil {
		log.Printf("Failed to seed admin: %v", err)
		return
	}
	log.Printf("Admin seeded: %s / %s", cfg.AdminEmail, cfg.AdminUser)
}
