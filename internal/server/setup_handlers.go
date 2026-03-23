package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/fresp/StatusForge/configs"
	"github.com/fresp/StatusForge/internal/database"
)

type setupStatusResponse struct {
	SetupDone bool                  `json:"setupDone"`
	Engine    string                `json:"engine"`
	DBStatus  database.EngineStatus `json:"dbStatus"`
}

type setupSaveRequest struct {
	Engine      string `json:"engine"`
	MongoURI    string `json:"mongoUri"`
	MongoDBName string `json:"mongoDbName"`
	SQLitePath  string `json:"sqlitePath"`
}

type mongoValidateRequest struct {
	MongoURI    string `json:"mongoUri"`
	MongoDBName string `json:"mongoDbName"`
}

func RegisterSetupRoutes(r *gin.Engine) {
	api := r.Group("/api/setup")
	api.GET("/status", getSetupStatus())
	api.POST("/validate/mongo", validateMongoSetup())
	api.POST("/save", saveSetupConfig())
}

func getSetupStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := configs.Load()
		c.JSON(http.StatusOK, setupStatusResponse{
			SetupDone: cfg.SetupDone,
			Engine:    cfg.DBEngine,
			DBStatus:  database.BuildStatus(cfg),
		})
	}
}

func validateMongoSetup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req mongoValidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := database.ValidateMongoConnection(req.MongoURI, req.MongoDBName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"valid": true})
	}
}

func saveSetupConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req setupSaveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		engine := strings.ToLower(strings.TrimSpace(req.Engine))
		switch engine {
		case "mongodb":
			if err := database.ValidateMongoConnection(req.MongoURI, req.MongoDBName); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		case "sqlite":
			if err := database.ValidateSQLitePath(req.SQLitePath); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("engine must be mongodb or sqlite").Error()})
			return
		}

		if err := configs.SaveDatabaseConfig(engine, req.MongoURI, req.MongoDBName, req.SQLitePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cfg := configs.Load()
		c.JSON(http.StatusOK, setupStatusResponse{
			SetupDone: cfg.SetupDone,
			Engine:    cfg.DBEngine,
			DBStatus:  database.BuildStatus(cfg),
		})
	}
}
