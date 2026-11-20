// Package consumer holds the driving adapters that react to domain events.
package consumer

import (
	"log"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/application/command"
	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// SubscribeConfirmationDepot sends the confirmation de dépôt on each
// AnnonceDeposee. A send failure is logged, never retried, and never fails the
// deposit.
func SubscribeConfirmationDepot(bus eventbus.Bus, envoyer command.EnvoyerConfirmationDepotCommand) error {
	return bus.Subscribe(contracts.AnnonceDeposeeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(contracts.AnnonceDeposeeEvent)
		if !ok {
			return nil
		}
		if err := envoyer(e.Annonce.AnnonceID); err != nil {
			log.Printf("confirmation de depot: annonce %s: %v", e.Annonce.AnnonceID, err)
		}
		return nil
	})
}
