package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// RefuserAnnonceCommandArgs is a moderator's decision on an annonce.
type RefuserAnnonceCommandArgs struct {
	AnnonceID  string
	Moderateur string
}

// RefuserAnnonceCommand refuses the annonce locked by the moderator, for fraud.
type RefuserAnnonceCommand func(args RefuserAnnonceCommandArgs) error

// BuildRefuserAnnonceCommand checks the moderateur, replays the stream, has the
// aggregate decide, appends the event then publishes AnnonceRefusee on the bus.
// Any error stores and publishes nothing.
func BuildRefuserAnnonceCommand(repo domain.HistoriqueAnnonceRepository, bus eventbus.Bus, now func() time.Time) RefuserAnnonceCommand {
	return func(args RefuserAnnonceCommandArgs) error {
		moderateur, err := domain.NewModerateur(args.Moderateur)
		if err != nil {
			return err
		}
		h, err := repo.Charger(args.AnnonceID)
		if err != nil {
			return err
		}
		event, err := h.Refuser(moderateur, now())
		if err != nil {
			return err
		}
		if err := repo.Enregistrer(h); err != nil {
			return err
		}
		return bus.Publish(annonceRefuseeEvent(event))
	}
}
