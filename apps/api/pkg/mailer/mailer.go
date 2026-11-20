package mailer

// Email is a plain-text email. ReplyTo is optional: when set, the recipient's
// « Répondre » writes to that address instead of From.
type Email struct {
	To      string
	From    string
	ReplyTo string
	Title   string
	Body    string
}

// Mailer defines the interface for sending emails
type Mailer interface {
	SendEmail(email Email) error
	SendTemplatedEmail(toEmail, fromEmail, templateId string, templateData map[string]any) error
}
