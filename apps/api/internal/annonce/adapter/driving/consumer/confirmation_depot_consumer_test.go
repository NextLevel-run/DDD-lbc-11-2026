package consumer_test

import (
	"errors"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/annonce/adapter/driving/consumer"
	"ddd-second-hand-marketplace/internal/annonce/application/command"
	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	mailerspy "ddd-second-hand-marketplace/pkg/mailer/testing"
)

func newEvent(t *testing.T) *domain.AnnonceDeposeeEvent {
	t.Helper()
	a, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: "Vélo de ville", Description: "Un vélo en très bon état.", Prix: 100,
		Categorie: "vehicules", Ville: "Paris", CodePostal: "75011",
		Pseudo: "marie", Email: "marie@example.com", MotDePasse: "secret-1234",
		DeposeeLe: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return domain.NewAnnonceDeposeeEvent(time.Now(), a)
}

func TestConsumerSendsConfirmationOnAnnonceDeposee(t *testing.T) {
	bus := eventbus.NewSyncInMemoryEventBus()
	spy := mailerspy.NewMailerSpy()
	envoyer := command.BuildEnvoyerConfirmationDepotCommand(spy, "noreply@example.com", "http://front.test")
	if err := consumer.SubscribeConfirmationDepot(bus, envoyer); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(newEvent(t)); err != nil {
		t.Fatal(err)
	}
	if n := len(spy.GetSentEmails()); n != 1 {
		t.Fatalf("want 1 email, got %d", n)
	}
}

func TestConsumerSwallowsSendFailure(t *testing.T) {
	bus := eventbus.NewSyncInMemoryEventBus()
	failing := func(*domain.Annonce) error { return errors.New("smtp down") }
	if err := consumer.SubscribeConfirmationDepot(bus, failing); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(newEvent(t)); err != nil {
		t.Fatalf("a send failure must not fail the publish: %v", err)
	}
}
