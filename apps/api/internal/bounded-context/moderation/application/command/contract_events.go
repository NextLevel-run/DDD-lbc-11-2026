package command

import (
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"
	"ddd-second-hand-marketplace/internal/contracts"
)

// The decisions are stored in the stream as domain events (the storage format)
// and published as contracts: the stream may change without breaking annonce.

func annonceValideeEvent(e domain.AnnonceValideeEvent) contracts.AnnonceValideeEvent {
	return contracts.AnnonceValideeEvent{AnnonceID: e.AnnonceID(), Date: e.OccurredAt()}
}

func annonceRefuseeEvent(e domain.AnnonceRefuseeEvent) contracts.AnnonceRefuseeEvent {
	return contracts.AnnonceRefuseeEvent{AnnonceID: e.AnnonceID(), Date: e.OccurredAt()}
}

func annonceChallengeeEvent(e domain.AnnonceChallengeeEvent) contracts.AnnonceChallengeeEvent {
	return contracts.AnnonceChallengeeEvent{AnnonceID: e.AnnonceID(), Motif: e.Motif(), Date: e.OccurredAt()}
}
