package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// CorrigerAnnonceCommandArgs is what a vendeur sends to correct a challenged
// annonce: the proof of ownership and the fields of a dépôt they may change.
type CorrigerAnnonceCommandArgs struct {
	ID          string
	MotDePasse  string // the mot de passe de l'annonce
	Titre       string
	Description string
	Prix        int64 // in centimes; zero is a don
	Categorie   string
	Ville       string
	CodePostal  string
}

// CorrigerAnnonceCommand sends a corrected annonce back to moderation.
type CorrigerAnnonceCommand func(args CorrigerAnnonceCommandArgs) error

// BuildCorrigerAnnonceCommand checks the ID, loads the annonce, has it
// corrected, persists it then publishes AnnonceCorrigee. Any error stores and
// publishes nothing.
func BuildCorrigerAnnonceCommand(repo domain.AnnonceRepository, bus eventbus.Bus, now func() time.Time) CorrigerAnnonceCommand {
	return func(args CorrigerAnnonceCommandArgs) error {
		id, err := domain.ParseAnnonceID(args.ID)
		if err != nil {
			return err
		}
		annonce, err := repo.GetByID(id)
		if err != nil {
			return err
		}
		err = annonce.Corriger(args.MotDePasse, domain.CorrectionArgs{
			Titre:       args.Titre,
			Description: args.Description,
			Prix:        args.Prix,
			Categorie:   args.Categorie,
			Ville:       args.Ville,
			CodePostal:  args.CodePostal,
		})
		if err != nil {
			return err
		}
		if err := repo.Save(annonce); err != nil {
			return err
		}
		return bus.Publish(annonceCorrigeeEvent(now(), annonce))
	}
}
