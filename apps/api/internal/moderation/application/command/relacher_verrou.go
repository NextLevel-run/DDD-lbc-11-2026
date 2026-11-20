package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// RelacherVerrouCommandArgs is a moderator giving an annonce back.
type RelacherVerrouCommandArgs struct {
	AnnonceID  string
	Moderateur string
}

// RelacherVerrouCommand releases the lock of an annonce, back to the queue.
type RelacherVerrouCommand func(args RelacherVerrouCommandArgs) error

// BuildRelacherVerrouCommand checks the moderateur, replays the stream, has
// the aggregate release and appends the event. It publishes nothing.
func BuildRelacherVerrouCommand(repo domain.HistoriqueAnnonceRepository, now func() time.Time) RelacherVerrouCommand {
	return func(args RelacherVerrouCommandArgs) error {
		moderateur, err := domain.NewModerateur(args.Moderateur)
		if err != nil {
			return err
		}
		h, err := repo.Charger(args.AnnonceID)
		if err != nil {
			return err
		}
		if err := h.Relacher(moderateur, now()); err != nil {
			return err
		}
		return repo.Enregistrer(h)
	}
}
