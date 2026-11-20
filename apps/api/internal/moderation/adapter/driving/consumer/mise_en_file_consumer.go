// Package consumer holds the driving adapters that react to events of other
// bounded contexts.
package consumer

import (
	"log"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// annonceDeposeeEvent is what moderation needs from an AnnonceDeposee: a local
// interface, so the annonce context is never imported.
type annonceDeposeeEvent interface {
	AnnonceID() string
	Titre() string
	Description() string
	Prix() int64
	Categorie() string
	Ville() string
	CodePostal() string
	Pseudo() string
	DeposeeLe() time.Time
}

// The stable types of the annonce context's events: AnnonceCorrigee carries
// the same accessors as AnnonceDeposee.
const (
	annonceDeposeeEventType  = "AnnonceDeposee"
	annonceCorrigeeEventType = "AnnonceCorrigee"
)

// SubscribeMiseEnFile puts each deposited annonce, and each corrected one, in
// the moderation queue. A failure is logged: the bus is fire-and-forget.
func SubscribeMiseEnFile(bus eventbus.Bus, mettreEnFile command.MettreEnFileCommand) error {
	handler := func(event eventbus.DomainEvent) error {
		e, ok := event.(annonceDeposeeEvent)
		if !ok {
			return nil
		}
		err := mettreEnFile(command.MettreEnFileCommandArgs{
			AnnonceID: e.AnnonceID(), Titre: e.Titre(), Description: e.Description(), Prix: e.Prix(),
			Categorie: e.Categorie(), Ville: e.Ville(), CodePostal: e.CodePostal(), Pseudo: e.Pseudo(),
			DeposeeLe: e.DeposeeLe(),
		})
		if err != nil {
			log.Printf("mise en file: annonce %s: %v", e.AnnonceID(), err)
		}
		return err
	}
	if err := bus.Subscribe(annonceDeposeeEventType, handler); err != nil {
		return err
	}
	return bus.Subscribe(annonceCorrigeeEventType, handler)
}
