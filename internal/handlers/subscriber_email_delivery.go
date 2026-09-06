package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fresp/Statora/configs"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/security/pii"
	"github.com/fresp/Statora/internal/services/mailer"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// subscriberEventTypes supported by DispatchSubscriberEmail.
const (
	eventIncidentCreated    = "incident_created"
	eventIncidentResolved   = "incident_resolved"
	eventMaintenanceCreated = "maintenance_created"
	eventMaintenanceUpdated = "maintenance_updated"
)

// DispatchSubscriberEmail sends a notification email about eventType to every
// verified, subscribed recipient. Fire-and-forget, mirroring DispatchWebhookEvent:
// best-effort, never panics, failures logged individually.
func DispatchSubscriberEmail(db *mongo.Database, cfg *configs.Config, eventType string, data interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[MAIL] Subscriber email dispatch panicked: %v", r)
			}
		}()

		mailerInstance, settings, err := loadMailerWithSettings(db, cfg)
		if err != nil {
			log.Printf("[MAIL] Failed to initialize mailer for %s: %v", eventType, err)
			return
		}
		if mailerInstance == nil {
			log.Printf("[MAIL] Mail delivery disabled; skipping %s subscriber emails", eventType)
			return
		}

		subscribers, err := subscriberService(db).ListVerified(context.Background())
		if err != nil {
			log.Printf("[MAIL] Failed to load subscribers for %s: %v", eventType, err)
			return
		}
		if len(subscribers) == 0 {
			return
		}

		baseURL := strings.TrimRight(settings.Mail.BaseURL, "/")
		siteName := settings.Branding.SiteName
		if siteName == "" {
			siteName = "Statora"
		}

		for _, sub := range subscribers {
			msg, err := buildSubscriberMessage(eventType, data, sub, siteName, baseURL)
			if err != nil {
				log.Printf("[MAIL] Failed to render %s email for %s: %v", eventType, sub.Email, err)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			err = mailerInstance.Send(ctx, msg)
			cancel()
			if err != nil {
				log.Printf("[MAIL] Failed to send %s email to %s: %v", eventType, sub.Email, err)
			}
		}
	}()
}

// buildSubscriberMessage renders the per-subscriber message with a unique
// unsubscribe link.
func buildSubscriberMessage(eventType string, data interface{}, sub models.Subscriber, siteName, baseURL string) (mailer.Message, error) {
	unsubscribeURL := ""
	if baseURL != "" {
		unsubscribeURL = baseURL + "/api/subscribe/unsubscribe?token=" + sub.UnsubscribeToken
	}

	switch eventType {
	case eventIncidentCreated, eventIncidentResolved:
		incident, ok := data.(models.Incident)
		if !ok {
			return mailer.Message{}, errUnexpectedPayload(eventType)
		}
		emailData := mailer.IncidentEmailData{
			SiteName:       siteName,
			IncidentTitle:  incident.Title,
			IncidentStatus: string(incident.Status),
			Impact:         string(incident.Impact),
			Description:    incident.Description,
			StartedAt:      incident.CreatedAt,
			IncidentURL:    statusPagePath(baseURL, "/history"),
			UnsubscribeURL: unsubscribeURL,
		}
		if incident.ResolvedAt != nil {
			emailData.ResolvedAt = *incident.ResolvedAt
		}
		if eventType == eventIncidentCreated {
			subject, text, html, err := mailer.RenderIncidentCreatedEmail(emailData)
			return mailer.Message{To: sub.Email, Subject: subject, TextBody: text, HTMLBody: html, UnsubscribeURL: unsubscribeURL}, err
		}
		subject, text, html, err := mailer.RenderIncidentResolvedEmail(emailData)
		return mailer.Message{To: sub.Email, Subject: subject, TextBody: text, HTMLBody: html, UnsubscribeURL: unsubscribeURL}, err

	case eventMaintenanceCreated, eventMaintenanceUpdated:
		m, ok := data.(models.Maintenance)
		if !ok {
			return mailer.Message{}, errUnexpectedPayload(eventType)
		}
		emailData := mailer.MaintenanceEmailData{
			SiteName:       siteName,
			Title:          m.Title,
			Description:    m.Description,
			StartTime:      m.StartTime,
			EndTime:        m.EndTime,
			UnsubscribeURL: unsubscribeURL,
		}
		var subject, text, html string
		var err error
		if eventType == eventMaintenanceCreated {
			subject, text, html, err = mailer.RenderMaintenanceCreatedEmail(emailData)
		} else {
			subject, text, html, err = mailer.RenderMaintenanceUpdatedEmail(emailData)
		}
		return mailer.Message{To: sub.Email, Subject: subject, TextBody: text, HTMLBody: html, UnsubscribeURL: unsubscribeURL}, err

	default:
		return mailer.Message{}, errUnexpectedPayload(eventType)
	}
}

func errUnexpectedPayload(eventType string) error {
	return fmt.Errorf("unexpected payload type for event %s", eventType)
}

// statusPagePath joins the base URL with an app path.
func statusPagePath(baseURL, path string) string {
	if baseURL == "" {
		return path
	}
	return baseURL + path
}

// SendTestEmail is an admin endpoint that sends a test message using the
// currently stored mail settings.
func SendTestEmail(db *mongo.Database, cfg *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			To string `json:"to" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		mailerInstance, _, err := loadMailerWithSettings(db, cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize mail settings: " + err.Error()})
			return
		}
		if mailerInstance == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mail delivery is not configured; select a provider and save settings first"})
			return
		}

		subject, text, html, err := mailer.RenderTestEmail(mailer.TestEmailData{
			SiteName:  subscriberSiteName(db),
			Provider:  string(currentMailProvider(db)),
			SentAt:    time.Now(),
			Recipient: pii.Normalize(req.To),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render test email"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		defer cancel()

		if err := mailerInstance.Send(ctx, mailer.Message{
			To:       pii.Normalize(req.To),
			Subject:  subject,
			TextBody: text,
			HTMLBody: html,
		}); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "test email failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "test email sent"})
	}
}
