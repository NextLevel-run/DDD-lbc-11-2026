package query

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
)

// AnnonceView is the detail of an annonce as an acheteur sees it. The vendeur
// shows only their pseudo: never their email nor their telephone.
type AnnonceView struct {
	ID           string
	Titre        string
	Description  string
	Prix         int64 // in centimes; zero is a don
	Categorie    CategorieView
	Localisation LocalisationView
	Vendeur      VendeurView
	Images       []string
	DeposeeLe    time.Time
}

type CategorieView struct {
	Code    string
	Libelle string
}

type LocalisationView struct {
	Ville       string
	CodePostal  string
	Departement DepartementView
	Region      RegionView
}

type DepartementView struct {
	Code string
	Nom  string
}

type RegionView struct {
	Code string
	Nom  string
}

type VendeurView struct {
	Pseudo string
}

// ConsulterAnnonceQuery returns the detail of an activee annonce.
type ConsulterAnnonceQuery func(id string) (*AnnonceView, error)

// BuildConsulterAnnonceQuery answers ErrIDAnnonceInvalide for an ID that is
// not a UUID, and ErrAnnonceIntrouvable for an unknown annonce or one that is
// not activee.
func BuildConsulterAnnonceQuery(repo domain.AnnonceRepository) ConsulterAnnonceQuery {
	return func(id string) (*AnnonceView, error) {
		id, err := domain.ParseAnnonceID(id)
		if err != nil {
			return nil, err
		}
		annonce, err := repo.GetByID(id)
		if err != nil {
			return nil, err
		}
		if !annonce.EstActivee() {
			return nil, domain.ErrAnnonceIntrouvable
		}
		view := toAnnonceView(annonce)
		return &view, nil
	}
}

func toAnnonceView(annonce *domain.Annonce) AnnonceView {
	localisation := annonce.Localisation()
	return AnnonceView{
		ID:          annonce.ID(),
		Titre:       annonce.Titre(),
		Description: annonce.Description(),
		Prix:        annonce.Prix().Centimes(),
		Categorie: CategorieView{
			Code:    annonce.Categorie().Code(),
			Libelle: annonce.Categorie().Libelle(),
		},
		Localisation: LocalisationView{
			Ville:      localisation.Ville(),
			CodePostal: localisation.CodePostal(),
			Departement: DepartementView{
				Code: localisation.Departement().Code(),
				Nom:  localisation.Departement().Nom(),
			},
			Region: RegionView{
				Code: localisation.Region().Code(),
				Nom:  localisation.Region().Nom(),
			},
		},
		Vendeur:   VendeurView{Pseudo: annonce.Vendeur().Pseudo()},
		Images:    annonce.Images(),
		DeposeeLe: annonce.DeposeeLe(),
	}
}
