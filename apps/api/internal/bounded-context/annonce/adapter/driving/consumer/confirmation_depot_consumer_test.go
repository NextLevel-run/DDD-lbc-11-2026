package consumer_test

import (
	"errors"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driving/consumer"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/application/command"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"
	mailerspy "ddd-second-hand-marketplace/pkg/mailer/testing"
)

func newAnnonce(t *testing.T) *domain.Annonce {
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
	return a
}

func deposee(annonce *domain.Annonce) contracts.AnnonceDeposeeEvent {
	return contracts.AnnonceDeposeeEvent{
		Annonce: contracts.InstantaneAnnonce{AnnonceID: annonce.ID(), Titre: annonce.Titre()},
		Date:    time.Now(),
	}
}

func TestConsumerSendsConfirmationOnAnnonceDeposee(t *testing.T) {
	bus := eventbus.NewSyncInMemoryEventBus()
	spy := mailerspy.NewMailerSpy()
	repo := inmemory.NewInMemoryAnnonceRepository()
	annonce := newAnnonce(t)
	if err := repo.Save(annonce); err != nil {
		t.Fatal(err)
	}
	envoyer := command.BuildEnvoyerConfirmationDepotCommand(repo, spy, "noreply@example.com", "http://front.test")
	if err := consumer.SubscribeConfirmationDepot(bus, envoyer); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(deposee(annonce)); err != nil {
		t.Fatal(err)
	}
	sent := spy.GetSentSimpleEmails()
	if len(sent) != 1 {
		t.Fatalf("want 1 email, got %d", len(sent))
	}
	if sent[0].To != "marie@example.com" {
		t.Errorf("to = %s: the vendeur's email comes from the annonce, not the event", sent[0].To)
	}
}

func TestConsumerSwallowsSendFailure(t *testing.T) {
	bus := eventbus.NewSyncInMemoryEventBus()
	failing := func(string) error { return errors.New("smtp down") }
	if err := consumer.SubscribeConfirmationDepot(bus, failing); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(deposee(newAnnonce(t))); err != nil {
		t.Fatalf("a send failure must not fail the publish: %v", err)
	}
}
