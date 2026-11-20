package mailer_test

import (
	"bufio"
	"mime"
	"net"
	"net/mail"
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/pkg/mailer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// smtpServer is a minimal SMTP server that answers every command and
// sends the DATA section of the first message on its channel.
func smtpServer(t *testing.T) (string, <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	received := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		reply := func(s string) { conn.Write([]byte(s + "\r\n")) }

		reply("220 localhost ready")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			switch cmd := strings.ToUpper(strings.TrimSpace(line)); {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				reply("250 localhost")
			case cmd == "DATA":
				reply("354 end with <CRLF>.<CRLF>")
				var data strings.Builder
				for {
					l, err := r.ReadString('\n')
					if err != nil || l == ".\r\n" {
						break
					}
					data.WriteString(l)
				}
				received <- data.String()
				reply("250 OK")
			case cmd == "QUIT":
				reply("221 bye")
				return
			default:
				reply("250 OK")
			}
		}
	}()
	return ln.Addr().String(), received
}

// receive waits for the message the server got, and fails the test instead
// of hanging when none arrives.
func receive(t *testing.T, received <-chan string) string {
	t.Helper()
	select {
	case msg := <-received:
		return msg
	case <-time.After(2 * time.Second):
		t.Fatal("the SMTP server received no message")
		return ""
	}
}

func TestSMTPMailer_SendEmail(t *testing.T) {
	addr, received := smtpServer(t)

	err := mailer.NewSMTPMailer(addr).SendEmail(mailer.Email{To: "bob@example.com", From: "shop@example.com", Title: "Hello", Body: "Your ad is online"})

	require.NoError(t, err)
	msg := receive(t, received)
	assert.Contains(t, msg, "From: shop@example.com")
	assert.Contains(t, msg, "To: bob@example.com")
	assert.Contains(t, msg, "Subject: Hello")
	assert.Contains(t, msg, "Your ad is online")
	assert.NotContains(t, msg, "Reply-To:", "no Reply-To header without ReplyTo")
}

func TestSMTPMailer_SendEmailWithReplyTo(t *testing.T) {
	addr, received := smtpServer(t)

	err := mailer.NewSMTPMailer(addr).SendEmail(mailer.Email{
		To: "bob@example.com", From: "shop@example.com", ReplyTo: "alice@example.com", Title: "Hello", Body: "body",
	})

	require.NoError(t, err)
	msg, err := mail.ReadMessage(strings.NewReader(receive(t, received)))
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", msg.Header.Get("Reply-To"))
	assert.Equal(t, "shop@example.com", msg.Header.Get("From"))
}

func TestSMTPMailer_RefusesAnInvalidReplyTo(t *testing.T) {
	addr, received := smtpServer(t)

	err := mailer.NewSMTPMailer(addr).SendEmail(mailer.Email{
		To: "bob@example.com", From: "shop@example.com", ReplyTo: "alice@example.com\r\nBcc: eve@example.com", Title: "Hello", Body: "body",
	})

	assert.Error(t, err)
	assert.Empty(t, received, "nothing must be sent")
}

func TestSMTPMailer_SendTemplatedEmail(t *testing.T) {
	addr, received := smtpServer(t)

	err := mailer.NewSMTPMailer(addr).SendTemplatedEmail("bob@example.com", "shop@example.com", "ad-published", map[string]any{"title": "Bike"})

	require.NoError(t, err)
	msg := receive(t, received)
	assert.Contains(t, msg, "Subject: ad-published")
	assert.Contains(t, msg, `"title": "Bike"`)
}

func TestSMTPMailer_FailsWhenServerIsUnreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	ln.Close()

	err = mailer.NewSMTPMailer(addr).SendEmail(mailer.Email{To: "bob@example.com", From: "shop@example.com", Title: "Hello", Body: "body"})

	assert.Error(t, err)
}

func TestSMTPMailer_EncodesANonASCIISubject(t *testing.T) {
	addr, received := smtpServer(t)

	err := mailer.NewSMTPMailer(addr).SendEmail(mailer.Email{To: "bob@example.com", From: "shop@example.com", Title: "Votre annonce est publiée", Body: "body"})

	require.NoError(t, err)
	msg, err := mail.ReadMessage(strings.NewReader(receive(t, received)))
	require.NoError(t, err)
	raw := msg.Header.Get("Subject")
	assert.NotContains(t, raw, "é", "a header must stay ASCII")
	subject, err := new(mime.WordDecoder).DecodeHeader(raw)
	require.NoError(t, err)
	assert.Equal(t, "Votre annonce est publiée", subject)
}

func TestSMTPMailer_RefusesALineBreakInAHeader(t *testing.T) {
	cases := map[string][3]string{
		"subject": {"bob@example.com", "shop@example.com", "Hello\r\nBcc: eve@example.com"},
		"to":      {"bob@example.com\r\nBcc: eve@example.com", "shop@example.com", "Hello"},
		"from":    {"bob@example.com", "shop@example.com\nBcc: eve@example.com", "Hello"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			addr, received := smtpServer(t)

			err := mailer.NewSMTPMailer(addr).SendEmail(mailer.Email{To: c[0], From: c[1], Title: c[2], Body: "body"})

			assert.Error(t, err)
			assert.Empty(t, received, "nothing must be sent")
		})
	}
}

func TestSMTPMailer_RefusesAnInvalidAddress(t *testing.T) {
	addr, received := smtpServer(t)

	err := mailer.NewSMTPMailer(addr).SendEmail(mailer.Email{To: "not an address", From: "shop@example.com", Title: "Hello", Body: "body"})

	assert.Error(t, err)
	assert.Empty(t, received, "nothing must be sent")
}
