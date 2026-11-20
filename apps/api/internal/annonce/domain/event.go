package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnnonceDeposeeEvent is emitted once a new annonce has been deposited and
// persisted. Moderation subscribes to it through the primitive accessors,
// never importing this type.
type AnnonceDeposeeEvent struct {
	id         string
	occurredAt time.Time
	Annonce    *Annonce
}

const AnnonceDeposeeEventType = "AnnonceDeposee"

func NewAnnonceDeposeeEvent(occurredAt time.Time, annonce *Annonce) *AnnonceDeposeeEvent {
	return &AnnonceDeposeeEvent{id: uuid.NewString(), occurredAt: occurredAt, Annonce: annonce}
}

// EventType implements eventbus.DomainEvent.
func (e *AnnonceDeposeeEvent) EventType() string { return AnnonceDeposeeEventType }
func (e *AnnonceDeposeeEvent) ID() string        { return e.id }
func (e *AnnonceDeposeeEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *AnnonceDeposeeEvent) AnnonceID() string    { return e.Annonce.ID() }
func (e *AnnonceDeposeeEvent) Titre() string        { return e.Annonce.Titre() }
func (e *AnnonceDeposeeEvent) Description() string  { return e.Annonce.Description() }
func (e *AnnonceDeposeeEvent) Prix() int64          { return e.Annonce.Prix().Centimes() }
func (e *AnnonceDeposeeEvent) Categorie() string    { return e.Annonce.Categorie().Code() }
func (e *AnnonceDeposeeEvent) Ville() string        { return e.Annonce.Localisation().Ville() }
func (e *AnnonceDeposeeEvent) CodePostal() string   { return e.Annonce.Localisation().CodePostal() }
func (e *AnnonceDeposeeEvent) Pseudo() string       { return e.Annonce.Vendeur().Pseudo() }
func (e *AnnonceDeposeeEvent) DeposeeLe() time.Time { return e.Annonce.DeposeeLe() }

// AnnonceCorrigeeEvent is emitted once a vendeur has corrected a challenged
// annonce and the change is persisted. It exposes the same primitive accessors
// as AnnonceDeposeeEvent, so moderation queues it back through one interface.
type AnnonceCorrigeeEvent struct {
	id         string
	occurredAt time.Time
	Annonce    *Annonce
}

const AnnonceCorrigeeEventType = "AnnonceCorrigee"

func NewAnnonceCorrigeeEvent(occurredAt time.Time, annonce *Annonce) *AnnonceCorrigeeEvent {
	return &AnnonceCorrigeeEvent{id: uuid.NewString(), occurredAt: occurredAt, Annonce: annonce}
}

// EventType implements eventbus.DomainEvent.
func (e *AnnonceCorrigeeEvent) EventType() string     { return AnnonceCorrigeeEventType }
func (e *AnnonceCorrigeeEvent) ID() string            { return e.id }
func (e *AnnonceCorrigeeEvent) OccurredAt() time.Time { return e.occurredAt }

func (e *AnnonceCorrigeeEvent) AnnonceID() string    { return e.Annonce.ID() }
func (e *AnnonceCorrigeeEvent) Titre() string        { return e.Annonce.Titre() }
func (e *AnnonceCorrigeeEvent) Description() string  { return e.Annonce.Description() }
func (e *AnnonceCorrigeeEvent) Prix() int64          { return e.Annonce.Prix().Centimes() }
func (e *AnnonceCorrigeeEvent) Categorie() string    { return e.Annonce.Categorie().Code() }
func (e *AnnonceCorrigeeEvent) Ville() string        { return e.Annonce.Localisation().Ville() }
func (e *AnnonceCorrigeeEvent) CodePostal() string   { return e.Annonce.Localisation().CodePostal() }
func (e *AnnonceCorrigeeEvent) Pseudo() string       { return e.Annonce.Vendeur().Pseudo() }
func (e *AnnonceCorrigeeEvent) DeposeeLe() time.Time { return e.Annonce.DeposeeLe() }

// AnnonceSupprimeeEvent is emitted once a vendeur has withdrawn his annonce
// and the change is persisted. It carries the raison: the annonce does not.
type AnnonceSupprimeeEvent struct {
	id          string
	occurredAt  time.Time
	AnnonceID   string
	Raison      RaisonSuppression
	SupprimeeLe time.Time
}

const AnnonceSupprimeeEventType = "AnnonceSupprimee"

func NewAnnonceSupprimeeEvent(occurredAt time.Time, annonceID string, raison RaisonSuppression) *AnnonceSupprimeeEvent {
	return &AnnonceSupprimeeEvent{id: uuid.NewString(), occurredAt: occurredAt, AnnonceID: annonceID, Raison: raison, SupprimeeLe: occurredAt}
}

// EventType implements eventbus.DomainEvent.
func (e *AnnonceSupprimeeEvent) EventType() string { return AnnonceSupprimeeEventType }
func (e *AnnonceSupprimeeEvent) ID() string        { return e.id }
func (e *AnnonceSupprimeeEvent) OccurredAt() time.Time {
	return e.occurredAt
}
