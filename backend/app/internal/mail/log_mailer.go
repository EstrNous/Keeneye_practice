package mail

import (
	"context"
	"log/slog"
)

type LogMailer struct{}

func NewLogMailer() *LogMailer {
	return &LogMailer{}
}

func (m *LogMailer) SendRegistrationLink(ctx context.Context, email, link string) error {
	slog.InfoContext(ctx, "registration email sent (log mailer)",
		"email", email,
		"link", link,
	)
	return nil
}
