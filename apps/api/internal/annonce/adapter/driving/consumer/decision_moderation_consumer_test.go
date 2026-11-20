package consumer_test

import (
	"errors"
	"testing"

	"ddd-second-hand-marketplace/internal/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/annonce/adapter/driving/consumer"
	"ddd-second-hand-marketplace/internal/annonce/application/command"
	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	mailerspy "ddd-second-hand-marketplace/pkg/mailer/testing"
)

// decisionEvent stands for an event of moderation: this package knows only the
// accessors, never the type.
type decisionEvent struct {
	eventType string
	id        string
	motif     string
}

func (e decisionEvent) EventType() string { return e.eventType }
func (e decisionEvent) AnnonceID() string { return e.id }
func (e decisionEvent) Motif() string     { return e.motif }

type setup struct {
	bus  eventbus.Bus
	repo *inmemory.InMemoryAnnonceRepository
	spy  *mailerspy.MailerSpy
	id   string
}

func newSetup(t *testing.T) setup {
	t.Helper()
	s := setup{bus: eventbus.NewSyncInMemoryEventBus(), repo: inmemory.NewInMemoryAnnonceRepository(), spy: mailerspy.NewMailerSpy()}
	annonce := newEvent(t).Annonce
	if err := s.repo.Save(annonce); err != nil {
		t.Fatal(err)
	}
	s.id = annonce.ID()
	if err := consumer.SubscribeAnnonceValidee(s.bus, command.BuildActiverAnnonceCommand(s.repo),
		command.BuildEnvoyerAnnonceValideeCommand(s.spy, "noreply@example.com", "http://front.test")); err != nil {
		t.Fatal(err)
	}
	if err := consumer.SubscribeAnnonceChallengee(s.bus, command.BuildMarquerAnnonceChallengeeCommand(s.repo),
		command.BuildEnvoyerAnnonceChallengeeCommand(s.spy, "noreply@example.com", "http://front.test")); err != nil {
		t.Fatal(err)
	}
	if err := consumer.SubscribeAnnonceRefusee(s.bus, command.BuildMarquerAnnonceRefuseeCommand(s.repo)); err != nil {
		t.Fatal(err)
	}
	return s
}

func (s setup) statut(t *testing.T) domain.Statut {
	t.Helper()
	annonce, err := s.repo.GetByID(s.id)
	if err != nil {
		t.Fatal(err)
	}
	return annonce.Statut()
}

func TestAnnonceValideeActivatesTheAnnonceAndTellsTheVendeur(t *testing.T) {
	s := newSetup(t)

	if err := s.bus.Publish(decisionEvent{eventType: "AnnonceValidee", id: s.id}); err != nil {
		t.Fatal(err)
	}

	if s.statut(t) != domain.Activee {
		t.Errorf("statut = %s", s.statut(t))
	}
	if n := len(s.spy.GetSentSimpleEmails()); n != 1 {
		t.Fatalf("want 1 email, got %d", n)
	}
}

func TestAnnonceChallengeeKeepsTheMotifAndTellsTheVendeur(t *testing.T) {
	s := newSetup(t)

	if err := s.bus.Publish(decisionEvent{eventType: "AnnonceChallengee", id: s.id, motif: "Photos manquantes"}); err != nil {
		t.Fatal(err)
	}

	annonce, _ := s.repo.GetByID(s.id)
	if annonce.Statut() != domain.Challengee || annonce.Motif() != "Photos manquantes" {
		t.Errorf("statut = %s, motif = %q", annonce.Statut(), annonce.Motif())
	}
	sent := s.spy.GetSentSimpleEmails()
	if len(sent) != 1 {
		t.Fatalf("want 1 email, got %d", len(sent))
	}
}

func TestAnnonceRefuseeMarksTheAnnonceAndSendsNothing(t *testing.T) {
	s := newSetup(t)

	if err := s.bus.Publish(decisionEvent{eventType: "AnnonceRefusee", id: s.id}); err != nil {
		t.Fatal(err)
	}

	if s.statut(t) != domain.Refusee {
		t.Errorf("statut = %s", s.statut(t))
	}
	if n := len(s.spy.GetSentSimpleEmails()); n != 0 {
		t.Errorf("a refusal sends no email, got %d", n)
	}
}

func TestDecisionsSwallowFailuresAndNeverFailThePublish(t *testing.T) {
	s := newSetup(t)
	failing := func(*domain.Annonce) error { return errors.New("smtp down") }
	bus := eventbus.NewSyncInMemoryEventBus()
	if err := consumer.SubscribeAnnonceValidee(bus, command.BuildActiverAnnonceCommand(s.repo), failing); err != nil {
		t.Fatal(err)
	}
	if err := consumer.SubscribeAnnonceChallengee(bus, command.BuildMarquerAnnonceChallengeeCommand(s.repo), failing); err != nil {
		t.Fatal(err)
	}

	if err := bus.Publish(decisionEvent{eventType: "AnnonceValidee", id: s.id}); err != nil {
		t.Fatalf("a send failure must not fail the publish: %v", err)
	}
	if s.statut(t) != domain.Activee {
		t.Errorf("the annonce stays activee, got %s", s.statut(t))
	}
	if err := bus.Publish(decisionEvent{eventType: "AnnonceValidee", id: "inconnue"}); err != nil {
		t.Errorf("an unknown annonce must not fail the publish: %v", err)
	}
	if err := bus.Publish(decisionEvent{eventType: "AnnonceChallengee", id: s.id, motif: "x"}); err != nil {
		t.Errorf("an invalid transition must not fail the publish: %v", err)
	}
}
