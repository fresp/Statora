package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fresp/Statora/configs"
	shared "github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/fresp/Statora/internal/security/pii"
	"github.com/fresp/Statora/internal/services/mailer"
	subscriberservice "github.com/fresp/Statora/internal/services/subscriber"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func subscriberService(db *mongo.Database) *subscriberservice.Service {
	return subscriberservice.NewService(repository.NewMongoSubscriberRepository(db))
}

// wantsHTML reports whether the request is a browser navigation (Accept: text/html)
// rather than an API client call, so the handler can redirect instead of returning JSON.
func wantsHTML(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return accept != "" && strings.Contains(accept, "text/html")
}

// Subscribe accepts a public subscription request and dispatches the double
// opt-in verification email. Never leaks whether an email already exists beyond
// the 409 for active verified duplicates.
func Subscribe(db *mongo.Database, cfg *configs.Config) gin.HandlerFunc {
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

		normalized := pii.Normalize(req.Email)
		service := subscriberService(db)
		sub, err := service.Subscribe(ctx, normalized)
		if err != nil {
			if errors.Is(err, shared.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "email already subscribed"})
				return
			}
			writeDomainError(c, err)
			return
		}

		go sendVerificationEmail(db, cfg, sub)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Please check your inbox to confirm your subscription.",
			"id":      sub.ID,
		})
	}
}

// VerifySubscriber consumes a verification token. Browsers are redirected to
// the status page with a result banner; API clients receive JSON.
func VerifySubscriber(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Query("token"))
		if token == "" {
			if wantsHTML(c) {
				c.Redirect(http.StatusFound, "/?error=invalid_or_expired_token")
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := subscriberService(db).Verify(ctx, token)
		if err != nil {
			if wantsHTML(c) {
				c.Redirect(http.StatusFound, "/?error=invalid_or_expired_token")
				return
			}
			writeDomainError(c, err)
			return
		}

		if wantsHTML(c) {
			c.Redirect(http.StatusFound, "/?verified=true")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "subscription confirmed successfully"})
	}
}

// Unsubscribe consumes an unsubscribe token. Browsers are redirected to the
// status page; API clients receive JSON.
func Unsubscribe(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Query("token"))
		if token == "" {
			if wantsHTML(c) {
				c.Redirect(http.StatusFound, "/?error=invalid_token")
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := subscriberService(db).Unsubscribe(ctx, token)
		if err != nil {
			if wantsHTML(c) {
				c.Redirect(http.StatusFound, "/?error=invalid_token")
				return
			}
			writeDomainError(c, err)
			return
		}

		if wantsHTML(c) {
			c.Redirect(http.StatusFound, "/?unsubscribed=true")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "unsubscribed successfully"})
	}
}

// GetSubscribers lists subscribers for the admin panel.
func GetSubscribers(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, limit, err := parsePaginationParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		subs, total, err := subscriberService(db).List(ctx, page, limit)
		if err != nil {
			writeDomainError(c, err)
			return
		}
		if subs == nil {
			subs = []models.Subscriber{}
		}
		writePaginatedResponse(c, subs, int(total), page, limit)
	}
}

// DeleteSubscriber removes a subscriber record (admin action).
func DeleteSubscriber(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := subscriberService(db).DeleteByID(ctx, id); err != nil {
			writeDomainError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// ResendVerification is an admin action that re-issues a verification email
// for a pending subscriber.
func ResendVerification(db *mongo.Database, cfg *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sub, err := subscriberService(db).ResendVerification(ctx, id)
		if err != nil {
			if errors.Is(err, shared.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "subscriber already verified"})
				return
			}
			writeDomainError(c, err)
			return
		}

		go sendVerificationEmail(db, cfg, sub)

		c.JSON(http.StatusOK, gin.H{"message": "verification email sent"})
	}
}

// sendVerificationEmail renders and sends the double opt-in email. Fire-and-forget.
func sendVerificationEmail(db *mongo.Database, cfg *configs.Config, sub models.Subscriber) {
	baseURL := subscriberBaseURL(db, nil, cfg)
	if baseURL == "" {
		log.Println("[MAIL] Skipping verification email: base URL could not be resolved")
		return
	}

	mailerInstance, err := loadMailer(db, cfg)
	if err != nil {
		log.Printf("[MAIL] Failed to initialize mailer for verification email: %v", err)
		return
	}
	if mailerInstance == nil {
		log.Println("[MAIL] Mail delivery disabled; verification email not sent")
		return
	}

	verifyURL := baseURL + "/api/subscribe/verify?token=" + sub.VerificationToken
	subject, text, html, err := mailer.RenderVerificationEmail(mailer.VerificationEmailData{
		SiteName:        subscriberSiteName(db),
		VerificationURL: verifyURL,
		ExpiresAt:       time.Now().Add(subscriberservice.VerificationTokenTTL),
	})
	if err != nil {
		log.Printf("[MAIL] Failed to render verification email: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := mailerInstance.Send(ctx, mailer.Message{
		To:       sub.Email,
		Subject:  subject,
		TextBody: text,
		HTMLBody: html,
	}); err != nil {
		log.Printf("[MAIL] Failed to send verification email to %s: %v", sub.Email, err)
	}
}

// loadMailer instantiates the configured Mailer from stored settings plus the
// process encryption key. A nil mailer with nil error means delivery is disabled.
func loadMailer(db *mongo.Database, cfg *configs.Config) (mailer.Mailer, error) {
	m, _, err := loadMailerWithSettings(db, cfg)
	return m, err
}

// loadMailerWithSettings returns the mailer plus the stored settings document
// so callers can read BaseURL and branding without a second query.
func loadMailerWithSettings(db *mongo.Database, cfg *configs.Config) (mailer.Mailer, models.StatusPageSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := fetchOrCreateStatusPageSettings(ctx, db)
	if err != nil {
		return nil, models.StatusPageSettings{}, err
	}
	var key []byte
	if cfg != nil {
		key = []byte(cfg.EmailEncryptionKey)
	}
	m, err := mailer.NewMailer(settings.Mail, key)
	return m, settings, err
}

// currentMailProvider reads the stored provider type for diagnostics.
func currentMailProvider(db *mongo.Database) models.MailProviderType {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	settings, err := fetchOrCreateStatusPageSettings(ctx, db)
	if err != nil {
		return models.MailProviderNone
	}
	return settings.Mail.Provider
}

// subscriberSiteName returns the configured site name for email branding.
func subscriberSiteName(db *mongo.Database) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	settings, err := fetchOrCreateStatusPageSettings(ctx, db)
	if err != nil || strings.TrimSpace(settings.Branding.SiteName) == "" {
		return "Statora"
	}
	return settings.Branding.SiteName
}

// subscriberBaseURL resolves the public base URL used in notification links:
// stored Mail.BaseURL first, then the incoming request scheme+host.
func subscriberBaseURL(db *mongo.Database, c *gin.Context, cfg *configs.Config) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := fetchOrCreateStatusPageSettings(ctx, db)
	if err == nil && strings.TrimSpace(settings.Mail.BaseURL) != "" {
		return strings.TrimRight(settings.Mail.BaseURL, "/")
	}
	if c == nil {
		return ""
	}
	scheme := "http"
	if c.Request != nil && c.Request.TLS != nil {
		scheme = "https"
	}
	if host := c.Request.Host; host != "" {
		return scheme + "://" + host
	}
	return ""
}
