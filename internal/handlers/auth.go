package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"status-platform/internal/middleware"
	"status-platform/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func Login(db *mongo.Database, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var admin models.Admin
		err := db.Collection("admins").FindOne(ctx, bson.M{"email": req.Email}).Decode(&admin)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := middleware.GenerateToken(admin.ID.Hex(), admin.Username, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"admin": gin.H{
				"id":       admin.ID,
				"username": admin.Username,
				"email":    admin.Email,
			},
		})
	}
}

func GetMe(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, _ := c.Get("adminId")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{
			"adminId":  adminID,
			"username": username,
		})
	}
}
