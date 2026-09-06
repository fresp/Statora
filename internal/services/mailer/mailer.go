// Package mailer implements pluggable outbound email delivery for subscriber
// notifications. It supports generic SMTP (STARTTLS, implicit TLS, plaintext)
// and the SendGrid v3 REST API, chosen at runtime from admin mail settings.
package mailer

import (
	"context"
)

// Message is a provider-agnostic outbound email.
type Message struct {
	To             string
	Subject        string
	TextBody       string
	HTMLBody       string
	UnsubscribeURL string
}

// Mailer sends a single message. Implementations must be safe for concurrent use.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}
