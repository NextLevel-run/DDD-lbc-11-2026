// Package consumer holds the driving adapters that react to events of other
// bounded contexts.
package consumer

import (
	"log"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/application/command"
	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// SubscribeMiseEnFile puts each deposited annonce, and each corrected one, in
// the moderation queue. A failure is logged: the bus is fire-and-forget.
func SubscribeMiseEnFile(bus eventbus.Bus, mettreEnFile command.MettreEnFileCommand) error {
	handle := func(a contracts.InstantaneAnnonce) error {
		err := mettreEnFile(command.MettreEnFileCommandArgs{
			AnnonceID: a.AnnonceID, Titre: a.Titre, Description: a.Description, Prix: a.Prix,
			Categorie: a.Categorie, Ville: a.Ville, CodePostal: a.CodePostal, Pseudo: a.Pseudo,
			DeposeeLe: a.DeposeeLe,
		})
		if err != nil {
			log.Printf("mise en file: annonce %s: %v", a.AnnonceID, err)
		}
		return err
	}
	err := bus.Subscribe(contracts.AnnonceDeposeeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(contracts.AnnonceDeposeeEvent)
		if !ok {
			return nil
		}
		return handle(e.Annonce)
	})
	if err != nil {
		return err
	}
	return bus.Subscribe(contracts.AnnonceCorrigeeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(contracts.AnnonceCorrigeeEvent)
		if !ok {
			return nil
		}
		return handle(e.Annonce)
	})
}
