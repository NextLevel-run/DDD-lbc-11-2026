package inmemory

import (
	"sort"
	"sync"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
)

// InMemoryAnnonceRepository keeps the annonces in a map, safe for concurrent use.
type InMemoryAnnonceRepository struct {
	mu       sync.RWMutex
	annonces map[string]*domain.Annonce
}

var _ domain.AnnonceRepository = (*InMemoryAnnonceRepository)(nil)

func NewInMemoryAnnonceRepository() *InMemoryAnnonceRepository {
	return &InMemoryAnnonceRepository{annonces: make(map[string]*domain.Annonce)}
}

func (r *InMemoryAnnonceRepository) Save(annonce *domain.Annonce) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.annonces[annonce.ID()] = annonce
	return nil
}

func (r *InMemoryAnnonceRepository) GetByID(id string) (*domain.Annonce, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	annonce, ok := r.annonces[id]
	if !ok {
		return nil, domain.ErrAnnonceIntrouvable
	}
	return annonce, nil
}

// Rechercher scans every annonce: fine for the demo volumes an in-memory
// repository holds. Equal dates are ordered by ID so pages stay stable.
func (r *InMemoryAnnonceRepository) Rechercher(criteres domain.CriteresRecherche, page domain.Page) (domain.ResultatRecherche, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	retenues := []*domain.Annonce{}
	for _, annonce := range r.annonces {
		if criteres.Retient(annonce) {
			retenues = append(retenues, annonce)
		}
	}
	sort.Slice(retenues, func(i, j int) bool {
		a, b := retenues[i], retenues[j]
		if !a.DeposeeLe().Equal(b.DeposeeLe()) {
			return a.DeposeeLe().After(b.DeposeeLe())
		}
		return a.ID() < b.ID()
	})
	start := min(page.Offset(), len(retenues))
	end := min(start+domain.ParPage, len(retenues))
	return domain.ResultatRecherche{
		Annonces: append([]*domain.Annonce{}, retenues[start:end]...),
		Total:    len(retenues),
	}, nil
}
