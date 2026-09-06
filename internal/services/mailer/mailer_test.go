package mailer

import (
	"strings"
	"testing"
	"time"
)

func TestRenderVerificationEmail(t *testing.T) {
	expires := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	subject, text, html, err := RenderVerificationEmail(VerificationEmailData{
		SiteName:        "Statora",
		VerificationURL: "https://status.example.com/api/subscribe/verify?token=abc123",
		ExpiresAt:       expires,
	})
	if err != nil {
		t.Fatalf("RenderVerificationEmail() error = %v", err)
	}

	if !strings.Contains(subject, "Statora") {
		t.Errorf("subject = %q, want site name", subject)
	}
	if !strings.Contains(text, "https://status.example.com/api/subscribe/verify?token=abc123") {
		t.Errorf("text body missing verification URL, got %q", text)
	}
	if !strings.Contains(text, "2026-09-08 12:00 UTC") {
		t.Errorf("text body missing expiry, got %q", text)
	}
	for _, want := range []string{"Confirm subscription", "https://status.example.com/api/subscribe/verify?token=abc123", "2026-09-08 12:00 UTC"} {
		if !strings.Contains(html, want) {
			t.Errorf("html body missing %q", want)
		}
	}
}

func TestRenderVerificationEmailRequiresURL(t *testing.T) {
	_, _, _, err := RenderVerificationEmail(VerificationEmailData{SiteName: "Statora"})
	if err == nil {
		t.Fatal("RenderVerificationEmail() error = nil, want error for empty verification URL")
	}
}

func TestRenderIncidentCreatedEmail(t *testing.T) {
	started := time.Date(2026, 9, 6, 10, 30, 0, 0, time.UTC)
	subject, text, html, err := RenderIncidentCreatedEmail(IncidentEmailData{
		SiteName:       "Statora",
		IncidentTitle:  "API <outage>",
		Impact:         "major",
		Description:    "Investigating elevated errors",
		StartedAt:      started,
		IncidentURL:    "https://status.example.com/history",
		UnsubscribeURL: "https://status.example.com/api/subscribe/unsubscribe?token=tok",
	})
	if err != nil {
		t.Fatalf("RenderIncidentCreatedEmail() error = %v", err)
	}

	if !strings.Contains(subject, "New incident") || !strings.Contains(subject, "API <outage>") {
		t.Errorf("subject = %q, want incident title", subject)
	}
	// HTML must escape the title.
	if strings.Contains(html, "API <outage>") {
		t.Errorf("html body contains unescaped title")
	}
	if !strings.Contains(html, "API &lt;outage&gt;") {
		t.Errorf("html body missing escaped title")
	}
	if !strings.Contains(html, "MAJOR") {
		t.Errorf("html body missing impact, got %q", html)
	}
	// Plain text carries the unsubscribe URL; the HTML footer wraps it in a link tag.
	if !strings.Contains(text, "https://status.example.com/api/subscribe/unsubscribe?token=tok") {
		t.Errorf("text body missing unsubscribe URL")
	}
	if !strings.Contains(html, `href="https://status.example.com/api/subscribe/unsubscribe?token=tok"`) {
		t.Errorf("html body missing unsubscribe link")
	}
}

func TestRenderIncidentResolvedEmail(t *testing.T) {
	resolved := time.Date(2026, 9, 6, 11, 45, 0, 0, time.UTC)
	subject, text, html, err := RenderIncidentResolvedEmail(IncidentEmailData{
		SiteName:      "Statora",
		IncidentTitle: "API outage",
		ResolvedAt:    resolved,
		Description:   "All systems back to normal",
	})
	if err != nil {
		t.Fatalf("RenderIncidentResolvedEmail() error = %v", err)
	}

	if !strings.Contains(subject, "Resolved") {
		t.Errorf("subject = %q, want resolved marker", subject)
	}
	if !strings.Contains(text, "2026-09-06 11:45 UTC") {
		t.Errorf("text body missing resolved time, got %q", text)
	}
	if !strings.Contains(html, "resolved") {
		t.Errorf("html body missing resolved marker")
	}
}

