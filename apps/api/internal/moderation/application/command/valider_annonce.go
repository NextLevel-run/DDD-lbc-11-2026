package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// ValiderAnnonceCommandArgs is a moderator's decision on an annonce.
type ValiderAnnonceCommandArgs struct {
	AnnonceID  string
	Moderateur string
}

// ValiderAnnonceCommand validates the annonce locked by the moderator: it goes online.
type ValiderAnnonceCommand func(args ValiderAnnonceCommandArgs) error

// BuildValiderAnnonceCommand checks the moderateur, replays the stream, has the
// aggregate decide, appends the event then publishes AnnonceValidee on the bus.
// Any error stores and publishes nothing.
func BuildValiderAnnonceCommand(repo domain.HistoriqueAnnonceRepository, bus eventbus.Bus, now func() time.Time) ValiderAnnonceCommand {
	return func(args ValiderAnnonceCommandArgs) error {
		moderateur, err := domain.NewModerateur(args.Moderateur)
		if err != nil {
			return err
		}
		h, err := repo.Charger(args.AnnonceID)
		if err != nil {
			return err
		}
		event, err := h.Valider(moderateur, now())
		if err != nil {
			return err
		}
		if err := repo.Enregistrer(h); err != nil {
			return err
		}
		return bus.Publish(event)
	}
}
