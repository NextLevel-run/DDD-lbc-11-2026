package command

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// ErrEnvoiImpossible is an application error: the email to the vendeur could
// not leave, so nothing happened and the acheteur may try again.
var ErrEnvoiImpossible = errors.New("le message n'a pas pu être envoyé")

// ContacterCommandArgs is what an acheteur sends to the vendeur of an annonce.
type ContacterCommandArgs struct {
	ID      string // the annonce
	Pseudo  string
	Email   string
	Message string
}

// ContacterCommand emails the message of an acheteur to the vendeur of an
// annonce.
type ContacterCommand func(args ContacterCommandArgs) error

// BuildContacterCommand checks the ID and the formulaire, loads the annonce
// (activee only), emails the vendeur from fromEmail with the acheteur as
// Reply-To and a link built on baseURL (the frontend), then publishes
// FormulaireContactSoumis. The email is sent synchronously: when it fails,
// the command answers ErrEnvoiImpossible and publishes nothing.
func BuildContacterCommand(repo domain.AnnonceRepository, m mailer.Mailer, bus eventbus.Bus, now func() time.Time, fromEmail, baseURL string) ContacterCommand {
	baseURL = strings.TrimRight(baseURL, "/")
	return func(args ContacterCommandArgs) error {
		id, err := domain.ParseAnnonceID(args.ID)
		if err != nil {
			return err
		}
		formulaire, err := domain.NewFormulaireContact(args.Pseudo, args.Email, args.Message)
		if err != nil {
			return err
		}
		annonce, err := repo.GetByID(id)
		if err != nil {
			return err
		}
		if !annonce.EstActivee() {
			return domain.ErrAnnonceIntrouvable
		}
		corps := fmt.Sprintf(`%s vous a écrit :

%s

Répondez directement à ce mail pour lui écrire.

Voir l'annonce : %s/annonces/%s
`, formulaire.Pseudo(), formulaire.Message(), baseURL, annonce.ID())
		if err := m.SendEmail(mailer.Email{
			To:      annonce.Vendeur().Email(),
			From:    fromEmail,
			ReplyTo: formulaire.Email(),
			Title:   fmt.Sprintf("Nouveau message pour votre annonce « %s »", annonce.Titre()),
			Body:    corps,
		}); err != nil {
			return fmt.Errorf("%w: %v", ErrEnvoiImpossible, err)
		}
		return bus.Publish(domain.NewFormulaireContactSoumisEvent(now(), annonce.ID(), formulaire))
	}
}
