package testing

import (
	"sync"

	"ddd-second-hand-marketplace/pkg/mailer"
)

// SentTemplatedEmail represents a templated email that was sent
type SentTemplatedEmail struct {
	To           string
	From         string
	TemplateId   string
	TemplateData map[string]any
}

// MailerSpy is a test spy that captures sent emails for verification
type MailerSpy struct {
	sentEmails          []mailer.Email
	sentTemplatedEmails []SentTemplatedEmail
	mu                  sync.Mutex
}

// NewMailerSpy creates a new MailerSpy instance
func NewMailerSpy() *MailerSpy {
	return &MailerSpy{
		sentEmails:          make([]mailer.Email, 0),
		sentTemplatedEmails: make([]SentTemplatedEmail, 0),
	}
}

// SendEmail captures the email, ReplyTo included, for testing
func (m *MailerSpy) SendEmail(email mailer.Email) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentEmails = append(m.sentEmails, email)
	return nil
}

// SendTemplatedEmail captures the templated email for testing
func (m *MailerSpy) SendTemplatedEmail(toEmail, fromEmail, templateId string, templateData map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentTemplatedEmails = append(m.sentTemplatedEmails, SentTemplatedEmail{
		To:           toEmail,
		From:         fromEmail,
		TemplateId:   templateId,
		TemplateData: templateData,
	})

	return nil
}

// GetSentEmails returns all captured emails
func (m *MailerSpy) GetSentEmails() []mailer.Email {
	m.mu.Lock()
	defer m.mu.Unlock()

	emails := make([]mailer.Email, len(m.sentEmails))
	copy(emails, m.sentEmails)
	return emails
}

// GetSentTemplatedEmails returns all captured templated emails
func (m *MailerSpy) GetSentTemplatedEmails() []SentTemplatedEmail {
	m.mu.Lock()
	defer m.mu.Unlock()

	emails := make([]SentTemplatedEmail, len(m.sentTemplatedEmails))
	copy(emails, m.sentTemplatedEmails)
	return emails
}

// Reset clears all captured emails
func (m *MailerSpy) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentEmails = make([]mailer.Email, 0)
	m.sentTemplatedEmails = make([]SentTemplatedEmail, 0)
}
