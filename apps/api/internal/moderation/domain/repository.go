package domain

// HistoriqueAnnonceRepository is the event store of the annonces' streams.
type HistoriqueAnnonceRepository interface {
	// Charger replays the stream of an annonce. An unknown annonce has an
	// empty stream: an empty historique, not an error.
	Charger(annonceID string) (*HistoriqueAnnonce, error)
	// Enregistrer appends the new events of h. It answers ErrConflitVersion
	// when the stream grew since h was loaded.
	Enregistrer(h *HistoriqueAnnonce) error
	// Tous loads every stream, for the reads that compute the queue.
	Tous() ([]*HistoriqueAnnonce, error)
}
