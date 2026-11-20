package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnnonceSupprimeeEvent is emitted once a vendeur has withdrawn his annonce
// and the change is persisted. It carries the raison: the annonce does not.
// No other context reads it yet: the events that cross a context boundary live
// in internal/contracts.
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
