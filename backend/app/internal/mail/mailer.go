package mail

import "context"

type Mailer interface {
	SendRegistrationLink(ctx context.Context, email, link string) error
}
