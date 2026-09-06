package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const sendGridEndpoint = "https://api.sendgrid.com/v3/mail/send"

// SendGridMailer sends email through the SendGrid v3 REST API.
type SendGridMailer struct {
	APIKey    string
	FromEmail string
	FromName  string

	HTTPClient *http.Client
}

// NewSendGridMailer validates inputs and returns a SendGridMailer.
func NewSendGridMailer(apiKey, fromEmail, fromName string) (*SendGridMailer, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("mailer: sendgrid apiKey is required")
	}
	if strings.TrimSpace(fromEmail) == "" {
		return nil, fmt.Errorf("mailer: sendgrid fromEmail is required")
	}
	return &SendGridMailer{
		APIKey:     apiKey,
		FromEmail:  fromEmail,
		FromName:   fromName,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type sendGridPersonalization struct {
	To []sendGridAddress `json:"to"`
}

type sendGridAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridAddress           `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
	Headers          map[string]string         `json:"headers,omitempty"`
}

// Send delivers msg via the SendGrid v3 mail/send endpoint.
func (m *SendGridMailer) Send(ctx context.Context, msg Message) error {
	reqBody := sendGridRequest{
		Personalizations: []sendGridPersonalization{{To: []sendGridAddress{{Email: msg.To}}}},
		From:             sendGridAddress{Email: m.FromEmail, Name: m.FromName},
		Subject:          msg.Subject,
		Content: []sendGridContent{
			{Type: "text/plain", Value: msg.TextBody},
			{Type: "text/html", Value: msg.HTMLBody},
		},
	}
	if msg.UnsubscribeURL != "" {
		reqBody.Headers = map[string]string{"List-Unsubscribe": "<" + msg.UnsubscribeURL + ">"}
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("mailer: encode sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendGridEndpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("mailer: build sendgrid request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: sendgrid request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("mailer: sendgrid returned status %d", resp.StatusCode)
	}
	return nil
}
