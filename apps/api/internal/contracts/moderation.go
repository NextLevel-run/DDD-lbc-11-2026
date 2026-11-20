package contracts

import "time"

// Types of the decisions published by the moderation context.
const (
	AnnonceValideeEventType    = "AnnonceValidee"
	AnnonceRefuseeEventType    = "AnnonceRefusee"
	AnnonceChallengeeEventType = "AnnonceChallengee"
)

// AnnonceValideeEvent: a modérateur validated the annonce, which goes online.
type AnnonceValideeEvent struct {
	AnnonceID string
	Date      time.Time
}

// EventType implements eventbus.DomainEvent.
func (AnnonceValideeEvent) EventType() string       { return AnnonceValideeEventType }
func (e AnnonceValideeEvent) OccurredAt() time.Time { return e.Date }

// AnnonceRefuseeEvent: a modérateur refused the annonce for good.
type AnnonceRefuseeEvent struct {
	AnnonceID string
	Date      time.Time
}

// EventType implements eventbus.DomainEvent.
func (AnnonceRefuseeEvent) EventType() string       { return AnnonceRefuseeEventType }
func (e AnnonceRefuseeEvent) OccurredAt() time.Time { return e.Date }

// AnnonceChallengeeEvent: a modérateur sent the annonce back to its vendeur,
// with the motif of the correction asked.
type AnnonceChallengeeEvent struct {
	AnnonceID string
	Motif     string
	Date      time.Time
}

// EventType implements eventbus.DomainEvent.
func (AnnonceChallengeeEvent) EventType() string       { return AnnonceChallengeeEventType }
func (e AnnonceChallengeeEvent) OccurredAt() time.Time { return e.Date }
