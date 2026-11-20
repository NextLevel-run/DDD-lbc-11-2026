package command_test

import (
	"strings"
	"testing"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/application/command"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	mailerspy "ddd-second-hand-marketplace/pkg/mailer/testing"
)

func depotePersiste(t *testing.T) (*inmemory.InMemoryAnnonceRepository, *domain.Annonce) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	annonce := annonceDeposee(t)
	if err := repo.Save(annonce); err != nil {
		t.Fatal(err)
	}
	return repo, annonce
}

func statutPersiste(t *testing.T, repo domain.AnnonceRepository, id string) domain.Statut {
	t.Helper()
	annonce, err := repo.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	return annonce.Statut()
}

func TestActiverAnnoncePutsTheAnnonceOnline(t *testing.T) {
	repo, annonce := depotePersiste(t)

	activee, err := command.BuildActiverAnnonceCommand(repo)(annonce.ID())

	if err != nil {
		t.Fatal(err)
	}
	if activee.ID() != annonce.ID() || statutPersiste(t, repo, annonce.ID()) != domain.Activee {
		t.Errorf("statut = %s", statutPersiste(t, repo, annonce.ID()))
	}
}

func TestActiverAnnonceRefusesAnUnknownOrAlreadyDecidedAnnonce(t *testing.T) {
	repo, annonce := depotePersiste(t)
	activer := command.BuildActiverAnnonceCommand(repo)

	if _, err := activer("inconnue"); err != domain.ErrAnnonceIntrouvable {
		t.Errorf("unknown: err = %v", err)
	}
	if _, err := activer(annonce.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := activer(annonce.ID()); err != domain.ErrTransitionStatutInvalide {
		t.Errorf("twice: err = %v", err)
	}
}

func TestMarquerAnnonceChallengeeKeepsTheMotif(t *testing.T) {
	repo, annonce := depotePersiste(t)

	challengee, err := command.BuildMarquerAnnonceChallengeeCommand(repo)(annonce.ID(), "Photos manquantes")

	if err != nil {
		t.Fatal(err)
	}
	if challengee.Motif() != "Photos manquantes" || statutPersiste(t, repo, annonce.ID()) != domain.Challengee {
		t.Errorf("statut = %s, motif = %q", challengee.Statut(), challengee.Motif())
	}
	if _, err := command.BuildMarquerAnnonceChallengeeCommand(repo)(annonce.ID(), "Encore"); err != domain.ErrTransitionStatutInvalide {
		t.Errorf("twice: err = %v", err)
	}
}

func TestMarquerAnnonceRefusee(t *testing.T) {
	repo, annonce := depotePersiste(t)
	marquer := command.BuildMarquerAnnonceRefuseeCommand(repo)

	if err := marquer(annonce.ID()); err != nil {
		t.Fatal(err)
	}
	if statutPersiste(t, repo, annonce.ID()) != domain.Refusee {
		t.Errorf("statut = %s", statutPersiste(t, repo, annonce.ID()))
	}
	if err := marquer("inconnue"); err != domain.ErrAnnonceIntrouvable {
		t.Errorf("unknown: err = %v", err)
	}
	if err := marquer(annonce.ID()); err != domain.ErrTransitionStatutInvalide {
		t.Errorf("twice: err = %v", err)
	}
}

func TestEnvoyerAnnonceValidee(t *testing.T) {
	spy := mailerspy.NewMailerSpy()
	annonce := annonceDeposee(t)

	if err := command.BuildEnvoyerAnnonceValideeCommand(spy, "noreply@example.com", "http://front.test/")(annonce); err != nil {
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
	if m.Title != "Votre annonce « Vélo de ville » est validée" {
		t.Errorf("title = %q", m.Title)
	}
	for _, want := range []string{"marie", "Vélo de ville", "http://front.test/annonces/" + annonce.ID()} {
		if !strings.Contains(m.Body, want) {
			t.Errorf("body misses %q:\n%s", want, m.Body)
		}
	}
	if strings.Contains(m.Body, "secret-1234") {
		t.Error("the email leaks the mot de passe")
	}
}

func TestEnvoyerAnnonceChallengeeCarriesTheMotifAndTheCorrectionLink(t *testing.T) {
	spy := mailerspy.NewMailerSpy()
	annonce := annonceDeposee(t)
	if err := annonce.Challenger("Photos manquantes"); err != nil {
		t.Fatal(err)
	}

	if err := command.BuildEnvoyerAnnonceChallengeeCommand(spy, "noreply@example.com", "http://front.test")(annonce); err != nil {
		t.Fatal(err)
	}

	sent := spy.GetSentSimpleEmails()
	if len(sent) != 1 {
		t.Fatalf("want 1 email, got %d", len(sent))
	}
	m := sent[0]
	if m.To != "marie@example.com" {
		t.Errorf("to = %s", m.To)
	}
	for _, want := range []string{"Photos manquantes", "http://front.test/annonces/" + annonce.ID() + "/corriger", "mot de passe"} {
		if !strings.Contains(m.Body, want) {
			t.Errorf("body misses %q:\n%s", want, m.Body)
		}
	}
	if strings.Contains(m.Body+m.Title, "secret-1234") {
		t.Error("the email leaks the mot de passe")
	}
}

func TestEnvoyerDecisionEmailsReturnTheSendFailure(t *testing.T) {
	annonce := annonceDeposee(t)
	if err := command.BuildEnvoyerAnnonceValideeCommand(failingMailer{}, "a@b.c", "http://front.test")(annonce); err == nil {
		t.Error("validee: want the send error")
	}
	if err := command.BuildEnvoyerAnnonceChallengeeCommand(failingMailer{}, "a@b.c", "http://front.test")(annonce); err == nil {
		t.Error("challengee: want the send error")
	}
}
