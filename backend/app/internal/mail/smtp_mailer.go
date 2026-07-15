package mail

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPMailer(host, port, username, password, from string) *SMTPMailer {
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *SMTPMailer) SendRegistrationLink(ctx context.Context, email, link string) error {
	_ = ctx
	subject := "Complete your registration"
	body := fmt.Sprintf("Please complete your registration by visiting:\n\n%s\n", link)
	msg := strings.Join([]string{
		fmt.Sprintf("To: %s", email),
		fmt.Sprintf("From: %s", m.from),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	return smtp.SendMail(addr, auth, m.from, []string{email}, []byte(msg))
}
