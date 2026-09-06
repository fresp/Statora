package mailer

import (
	"fmt"
	"log"

	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/security/pii"
)

// NewMailer builds a Mailer from stored settings, decrypting provider secrets
// with the 32-byte application encryption key. A nil Mailer with nil error
// means mail delivery is disabled (provider "none").
func NewMailer(settings models.MailSettings, encryptionKey []byte) (Mailer, error) {
	switch settings.Provider {
	case models.MailProviderNone, "":
		return nil, nil
	case models.MailProviderSMTP:
		password := ""
		if settings.SMTP.Password != "" {
			plain, err := pii.Decrypt(settings.SMTP.Password, encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("mailer: decrypt smtp password: %w", err)
			}
			password = plain
		}
		return NewSMTPMailer(
			settings.SMTP.Host,
			settings.SMTP.Port,
			settings.SMTP.Username,
			password,
			settings.SMTP.FromEmail,
			settings.SMTP.FromName,
			settings.SMTP.Encryption,
		)
	case models.MailProviderSendGrid:
		apiKey := ""
		if settings.SendGrid.APIKey != "" {
			plain, err := pii.Decrypt(settings.SendGrid.APIKey, encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("mailer: decrypt sendgrid api key: %w", err)
			}
			apiKey = plain
		}
		return NewSendGridMailer(apiKey, settings.SendGrid.FromEmail, settings.SendGrid.FromName)
	default:
		log.Printf("[MAIL] Unknown mail provider %q, skipping delivery", settings.Provider)
		return nil, nil
	}
}
