package command

import (
	"fmt"
	"strings"

	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// EnvoyerConfirmationDepotCommand emails the vendeur that their annonce is online.
type EnvoyerConfirmationDepotCommand func(annonce *domain.Annonce) error

// BuildEnvoyerConfirmationDepotCommand sends, from fromEmail, the confirmation
// de dépôt with the links to the annonce and to its suppression page, built on
// baseURL (the frontend). The mot de passe is never part of the email.
func BuildEnvoyerConfirmationDepotCommand(m mailer.Mailer, fromEmail, baseURL string) EnvoyerConfirmationDepotCommand {
	baseURL = strings.TrimRight(baseURL, "/")
	return func(annonce *domain.Annonce) error {
		lien := fmt.Sprintf("%s/annonces/%s", baseURL, annonce.ID())
		titre := fmt.Sprintf("Votre annonce « %s » est en ligne", annonce.Titre())
		corps := fmt.Sprintf(`Bonjour %s,

Votre annonce « %s » est en ligne.

Voir l'annonce : %s

Pour la supprimer : %s/supprimer
La suppression vous demandera le mot de passe que vous avez choisi au dépôt.
`, annonce.Vendeur().Pseudo(), annonce.Titre(), lien, lien)
		return m.SendEmail(mailer.Email{To: annonce.Vendeur().Email(), From: fromEmail, Title: titre, Body: corps})
	}
}
