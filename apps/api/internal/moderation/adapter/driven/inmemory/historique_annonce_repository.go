// Package inmemory holds the in-memory driven adapters of moderation.
package inmemory

import (
	"sort"
	"sync"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// InMemoryHistoriqueAnnonceRepository is an event store: one journal of events
// per annonce, appended under an optimistic version check.
type InMemoryHistoriqueAnnonceRepository struct {
	mu       sync.RWMutex
	journaux map[string][]domain.Evenement
}

var _ domain.HistoriqueAnnonceRepository = (*InMemoryHistoriqueAnnonceRepository)(nil)

func NewInMemoryHistoriqueAnnonceRepository() *InMemoryHistoriqueAnnonceRepository {
	return &InMemoryHistoriqueAnnonceRepository{journaux: make(map[string][]domain.Evenement)}
}

func (r *InMemoryHistoriqueAnnonceRepository) Charger(annonceID string) (*domain.HistoriqueAnnonce, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return domain.RestoreHistoriqueAnnonce(annonceID, r.journaux[annonceID]), nil
}

func (r *InMemoryHistoriqueAnnonceRepository) Enregistrer(h *domain.HistoriqueAnnonce) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	journal := r.journaux[h.AnnonceID()]
	if len(journal) != h.Version() {
		return domain.ErrConflitVersion
	}
	// A new slice, so a historique loaded earlier never sees the append.
	r.journaux[h.AnnonceID()] = append(append([]domain.Evenement(nil), journal...), h.NouveauxEvenements()...)
	return nil
}

// Tous loads every stream, ordered by annonce ID so reads are stable.
func (r *InMemoryHistoriqueAnnonceRepository) Tous() ([]*domain.HistoriqueAnnonce, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.journaux))
	for id := range r.journaux {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	historiques := make([]*domain.HistoriqueAnnonce, len(ids))
	for i, id := range ids {
		historiques[i] = domain.RestoreHistoriqueAnnonce(id, r.journaux[id])
	}
	return historiques, nil
}
