package query

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
)

// RechercherAnnoncesArgs are the criteria of a recherche, all optional, and
// the page wanted (zero is the first).
type RechercherAnnoncesArgs struct {
	MotCle    string
	Categorie string // a catégorie code; empty for every catégorie
	// Bounds of the fourchette de prix in centimes, both included; nil for none.
	PrixMin *int64
	PrixMax *int64
	// The lieu: a region or a departement INSEE code, never both; empty for
	// none.
	Region      string
	Departement string
	// Five digits are a code postal, anything else a ville; empty for none.
	VilleOuCodePostal string
	Page              int
}

// RechercherAnnoncesView is one page of results and what a pagination needs.
type RechercherAnnoncesView struct {
	Annonces []AnnonceResumeView
	Total    int
	Page     int
	ParPage  int
}

// AnnonceResumeView is an annonce as a list of results shows it.
type AnnonceResumeView struct {
	ID        string
	Titre     string
	Prix      int64 // in centimes; zero is a don
	Ville     string
	Images    []string
	DeposeeLe time.Time
}

// RechercherAnnoncesQuery finds the activee annonces matching the criteria,
// most recent first, one page at a time.
type RechercherAnnoncesQuery func(args RechercherAnnoncesArgs) (RechercherAnnoncesView, error)

func BuildRechercherAnnoncesQuery(repo domain.AnnonceRepository) RechercherAnnoncesQuery {
	return func(args RechercherAnnoncesArgs) (RechercherAnnoncesView, error) {
		criteres, err := domain.NewCriteresRecherche(domain.CriteresRechercheArgs{
			MotCle:            args.MotCle,
			Categorie:         args.Categorie,
			PrixMin:           args.PrixMin,
			PrixMax:           args.PrixMax,
			Region:            args.Region,
			Departement:       args.Departement,
			VilleOuCodePostal: args.VilleOuCodePostal,
		})
		if err != nil {
			return RechercherAnnoncesView{}, err
		}
		page, err := domain.NewPage(args.Page)
		if err != nil {
			return RechercherAnnoncesView{}, err
		}
		resultat, err := repo.Rechercher(criteres, page)
		if err != nil {
			return RechercherAnnoncesView{}, err
		}
		annonces := make([]AnnonceResumeView, len(resultat.Annonces))
		for i, annonce := range resultat.Annonces {
			annonces[i] = AnnonceResumeView{
				ID:        annonce.ID(),
				Titre:     annonce.Titre(),
				Prix:      annonce.Prix().Centimes(),
				Ville:     annonce.Localisation().Ville(),
				Images:    annonce.Images(),
				DeposeeLe: annonce.DeposeeLe(),
			}
		}
		return RechercherAnnoncesView{
			Annonces: annonces,
			Total:    resultat.Total,
			Page:     page.Numero(),
			ParPage:  domain.ParPage,
		}, nil
	}
}
