// Package query holds the read use cases of moderation. The queue is computed
// here, by replaying the streams: it is not an aggregate.
package query

import (
	"sort"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// AnnonceAModererResumeView is a line of the queue.
type AnnonceAModererResumeView struct {
	ID               string
	Titre            string
	Prix             int64 // in centimes
	Ville            string
	Pseudo           string
	MisEnFileLe      time.Time
	VerrouillePar    string    // empty when free
	VerrouilleDepuis time.Time // zero when free
}

// ListerAnnoncesAModererQuery returns the annonces waiting for a moderator,
// oldest first.
type ListerAnnoncesAModererQuery func() ([]AnnonceAModererResumeView, error)

// BuildListerAnnoncesAModererQuery replays every stream and keeps the annonces
// en file or verrouillee. A lock older than 2 hours counts as free, on the
// injected clock. Equal dates are ordered by ID so the list is stable.
func BuildListerAnnoncesAModererQuery(repo domain.HistoriqueAnnonceRepository, now func() time.Time) ListerAnnoncesAModererQuery {
	return func() ([]AnnonceAModererResumeView, error) {
		historiques, err := repo.Tous()
		if err != nil {
			return nil, err
		}
		maintenant := now()
		views := []AnnonceAModererResumeView{}
		for _, h := range historiques {
			if !h.EstAModerer() {
				continue
			}
			annonce := h.Annonce()
			view := AnnonceAModererResumeView{
				ID: h.AnnonceID(), Titre: annonce.Titre, Prix: annonce.Prix, Ville: annonce.Ville,
				Pseudo: annonce.Pseudo, MisEnFileLe: h.MisEnFileLe(),
			}
			if moderateur, depuis, ok := h.Verrou(maintenant); ok {
				view.VerrouillePar, view.VerrouilleDepuis = moderateur, depuis
			}
			views = append(views, view)
		}
		sort.SliceStable(views, func(i, j int) bool {
			if !views[i].MisEnFileLe.Equal(views[j].MisEnFileLe) {
				return views[i].MisEnFileLe.Before(views[j].MisEnFileLe)
			}
			return views[i].ID < views[j].ID
		})
		return views, nil
	}
}
