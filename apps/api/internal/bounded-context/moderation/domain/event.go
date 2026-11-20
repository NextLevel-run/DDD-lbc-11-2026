package domain

import "time"

// Evenement is a fact of an annonce's stream. The stream is the source of
// truth: the state is whatever replaying it gives.
type Evenement interface {
	EventType() string
	OccurredAt() time.Time
}

// Stable types of the stream's events.
const (
	AnnonceMiseEnFileEventType    = "AnnonceMiseEnFile"
	AnnonceVerrouilleeEventType   = "AnnonceVerrouillee"
	VerrouAnnonceRelacheEventType = "VerrouAnnonceRelache"
	AnnonceValideeEventType       = "AnnonceValidee"
	AnnonceRefuseeEventType       = "AnnonceRefusee"
	AnnonceChallengeeEventType    = "AnnonceChallengee"
)

// Causes of a VerrouAnnonceRelache.
const (
	CauseManuel     = "manuel"
	CauseExpiration = "expiration"
)

// InstantaneAnnonce is the annonce as it was when it entered the queue: what
// a moderator reads. It is made of primitives, so moderation never imports
// the annonce context.
type InstantaneAnnonce struct {
	AnnonceID   string
	Titre       string
	Description string
	Prix        int64 // in centimes
	Categorie   string
	Ville       string
	CodePostal  string
	Pseudo      string
	DeposeeLe   time.Time
}

// AnnonceMiseEnFileEvent: the annonce entered the moderation queue.
type AnnonceMiseEnFileEvent struct {
	Annonce InstantaneAnnonce
	Date    time.Time
}

func (e AnnonceMiseEnFileEvent) EventType() string     { return AnnonceMiseEnFileEventType }
func (e AnnonceMiseEnFileEvent) OccurredAt() time.Time { return e.Date }

// AnnonceVerrouilleeEvent: a moderator locked the annonce to treat it.
type AnnonceVerrouilleeEvent struct {
	Moderateur string
	Date       time.Time
}

func (e AnnonceVerrouilleeEvent) EventType() string     { return AnnonceVerrouilleeEventType }
func (e AnnonceVerrouilleeEvent) OccurredAt() time.Time { return e.Date }

// VerrouAnnonceRelacheEvent: the lock is gone, by its holder (Cause manuel)
// or because it is older than 2 hours (Cause expiration, no Moderateur).
type VerrouAnnonceRelacheEvent struct {
	Moderateur string
	Cause      string
	Date       time.Time
}

func (e VerrouAnnonceRelacheEvent) EventType() string     { return VerrouAnnonceRelacheEventType }
func (e VerrouAnnonceRelacheEvent) OccurredAt() time.Time { return e.Date }

// The three decision events are stored in the stream: they are the storage
// format. The application layer publishes them as the matching contracts
// (internal/contracts) once stored, so annonce never imports this package.

// AnnonceValideeEvent: a moderator validated the annonce.
type AnnonceValideeEvent struct {
	annonceID  string
	Moderateur string
	Date       time.Time
}

func NewAnnonceValideeEvent(annonceID, moderateur string, date time.Time) AnnonceValideeEvent {
	return AnnonceValideeEvent{annonceID: annonceID, Moderateur: moderateur, Date: date}
}

func (e AnnonceValideeEvent) EventType() string     { return AnnonceValideeEventType }
func (e AnnonceValideeEvent) OccurredAt() time.Time { return e.Date }
func (e AnnonceValideeEvent) AnnonceID() string     { return e.annonceID }

// AnnonceRefuseeEvent: a moderator refused the annonce (fraud).
type AnnonceRefuseeEvent struct {
	annonceID  string
	Moderateur string
	Date       time.Time
}

func NewAnnonceRefuseeEvent(annonceID, moderateur string, date time.Time) AnnonceRefuseeEvent {
	return AnnonceRefuseeEvent{annonceID: annonceID, Moderateur: moderateur, Date: date}
}

func (e AnnonceRefuseeEvent) EventType() string     { return AnnonceRefuseeEventType }
func (e AnnonceRefuseeEvent) OccurredAt() time.Time { return e.Date }
func (e AnnonceRefuseeEvent) AnnonceID() string     { return e.annonceID }

// AnnonceChallengeeEvent: a moderator sent the annonce back to its vendeur,
// with the motif of the correction asked.
type AnnonceChallengeeEvent struct {
	annonceID  string
	motif      string
	Moderateur string
	Date       time.Time
}

func NewAnnonceChallengeeEvent(annonceID, moderateur, motif string, date time.Time) AnnonceChallengeeEvent {
	return AnnonceChallengeeEvent{annonceID: annonceID, motif: motif, Moderateur: moderateur, Date: date}
}

func (e AnnonceChallengeeEvent) EventType() string     { return AnnonceChallengeeEventType }
func (e AnnonceChallengeeEvent) OccurredAt() time.Time { return e.Date }
func (e AnnonceChallengeeEvent) AnnonceID() string     { return e.annonceID }
func (e AnnonceChallengeeEvent) Motif() string         { return e.motif }
