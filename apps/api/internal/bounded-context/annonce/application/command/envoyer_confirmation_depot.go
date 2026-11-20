package command

import (
	"fmt"
	"strings"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// EnvoyerConfirmationDepotCommand emails the vendeur that their annonce awaits
// moderation. It loads the annonce itself: AnnonceDeposee does not carry the
// vendeur's email.
type EnvoyerConfirmationDepotCommand func(annonceID string) error

// BuildEnvoyerConfirmationDepotCommand loads the annonce, then sends, from
// fromEmail, the confirmation de dépôt with the links to the annonce and to its
// suppression page, built on baseURL (the frontend). The mot de passe is never
// part of the email.
func BuildEnvoyerConfirmationDepotCommand(repo domain.AnnonceRepository, m mailer.Mailer, fromEmail, baseURL string) EnvoyerConfirmationDepotCommand {
	baseURL = strings.TrimRight(baseURL, "/")
	return func(annonceID string) error {
		annonce, err := repo.GetByID(annonceID)
		if err != nil {
			return err
		}
		lien := fmt.Sprintf("%s/annonces/%s", baseURL, annonce.ID())
		titre := fmt.Sprintf("Votre annonce « %s » est en cours de modération", annonce.Titre())
		corps := fmt.Sprintf(`Bonjour %s,

Votre annonce « %s » est en cours de modération : elle sera en ligne dès qu'elle aura été validée.

Lien de l'annonce : %s

Pour la supprimer : %s/supprimer
La suppression vous demandera le mot de passe que vous avez choisi au dépôt.
`, annonce.Vendeur().Pseudo(), annonce.Titre(), lien, lien)
		return m.SendSimpleEmail(annonce.Vendeur().Email(), fromEmail, titre, corps)
	}
}
