package domain

// AnnonceRepository stores and loads annonces.
type AnnonceRepository interface {
	Save(annonce *Annonce) error
	// GetByID returns ErrAnnonceIntrouvable when no annonce has this ID.
	GetByID(id string) (*Annonce, error)
	// Rechercher returns the requested page of the annonces the criteres
	// retain, most recently deposited first, and their total count.
	Rechercher(criteres CriteresRecherche, page Page) (ResultatRecherche, error)
}
