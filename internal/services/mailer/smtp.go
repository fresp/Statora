package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// SMTPMailer sends email over SMTP with three transport modes:
//   - "starttls": plain TCP dial, then upgrade via STARTTLS (typical port 587)
//   - "tls":      implicit TLS from the first byte (typical port 465)
//   - "none":     plaintext TCP (local relays only)
type SMTPMailer struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromEmail  string
	FromName   string
	Encryption string // "starttls", "tls", "none"
}

// NewSMTPMailer validates encryption mode and returns an SMTPMailer.
func NewSMTPMailer(host string, port int, username, password, fromEmail, fromName, encryption string) (*SMTPMailer, error) {
	switch encryption {
	case "starttls", "tls", "none":
	default:
		return nil, fmt.Errorf("mailer: invalid smtp encryption %q (want starttls, tls, or none)", encryption)
	}
	if strings.TrimSpace(host) == "" {
		return nil, fmt.Errorf("mailer: smtp host is required")
	}
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("mailer: smtp port %d out of range", port)
	}
	if strings.TrimSpace(fromEmail) == "" {
		return nil, fmt.Errorf("mailer: smtp fromEmail is required")
	}
	return &SMTPMailer{
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		FromEmail:  fromEmail,
		FromName:   fromName,
		Encryption: encryption,
	}, nil
}

// Send delivers msg over the configured SMTP transport.
func (m *SMTPMailer) Send(ctx context.Context, msg Message) error {
	addr := net.JoinHostPort(m.Host, strconv.Itoa(m.Port))

	var conn net.Conn
	var err error
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	switch m.Encryption {
	case "tls":
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: m.Host})
	default:
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("mailer: dial smtp %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("mailer: smtp handshake: %w", err)
	}
	defer client.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if ok, _ := client.Extension("STARTTLS"); ok && m.Encryption == "starttls" {
		if err := client.StartTLS(&tls.Config{ServerName: m.Host}); err != nil {
			return fmt.Errorf("mailer: starttls: %w", err)
		}
	}

	if m.Username != "" && m.Password != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("mailer: smtp auth: %w", err)
			}
		}
	}

	if err := client.Mail(m.FromEmail); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	if err := client.Rcpt(msg.To); err != nil {
		return fmt.Errorf("mailer: RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := w.Write(m.buildMessage(msg)); err != nil {
		w.Close()
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}

	return client.Quit()
}

// buildMessage renders RFC 5322 headers plus a multipart/alternative body.
func (m *SMTPMailer) buildMessage(msg Message) []byte {
	boundary := "statora-boundary-3f9a1c"
	from := m.FromEmail
	if m.FromName != "" {
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", m.FromName), m.FromEmail)
	}

	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + msg.To + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", msg.Subject) + "\r\n")
	if msg.UnsubscribeURL != "" {
		b.WriteString("List-Unsubscribe: <" + msg.UnsubscribeURL + ">\r\n")
	}
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n")
	b.WriteString("\r\n")

	writePart := func(contentType, body string) {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: " + contentType + "; charset=UTF-8\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		b.WriteString("\r\n")
		b.WriteString(body)
		b.WriteString("\r\n")
	}

	writePart("text/plain", msg.TextBody)
	writePart("text/html", msg.HTMLBody)
	b.WriteString("--" + boundary + "--\r\n")

	return []byte(b.String())
}
