package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// VerrouillerAnnonceCommandArgs is a moderator taking an annonce.
type VerrouillerAnnonceCommandArgs struct {
	AnnonceID  string
	Moderateur string
}

// VerrouillerAnnonceCommand locks an annonce of the queue for a moderator.
type VerrouillerAnnonceCommand func(args VerrouillerAnnonceCommandArgs) error

// BuildVerrouillerAnnonceCommand checks the moderateur, replays the stream,
// has the aggregate decide (it frees a lock older than 2 hours by itself, on
// the injected clock) and appends the new events. It publishes nothing: lock
// events stay in the stream.
func BuildVerrouillerAnnonceCommand(repo domain.HistoriqueAnnonceRepository, now func() time.Time) VerrouillerAnnonceCommand {
	return func(args VerrouillerAnnonceCommandArgs) error {
		moderateur, err := domain.NewModerateur(args.Moderateur)
		if err != nil {
			return err
		}
		h, err := repo.Charger(args.AnnonceID)
		if err != nil {
			return err
		}
		if err := h.Verrouiller(moderateur, now()); err != nil {
			return err
		}
		return repo.Enregistrer(h)
	}
}
