// Package notifications is the notifications bounded context: it owns what users
// are notified about and how messages are composed, and delivers them through the
// platform transports (currently platform/mail). Other modules don't send email
// themselves — they ask notifications via a small port (see auth.Notifier), which
// keeps message composition and delivery in one place.
package notifications

import (
	"fmt"

	"github.com/yourorg/goapp/internal/platform/mail"
)

// renderWelcome composes the welcome email for a newly registered user. Kept
// separate from delivery so message content is easy to test and to evolve into
// real templates.
func renderWelcome(firstName, email string) mail.Message {
	greeting := firstName
	if greeting == "" {
		greeting = "there"
	}
	return mail.Message{
		To:      []string{email},
		Subject: "Welcome to goapp",
		Text: fmt.Sprintf(
			"Hi %s,\n\nWelcome to goapp — your account is ready.\n\nThanks for signing up!\n",
			greeting,
		),
	}
}
