package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/internal/contracts"
)

// instantaneAnnonce is the annonce as the other contexts may read it: primitives
// only, never the vendeur's email, telephone or mot de passe.
func instantaneAnnonce(annonce *domain.Annonce) contracts.InstantaneAnnonce {
	return contracts.InstantaneAnnonce{
		AnnonceID:   annonce.ID(),
		Titre:       annonce.Titre(),
		Description: annonce.Description(),
		Prix:        annonce.Prix().Centimes(),
		Categorie:   annonce.Categorie().Code(),
		Ville:       annonce.Localisation().Ville(),
		CodePostal:  annonce.Localisation().CodePostal(),
		Pseudo:      annonce.Vendeur().Pseudo(),
		DeposeeLe:   annonce.DeposeeLe(),
	}
}

func annonceDeposeeEvent(date time.Time, annonce *domain.Annonce) contracts.AnnonceDeposeeEvent {
	return contracts.AnnonceDeposeeEvent{Annonce: instantaneAnnonce(annonce), Date: date}
}

func annonceCorrigeeEvent(date time.Time, annonce *domain.Annonce) contracts.AnnonceCorrigeeEvent {
	return contracts.AnnonceCorrigeeEvent{Annonce: instantaneAnnonce(annonce), Date: date}
}