func TestRenderMaintenanceEmails(t *testing.T) {
	start := time.Date(2026, 9, 7, 2, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 7, 4, 0, 0, 0, time.UTC)
	data := MaintenanceEmailData{
		SiteName:    "Statora",
		Title:       "DB index rebuild",
		Description: "Planned database work",
		StartTime:   start,
		EndTime:     end,
	}

	createdSubject, createdText, createdHTML, err := RenderMaintenanceCreatedEmail(data)
	if err != nil {
		t.Fatalf("RenderMaintenanceCreatedEmail() error = %v", err)
	}
	if !strings.Contains(createdSubject, "Scheduled maintenance") {
		t.Errorf("created subject = %q", createdSubject)
	}
	if !strings.Contains(createdText, "2026-09-07 02:00 UTC") || !strings.Contains(createdText, "04:00 UTC") {
		t.Errorf("created text missing window, got %q", createdText)
	}
	if !strings.Contains(createdHTML, "DB index rebuild") {
		t.Errorf("created html missing title")
	}

	updatedSubject, _, _, err := RenderMaintenanceUpdatedEmail(data)
	if err != nil {
		t.Fatalf("RenderMaintenanceUpdatedEmail() error = %v", err)
	}
	if !strings.Contains(updatedSubject, "Maintenance updated") {
		t.Errorf("updated subject = %q", updatedSubject)
	}
}

func TestRenderTestEmail(t *testing.T) {
	subject, text, html, err := RenderTestEmail(TestEmailData{
		SiteName:  "Statora",
		Provider:  "smtp",
		SentAt:    time.Date(2026, 9, 6, 9, 0, 0, 0, time.UTC),
		Recipient: "admin@example.com",
	})
	if err != nil {
		t.Fatalf("RenderTestEmail() error = %v", err)
	}
	if !strings.Contains(subject, "Test email") {
		t.Errorf("subject = %q", subject)
	}
	for _, body := range []string{text, html} {
		if !strings.Contains(body, "smtp") || !strings.Contains(body, "admin@example.com") {
			t.Errorf("body missing provider/recipient: %q", body)
		}
	}
}

func TestSMTPMailerBuildMessage(t *testing.T) {
	m, err := NewSMTPMailer("smtp.example.com", 587, "user", "pass", "from@example.com", "Statora", "starttls")
	if err != nil {
		t.Fatalf("NewSMTPMailer() error = %v", err)
	}

	raw := string(m.buildMessage(Message{
		To:             "sub@example.com",
		Subject:        "Hello",
		TextBody:       "plain text",
		HTMLBody:       "<p>html</p>",
		UnsubscribeURL: "https://status.example.com/api/subscribe/unsubscribe?token=t",
	}))

	for _, want := range []string{
		"From: Statora <from@example.com>",
		"To: sub@example.com",
		"List-Unsubscribe: <https://status.example.com/api/subscribe/unsubscribe?token=t>",
		"MIME-Version: 1.0",
		`Content-Type: multipart/alternative; boundary="statora-boundary-3f9a1c"`,
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Type: text/html; charset=UTF-8",
		"plain text",
		"<p>html</p>",
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("message missing %q", want)
		}
	}
	if !strings.Contains(raw, "--statora-boundary-3f9a1c--") {
		t.Errorf("message missing closing boundary")
	}
}

func TestSMTPMailerValidation(t *testing.T) {
	if _, err := NewSMTPMailer("", 587, "", "", "from@example.com", "", "starttls"); err == nil {
		t.Error("empty host accepted")
	}
	if _, err := NewSMTPMailer("h", 0, "", "", "from@example.com", "", "starttls"); err == nil {
		t.Error("invalid port accepted")
	}
	if _, err := NewSMTPMailer("h", 587, "", "", "", "", "starttls"); err == nil {
		t.Error("empty fromEmail accepted")
	}
	if _, err := NewSMTPMailer("h", 587, "", "", "from@example.com", "", "bogus"); err == nil {
		t.Error("invalid encryption accepted")
	}
}

func TestSendGridMailerValidation(t *testing.T) {
	if _, err := NewSendGridMailer("", "from@example.com", ""); err == nil {
		t.Error("empty apiKey accepted")
	}
	if _, err := NewSendGridMailer("key", "", ""); err == nil {
		t.Error("empty fromEmail accepted")
	}
}
