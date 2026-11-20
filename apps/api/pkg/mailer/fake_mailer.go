package mailer

import (
	"fmt"
)

// FakeMailer is a fake implementation of the Mailer interface
// that logs emails to the console instead of actually sending them
type FakeMailer struct{}

// NewFakeMailer creates a new FakeMailer instance
func NewFakeMailer() *FakeMailer {
	return &FakeMailer{}
}

// SendEmail logs the email to the console
func (m *FakeMailer) SendEmail(email Email) error {
	replyTo := ""
	if email.ReplyTo != "" {
		replyTo = " (reply-to " + email.ReplyTo + ")"
	}
	fmt.Printf("Send email from %s%s to %s: %s\n", email.From, replyTo, email.To, email.Title)
	return nil
}

// SendTemplatedEmail logs the templated email to the console
func (m *FakeMailer) SendTemplatedEmail(toEmail, fromEmail, templateId string, templateData map[string]any) error {
	fmt.Printf("Send templated email from %s to %s using template %s\n", fromEmail, toEmail, templateId)
	return nil
}
