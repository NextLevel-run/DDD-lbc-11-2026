package query

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"
)

// EvenementView is a dated line of the history of an annonce.
type EvenementView struct {
	Type       string
	Date       time.Time
	Moderateur string // empty when no moderator is behind it (expiration)
	Motif      string
}

// AnnonceAModererView is what a moderator reads: the snapshot, the state and
// the lock, and the history.
type AnnonceAModererView struct {
	ID               string
	Titre            string
	Description      string
	Prix             int64 // in centimes
	Categorie        string
	Ville            string
	CodePostal       string
	Pseudo           string
	DeposeeLe        time.Time
	Statut           string // en-file, verrouillee, validee, refusee, challengee
	VerrouillePar    string
	VerrouilleDepuis time.Time
	Historique       []EvenementView
}

// ConsulterAnnonceAModererQuery returns the detail of an annonce of the queue.
type ConsulterAnnonceAModererQuery func(annonceID string) (*AnnonceAModererView, error)

// BuildConsulterAnnonceAModererQuery replays the stream of the annonce and
// answers ErrAnnonceIntrouvable when it is empty. An expired lock shows as
// free, the stream keeping its trace until the next lock.
func BuildConsulterAnnonceAModererQuery(repo domain.HistoriqueAnnonceRepository, now func() time.Time) ConsulterAnnonceAModererQuery {
	return func(annonceID string) (*AnnonceAModererView, error) {
		h, err := repo.Charger(annonceID)
		if err != nil {
			return nil, err
		}
		if h.Etat() == domain.EtatVide {
			return nil, domain.ErrAnnonceIntrouvable
		}
		annonce := h.Annonce()
		view := &AnnonceAModererView{
			ID: h.AnnonceID(), Titre: annonce.Titre, Description: annonce.Description, Prix: annonce.Prix,
			Categorie: annonce.Categorie, Ville: annonce.Ville, CodePostal: annonce.CodePostal,
			Pseudo: annonce.Pseudo, DeposeeLe: annonce.DeposeeLe,
			Statut:     string(h.Etat()),
			Historique: historique(h.Evenements()),
		}
		if moderateur, depuis, ok := h.Verrou(now()); ok {
			view.VerrouillePar, view.VerrouilleDepuis = moderateur, depuis
		} else if h.Etat() == domain.EtatVerrouillee {
			view.Statut = string(domain.EtatEnFile) // an expired lock counts as free
		}
		return view, nil
	}
}

func historique(evenements []domain.Evenement) []EvenementView {
	views := make([]EvenementView, len(evenements))
	for i, e := range evenements {
		view := EvenementView{Type: e.EventType(), Date: e.OccurredAt()}
		switch e := e.(type) {
		case domain.AnnonceVerrouilleeEvent:
			view.Moderateur = e.Moderateur
		case domain.VerrouAnnonceRelacheEvent:
			view.Moderateur = e.Moderateur
		case domain.AnnonceValideeEvent:
			view.Moderateur = e.Moderateur
		case domain.AnnonceRefuseeEvent:
			view.Moderateur = e.Moderateur
		case domain.AnnonceChallengeeEvent:
			view.Moderateur, view.Motif = e.Moderateur, e.Motif()
		}
		views[i] = view
	}
	return views
}
