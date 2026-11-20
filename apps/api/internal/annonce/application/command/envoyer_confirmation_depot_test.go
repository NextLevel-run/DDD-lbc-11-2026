package command_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/annonce/application/command"
	"ddd-second-hand-marketplace/internal/annonce/domain"
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

	sent := spy.GetSentSimpleEmails()
	if len(sent) != 1 {
		t.Fatalf("want 1 email, got %d", len(sent))
	}
	m := sent[0]
	if m.To != "marie@example.com" || m.From != "noreply@example.com" {
		t.Errorf("to/from = %s/%s", m.To, m.From)
	}
	if m.Title != "Votre annonce « Vélo de ville » est en cours de modération" {
		t.Errorf("title = %q", m.Title)
	}
	if strings.Contains(m.Body, "est en ligne") {
		t.Errorf("the email promises the annonce is online:\n%s", m.Body)
	}
	for _, want := range []string{
		"marie", "Vélo de ville", "en cours de modération",
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

func (failingMailer) SendSimpleEmail(string, string, string, string) error {
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
