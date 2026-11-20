package command

import (
	"fmt"
	"strings"

	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// EnvoyerAnnonceValideeCommand emails the vendeur that their annonce is online.
type EnvoyerAnnonceValideeCommand func(annonce *domain.Annonce) error

// BuildEnvoyerAnnonceValideeCommand sends, from fromEmail, the notice that the
// annonce is validated, with the link to its page built on baseURL (the
// frontend).
func BuildEnvoyerAnnonceValideeCommand(m mailer.Mailer, fromEmail, baseURL string) EnvoyerAnnonceValideeCommand {
	baseURL = strings.TrimRight(baseURL, "/")
	return func(annonce *domain.Annonce) error {
		titre := fmt.Sprintf("Votre annonce « %s » est validée", annonce.Titre())
		corps := fmt.Sprintf(`Bonjour %s,

Votre annonce « %s » a été validée : elle est en ligne.

Lien de l'annonce : %s/annonces/%s
`, annonce.Vendeur().Pseudo(), annonce.Titre(), baseURL, annonce.ID())
		return m.SendSimpleEmail(annonce.Vendeur().Email(), fromEmail, titre, corps)
	}
}
