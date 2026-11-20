package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// MettreEnFileCommandArgs is the annonce a dépôt produced, in primitives.
type MettreEnFileCommandArgs struct {
	AnnonceID   string
	Titre       string
	Description string
	Prix        int64 // in centimes
	Categorie   string
	Ville       string
	CodePostal  string
	Pseudo      string
	DeposeeLe   time.Time
}

// MettreEnFileCommand puts a deposited annonce in the moderation queue.
type MettreEnFileCommand func(args MettreEnFileCommandArgs) error

// BuildMettreEnFileCommand loads the stream of the annonce, appends its
// mise en file and stores it. An annonce already known is left alone.
func BuildMettreEnFileCommand(repo domain.HistoriqueAnnonceRepository, now func() time.Time) MettreEnFileCommand {
	return func(args MettreEnFileCommandArgs) error {
		h, err := repo.Charger(args.AnnonceID)
		if err != nil {
			return err
		}
		h.MettreEnFile(domain.InstantaneAnnonce{
			AnnonceID:   args.AnnonceID,
			Titre:       args.Titre,
			Description: args.Description,
			Prix:        args.Prix,
			Categorie:   args.Categorie,
			Ville:       args.Ville,
			CodePostal:  args.CodePostal,
			Pseudo:      args.Pseudo,
			DeposeeLe:   args.DeposeeLe,
		}, now())
		return repo.Enregistrer(h)
	}
}
