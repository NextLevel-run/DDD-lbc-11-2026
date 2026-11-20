package command_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/annonce/application/command"
	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/mailer"
	mailerspy "ddd-second-hand-marketplace/pkg/mailer/testing"
)

func annonceDeposee(t *testing.T) *domain.Annonce {
	t.Helper()
	a, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: "Vélo de ville", Description: "Un vélo en très bon état.", Prix: 15000,
		Categorie: "vehicules", Ville: "Paris", CodePostal: "75011",
		Pseudo: "marie", Email: "marie@example.com", MotDePasse: "secret-1234",
		DeposeeLe: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestEnvoyerConfirmationDepot(t *testing.T) {
	spy := mailerspy.NewMailerSpy()
	annonce := annonceDeposee(t)
	cmd := command.BuildEnvoyerConfirmationDepotCommand(spy, "noreply@example.com", "http://front.test")

	if err := cmd(annonce); err != nil {
		t.Fatal(err)
	}

	sent := spy.GetSentEmails()
	if len(sent) != 1 {
		t.Fatalf("want 1 email, got %d", len(sent))
	}
	m := sent[0]
	if m.To != "marie@example.com" || m.From != "noreply@example.com" || m.ReplyTo != "" {
		t.Errorf("to/from/reply-to = %s/%s/%q", m.To, m.From, m.ReplyTo)
	}
	if m.Title != "Votre annonce « Vélo de ville » est en ligne" {
		t.Errorf("title = %q", m.Title)
	}
	for _, want := range []string{
		"marie", "Vélo de ville",
		"http://front.test/annonces/" + annonce.ID(),
		"http://front.test/annonces/" + annonce.ID() + "/supprimer",
		"mot de passe",
	} {
		if !strings.Contains(m.Body, want) {
			t.Errorf("body misses %q:\n%s", want, m.Body)
		}
	}
	if strings.Contains(m.Body, "secret-1234") || strings.Contains(m.Title, "secret-1234") {
		t.Error("the email leaks the mot de passe")
	}
}

type failingMailer struct{}

func (failingMailer) SendEmail(mailer.Email) error {
	return errors.New("smtp down")
}
func (failingMailer) SendTemplatedEmail(string, string, string, map[string]any) error {
	return errors.New("smtp down")
}

func TestEnvoyerConfirmationDepotReturnsSendFailure(t *testing.T) {
	cmd := command.BuildEnvoyerConfirmationDepotCommand(failingMailer{}, "noreply@example.com", "http://front.test")
	if err := cmd(annonceDeposee(t)); err == nil {
		t.Fatal("want the send error")
	}
}
