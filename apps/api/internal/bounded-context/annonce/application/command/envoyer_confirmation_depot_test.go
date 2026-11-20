package command_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/application/command"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	mailerspy "ddd-second-hand-marketplace/pkg/mailer/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	repo := inmemory.NewInMemoryAnnonceRepository()
	annonce := annonceDeposee(t)
	require.NoError(t, repo.Save(annonce))
	cmd := command.BuildEnvoyerConfirmationDepotCommand(repo, spy, "noreply@example.com", "http://front.test")

	if err := cmd(annonce.ID()); err != nil {
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
	repo := inmemory.NewInMemoryAnnonceRepository()
	annonce := annonceDeposee(t)
	require.NoError(t, repo.Save(annonce))
	cmd := command.BuildEnvoyerConfirmationDepotCommand(repo, failingMailer{}, "noreply@example.com", "http://front.test")
	if err := cmd(annonce.ID()); err == nil {
		t.Fatal("want the send error")
	}
}

func TestEnvoyerConfirmationDepotFailsOnAnUnknownAnnonce(t *testing.T) {
	spy := mailerspy.NewMailerSpy()
	cmd := command.BuildEnvoyerConfirmationDepotCommand(inmemory.NewInMemoryAnnonceRepository(), spy, "noreply@example.com", "http://front.test")

	err := cmd("7d1cbd8e-2a2c-4f1c-a8f2-6c1f0a3f4b11")

	assert.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
	assert.Empty(t, spy.GetSentSimpleEmails())
}
