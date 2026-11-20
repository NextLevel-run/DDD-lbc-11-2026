package command

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// ChallengerAnnonceCommandArgs is a moderator sending an annonce back to its
// vendeur.
type ChallengerAnnonceCommandArgs struct {
	AnnonceID  string
	Moderateur string
	Motif      string
}

// ChallengerAnnonceCommand asks the vendeur to correct the annonce locked by
// the moderator, with a motif.
type ChallengerAnnonceCommand func(args ChallengerAnnonceCommandArgs) error

// BuildChallengerAnnonceCommand checks the moderateur and the motif, replays
// the stream, has the aggregate decide, appends the event then publishes
// AnnonceChallengee on the bus. Any error stores and publishes nothing.
func BuildChallengerAnnonceCommand(repo domain.HistoriqueAnnonceRepository, bus eventbus.Bus, now func() time.Time) ChallengerAnnonceCommand {
	return func(args ChallengerAnnonceCommandArgs) error {
		moderateur, err := domain.NewModerateur(args.Moderateur)
		if err != nil {
			return err
		}
		motif, err := domain.NewMotif(args.Motif)
		if err != nil {
			return err
		}
		h, err := repo.Charger(args.AnnonceID)
		if err != nil {
			return err
		}
		event, err := h.Challenger(moderateur, motif, now())
		if err != nil {
			return err
		}
		if err := repo.Enregistrer(h); err != nil {
			return err
		}
		return bus.Publish(annonceChallengeeEvent(event))
	}
}
