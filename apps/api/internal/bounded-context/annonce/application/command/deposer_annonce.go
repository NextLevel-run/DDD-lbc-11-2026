package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// DeposerAnnonceCommandArgs is what a vendeur fills in to deposit an annonce.
type DeposerAnnonceCommandArgs struct {
	Titre       string
	Description string
	Prix        int64 // in centimes; zero is a don
	Categorie   string
	Ville       string
	CodePostal  string
	Pseudo      string
	Email       string
	Telephone   string // optional: empty means none
	MotDePasse  string // the mot de passe de l'annonce, at least 8 characters
}

// DeposerAnnonceCommand publishes a new annonce and returns its ID.
type DeposerAnnonceCommand func(args DeposerAnnonceCommandArgs) (string, error)

// BuildDeposerAnnonceCommand creates the annonce activee, dated by now,
// persists it, then publishes AnnonceDeposee. A domain error leaves nothing
// stored and publishes nothing.
func BuildDeposerAnnonceCommand(repo domain.AnnonceRepository, bus eventbus.Bus, now func() time.Time) DeposerAnnonceCommand {
	return func(args DeposerAnnonceCommandArgs) (string, error) {
		deposeeLe := now()
		annonce, err := domain.NewAnnonce(domain.NewAnnonceArgs{
			Titre:       args.Titre,
			Description: args.Description,
			Prix:        args.Prix,
			Categorie:   args.Categorie,
			Ville:       args.Ville,
			CodePostal:  args.CodePostal,
			Pseudo:      args.Pseudo,
			Email:       args.Email,
			Telephone:   args.Telephone,
			MotDePasse:  args.MotDePasse,
			DeposeeLe:   deposeeLe,
		})
		if err != nil {
			return "", err
		}
		if err := repo.Save(annonce); err != nil {
			return "", err
		}
		if err := bus.Publish(annonceDeposeeEvent(deposeeLe, annonce)); err != nil {
			return "", err
		}
		return annonce.ID(), nil
	}
}
