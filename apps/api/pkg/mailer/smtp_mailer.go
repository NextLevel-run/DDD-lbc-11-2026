package mailer

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"
)

// SMTPMailer sends emails through an SMTP server without authentication,
// such as Mailcatcher in development.
type SMTPMailer struct {
	addr string
}

// NewSMTPMailer creates an SMTPMailer that talks to the server at addr ("host:port").
func NewSMTPMailer(addr string) *SMTPMailer {
	return &SMTPMailer{addr: addr}
}

// SendEmail sends a plain-text email, with a Reply-To header only when
// email.ReplyTo is set. It refuses a line break in a header, which would let
// the caller inject headers such as Bcc, and encodes a non-ASCII subject so
// the headers stay ASCII.
func (m *SMTPMailer) SendEmail(email Email) error {
	if err := checkAddress("to", email.To); err != nil {
		return err
	}
	if err := checkAddress("from", email.From); err != nil {
		return err
	}
	if email.ReplyTo != "" {
		if err := checkAddress("reply-to", email.ReplyTo); err != nil {
			return err
		}
	}
	if strings.ContainsAny(email.Title, "\r\n") {
		return fmt.Errorf("invalid subject: line break in a header")
	}
	headers := []string{"From: " + email.From, "To: " + email.To}
	if email.ReplyTo != "" {
		headers = append(headers, "Reply-To: "+email.ReplyTo)
	}
	msg := strings.Join(append(headers,
		"Subject: "+mime.QEncoding.Encode("utf-8", email.Title),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		email.Body,
	), "\r\n")
	if err := smtp.SendMail(m.addr, nil, email.From, []string{email.To}, []byte(msg)); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

// SendTemplatedEmail has no template engine to render with: it sends the
// template id as the subject and the template data as an indented JSON body.
func (m *SMTPMailer) SendTemplatedEmail(toEmail, fromEmail, templateId string, templateData map[string]any) error {
	data, err := json.MarshalIndent(templateData, "", "  ")
	if err != nil {
		return fmt.Errorf("encode template data of %s: %w", templateId, err)
	}
	return m.SendEmail(Email{To: toEmail, From: fromEmail, Title: templateId, Body: string(data)})
}

// checkAddress accepts a bare email address only: no display name, no line break.
func checkAddress(field, address string) error {
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Address != address {
		return fmt.Errorf("invalid %s address", field)
	}
	return nil
}
