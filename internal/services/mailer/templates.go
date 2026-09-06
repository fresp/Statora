package mailer

import (
	"fmt"
	"html"
	"strings"
	"time"
)

// VerificationEmailData renders the double opt-in confirmation email.
type VerificationEmailData struct {
	SiteName         string
	VerificationURL  string
	ExpiresAt        time.Time
	SupportReference string
}

// IncidentEmailData renders incident created/resolved notifications.
type IncidentEmailData struct {
	SiteName       string
	IncidentTitle  string
	IncidentStatus string
	Impact         string
	Description    string
	StartedAt      time.Time
	ResolvedAt     time.Time
	IncidentURL    string
	UnsubscribeURL string
}

// MaintenanceEmailData renders maintenance created/updated notifications.
type MaintenanceEmailData struct {
	SiteName       string
	Title          string
	Description    string
	StartTime      time.Time
	EndTime        time.Time
	UnsubscribeURL string
}

// TestEmailData renders the admin test email.
type TestEmailData struct {
	SiteName  string
	Provider  string
	SentAt    time.Time
	Recipient string
}

const layoutCSS = "max-width:560px;margin:0 auto;padding:24px;font-family:Arial,Helvetica,sans-serif;color:#1f2937;line-height:1.5"

func renderHTMLDocument(title, bodyHTML, unsubscribeURL string) string {
	unsub := ""
	if unsubscribeURL != "" {
		unsub = fmt.Sprintf(`<p style="font-size:12px;color:#6b7280;margin-top:24px">You are receiving this because you subscribed to status updates. <a href="%s" style="color:#6b7280">Unsubscribe</a></p>`, html.EscapeString(unsubscribeURL))
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="margin:0;background:#f9fafb">
<div style="%s">
<h2 style="margin:0 0 16px;font-size:18px">%s</h2>
%s
%s
</div>
</body>
</html>`, layoutCSS, html.EscapeString(title), bodyHTML, unsub)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

// RenderVerificationEmail returns subject, plain text, and HTML for the opt-in email.
func RenderVerificationEmail(data VerificationEmailData) (string, string, string, error) {
	if strings.TrimSpace(data.VerificationURL) == "" {
		return "", "", "", fmt.Errorf("mailer: verification URL is required")
	}
	site := data.SiteName
	if site == "" {
		site = "Status"
	}
	subject := "Confirm your " + site + " subscription"

	text := fmt.Sprintf("Hello,\n\nPlease confirm your subscription to %s status updates:\n\n%s\n\nThis link expires at %s.\nIf you did not request this, you can ignore this email.\n", site, data.VerificationURL, formatTime(data.ExpiresAt))
	bodyHTML := fmt.Sprintf(`<p>Hello,</p><p>Please confirm your subscription to <strong>%s</strong> status updates:</p><p><a href="%s" style="display:inline-block;background:#16a34a;color:#ffffff;padding:10px 18px;border-radius:6px;text-decoration:none">Confirm subscription</a></p><p style="font-size:13px;color:#6b7280">Or copy this link: %s</p><p style="font-size:13px;color:#6b7280">This link expires at %s. If you did not request this, ignore this email.</p>`, html.EscapeString(site), html.EscapeString(data.VerificationURL), html.EscapeString(data.VerificationURL), formatTime(data.ExpiresAt))

	return subject, text, renderHTMLDocument("Confirm your subscription", bodyHTML, ""), nil
}

// RenderIncidentCreatedEmail returns subject, plain text, and HTML for a new incident.
func RenderIncidentCreatedEmail(data IncidentEmailData) (string, string, string, error) {
	site := data.SiteName
	if site == "" {
		site = "Status"
	}
	impact := strings.ToUpper(strings.TrimSpace(data.Impact))
	if impact == "" {
		impact = "NOTICE"
	}
	subject := fmt.Sprintf("[%s] New incident: %s", site, data.IncidentTitle)

	text := fmt.Sprintf("A new incident was created on %s.\n\nTitle: %s\nImpact: %s\nStarted: %s\n\n%s\n\nView details: %s\nUnsubscribe: %s\n", site, data.IncidentTitle, impact, formatTime(data.StartedAt), data.Description, data.IncidentURL, data.UnsubscribeURL)
	bodyHTML := fmt.Sprintf(`<p>A new incident was created on <strong>%s</strong>.</p><p><strong>Title:</strong> %s<br/><strong>Impact:</strong> %s<br/><strong>Started:</strong> %s</p><div>%s</div><p><a href="%s">View details</a></p>`, html.EscapeString(site), html.EscapeString(data.IncidentTitle), html.EscapeString(impact), formatTime(data.StartedAt), renderDescriptionHTML(data.Description), html.EscapeString(data.IncidentURL))

	return subject, text, renderHTMLDocument(subject, bodyHTML, data.UnsubscribeURL), nil
}

// RenderIncidentResolvedEmail returns subject, plain text, and HTML for a resolved incident.
func RenderIncidentResolvedEmail(data IncidentEmailData) (string, string, string, error) {
	site := data.SiteName
	if site == "" {
		site = "Status"
	}
	subject := fmt.Sprintf("[%s] Resolved: %s", site, data.IncidentTitle)

	text := fmt.Sprintf("The incident \"%s\" on %s has been resolved at %s.\n\n%s\n\nView details: %s\nUnsubscribe: %s\n", data.IncidentTitle, site, formatTime(data.ResolvedAt), data.Description, data.IncidentURL, data.UnsubscribeURL)
	bodyHTML := fmt.Sprintf(`<p style="color:#16a34a;font-weight:600">✔ Incident resolved</p><p>The incident <strong>%s</strong> on <strong>%s</strong> has been resolved at %s.</p><div>%s</div><p><a href="%s">View details</a></p>`, html.EscapeString(data.IncidentTitle), html.EscapeString(site), formatTime(data.ResolvedAt), renderDescriptionHTML(data.Description), html.EscapeString(data.IncidentURL))

	return subject, text, renderHTMLDocument(subject, bodyHTML, data.UnsubscribeURL), nil
}

// RenderMaintenanceCreatedEmail returns subject, plain text, and HTML for scheduled maintenance.
func RenderMaintenanceCreatedEmail(data MaintenanceEmailData) (string, string, string, error) {
	site := data.SiteName
	if site == "" {
		site = "Status"
	}
	subject := fmt.Sprintf("[%s] Scheduled maintenance: %s", site, data.Title)

	text := maintenanceText(site, data)
	bodyHTML := maintenanceHTML(site, data)
	return subject, text, renderHTMLDocument(subject, bodyHTML, data.UnsubscribeURL), nil
}

// RenderMaintenanceUpdatedEmail returns subject, plain text, and HTML for maintenance changes.
func RenderMaintenanceUpdatedEmail(data MaintenanceEmailData) (string, string, string, error) {
	site := data.SiteName
	if site == "" {
		site = "Status"
	}
	subject := fmt.Sprintf("[%s] Maintenance updated: %s", site, data.Title)

	text := maintenanceText(site, data)
	bodyHTML := maintenanceHTML(site, data)
	return subject, text, renderHTMLDocument(subject, bodyHTML, data.UnsubscribeURL), nil
}

// RenderTestEmail returns subject, plain text, and HTML for the admin test email.
func RenderTestEmail(data TestEmailData) (string, string, string, error) {
	site := data.SiteName
	if site == "" {
		site = "Status"
	}
	subject := fmt.Sprintf("[%s] Test email", site)
	provider := data.Provider
	if provider == "" {
		provider = "unknown"
	}

	text := fmt.Sprintf("This is a test email from %s.\n\nProvider: %s\nSent at: %s\nRecipient: %s\n\nIf you received this, your mail configuration works.\n", site, provider, formatTime(data.SentAt), data.Recipient)
	bodyHTML := fmt.Sprintf(`<p>This is a <strong>test email</strong> from <strong>%s</strong>.</p><p><strong>Provider:</strong> %s<br/><strong>Sent at:</strong> %s<br/><strong>Recipient:</strong> %s</p><p>If you received this, your mail configuration works.</p>`, html.EscapeString(site), html.EscapeString(provider), formatTime(data.SentAt), html.EscapeString(data.Recipient))

	return subject, text, renderHTMLDocument(subject, bodyHTML, ""), nil
}

func maintenanceText(site string, data MaintenanceEmailData) string {
	return fmt.Sprintf("Maintenance on %s has been scheduled or updated.\n\nTitle: %s\nWindow: %s → %s\n\n%s\n", site, data.Title, formatTime(data.StartTime), formatTime(data.EndTime), data.Description)
}

func maintenanceHTML(site string, data MaintenanceEmailData) string {
	return fmt.Sprintf(`<p>Maintenance on <strong>%s</strong> has been scheduled or updated.</p><p><strong>Title:</strong> %s<br/><strong>Window:</strong> %s → %s</p><div>%s</div>`, html.EscapeString(site), html.EscapeString(data.Title), formatTime(data.StartTime), formatTime(data.EndTime), renderDescriptionHTML(data.Description))
}

func renderDescriptionHTML(description string) string {
	if strings.TrimSpace(description) == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(description), "\n")
	escaped := make([]string, 0, len(lines))
	for _, line := range lines {
		escaped = append(escaped, html.EscapeString(line))
	}
	return "<p style=\"white-space:pre-line\">" + strings.Join(escaped, "\n") + "</p>"
}
