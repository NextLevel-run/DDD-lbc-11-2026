// Package contracts is the published language shared by the bounded contexts:
// the events that cross a context boundary, as plain data. A context publishes
// and reads them here, so neither imports the other and no event type or field
// is written twice. It holds no behavior and imports only the standard library.
package contracts

import "time"

// Types of the events published by the annonce context.
const (
	AnnonceDeposeeEventType  = "AnnonceDeposee"
	AnnonceCorrigeeEventType = "AnnonceCorrigee"
)

// InstantaneAnnonce is the annonce as a reader outside its context may know it,
// at the moment of the event. It never carries the vendeur's email, telephone
// or mot de passe.
type InstantaneAnnonce struct {
	AnnonceID   string
	Titre       string
	Description string
	Prix        int64  // in centimes
	Categorie   string // code of the catégorie
	Ville       string
	CodePostal  string
	Pseudo      string
	DeposeeLe   time.Time
}

// AnnonceDeposeeEvent: a vendeur deposited an annonce, now persisted.
type AnnonceDeposeeEvent struct {
	Annonce InstantaneAnnonce
	Date    time.Time
}

// EventType implements eventbus.DomainEvent.
func (AnnonceDeposeeEvent) EventType() string       { return AnnonceDeposeeEventType }
func (e AnnonceDeposeeEvent) OccurredAt() time.Time { return e.Date }

// AnnonceCorrigeeEvent: a vendeur corrected a challenged annonce, now
// persisted. It carries the same snapshot as the dépôt.
type AnnonceCorrigeeEvent struct {
	Annonce InstantaneAnnonce
	Date    time.Time
}

// EventType implements eventbus.DomainEvent.
func (AnnonceCorrigeeEvent) EventType() string       { return AnnonceCorrigeeEventType }
func (e AnnonceCorrigeeEvent) OccurredAt() time.Time { return e.Date }
