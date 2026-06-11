// Package mail is the email transport: a Mailer port plus a couple of
// implementations. It is pure infrastructure (no business logic) — modules
// depend on the Mailer interface and the notifications module composes the
// actual messages. Driver "log" prints emails for local development; "smtp"
// sends via a server such as Mailpit (locally) or a provider (production).
package mail

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/yourorg/goapp/internal/platform/config"
)

// Message is a single outbound email.
type Message struct {
	To      []string
	Subject string
	Text    string
}

// Mailer sends email. Implementations must be safe for concurrent use.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// New selects a Mailer from configuration.
func New(cfg config.MailConfig, logger *slog.Logger) Mailer {
	if strings.EqualFold(cfg.Driver, "smtp") {
		return &SMTPMailer{cfg: cfg}
	}
	return &LogMailer{from: cfg.From, logger: logger}
}

// LogMailer logs emails instead of sending them — the default for local
// development and tests.
type LogMailer struct {
	from   string
	logger *slog.Logger
}

func (m *LogMailer) Send(_ context.Context, msg Message) error {
	m.logger.Info("email (not sent: log driver)",
		slog.String("from", m.from),
		slog.String("to", strings.Join(msg.To, ", ")),
		slog.String("subject", msg.Subject),
		slog.String("body", msg.Text),
	)
	return nil
}

// SMTPMailer sends email over SMTP (Mailpit/Mailhog locally, a provider in
// production). It uses PLAIN auth when a username is configured.
type SMTPMailer struct {
	cfg config.MailConfig
}

func (m *SMTPMailer) Send(_ context.Context, msg Message) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.SMTP.Host, m.cfg.SMTP.Port)

	var auth smtp.Auth
	if m.cfg.SMTP.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTP.Username, m.cfg.SMTP.Password, m.cfg.SMTP.Host)
	}

	if err := smtp.SendMail(addr, auth, m.cfg.From, msg.To, m.build(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

// build renders a minimal RFC 5322 text/plain message.
func (m *SMTPMailer) build(msg Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", m.cfg.From)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(msg.To, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Text)
	return []byte(b.String())
}
