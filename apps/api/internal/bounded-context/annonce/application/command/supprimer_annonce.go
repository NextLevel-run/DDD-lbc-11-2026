package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// SupprimerAnnonceCommandArgs is what a vendeur sends to withdraw an annonce.
type SupprimerAnnonceCommandArgs struct {
	ID         string
	MotDePasse string // the mot de passe de l'annonce
	Raison     string // a code of the closed list of raisons
}

// SupprimerAnnonceCommand withdraws an annonce, the mot de passe being the
// proof of ownership.
type SupprimerAnnonceCommand func(args SupprimerAnnonceCommandArgs) error

// BuildSupprimerAnnonceCommand checks the ID and the raison, loads the
// annonce, has it deleted, persists it (it stays in the repository, statut
// Supprimee) then publishes AnnonceSupprimee. Any error stores and publishes
// nothing.
func BuildSupprimerAnnonceCommand(repo domain.AnnonceRepository, bus eventbus.Bus, now func() time.Time) SupprimerAnnonceCommand {
	return func(args SupprimerAnnonceCommandArgs) error {
		id, err := domain.ParseAnnonceID(args.ID)
		if err != nil {
			return err
		}
		raison, err := domain.NewRaisonSuppression(args.Raison)
		if err != nil {
			return err
		}
		annonce, err := repo.GetByID(id)
		if err != nil {
			return err
		}
		if err := annonce.Supprimer(args.MotDePasse); err != nil {
			return err
		}
		if err := repo.Save(annonce); err != nil {
			return err
		}
		return bus.Publish(domain.NewAnnonceSupprimeeEvent(now(), annonce.ID(), raison))
	}
}
