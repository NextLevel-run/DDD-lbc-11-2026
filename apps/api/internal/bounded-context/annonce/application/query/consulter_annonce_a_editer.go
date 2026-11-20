package query

import "ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

// AnnonceAEditerView is what the correction form of a vendeur is prefilled
// with: the fields they may change and the motif of the challenge. Never the
// vendeur's email nor telephone.
type AnnonceAEditerView struct {
	Titre       string
	Description string
	Prix        int64 // in centimes; zero is a don
	Categorie   string
	Ville       string
	CodePostal  string
	Motif       string
}

// ConsulterAnnonceAEditerQuery returns the editable fields of a challengee
// annonce to its vendeur, who proves it with the mot de passe.
type ConsulterAnnonceAEditerQuery func(id, motDePasse string) (*AnnonceAEditerView, error)

// BuildConsulterAnnonceAEditerQuery answers ErrIDAnnonceInvalide for an ID
// that is not a UUID, ErrAnnonceIntrouvable for an unknown annonce or one that
// is not challengee, and ErrMotDePasseIncorrect for a wrong mot de passe.
func BuildConsulterAnnonceAEditerQuery(repo domain.AnnonceRepository) ConsulterAnnonceAEditerQuery {
	return func(id, motDePasse string) (*AnnonceAEditerView, error) {
		id, err := domain.ParseAnnonceID(id)
		if err != nil {
			return nil, err
		}
		annonce, err := repo.GetByID(id)
		if err != nil {
			return nil, err
		}
		if err := annonce.AutoriserCorrection(motDePasse); err != nil {
			return nil, err
		}
		localisation := annonce.Localisation()
		return &AnnonceAEditerView{
			Titre:       annonce.Titre(),
			Description: annonce.Description(),
			Prix:        annonce.Prix().Centimes(),
			Categorie:   annonce.Categorie().Code(),
			Ville:       localisation.Ville(),
			CodePostal:  localisation.CodePostal(),
			Motif:       annonce.Motif(),
		}, nil
	}
}
