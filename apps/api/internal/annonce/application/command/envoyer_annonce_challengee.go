package command

import (
	"fmt"
	"strings"

	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// EnvoyerAnnonceChallengeeCommand emails the vendeur why their annonce must be
// corrected, and where.
type EnvoyerAnnonceChallengeeCommand func(annonce *domain.Annonce) error

// BuildEnvoyerAnnonceChallengeeCommand sends, from fromEmail, the motif of the
// moderator (kept on the annonce) and the link to the correction page built on
// baseURL (the frontend).
func BuildEnvoyerAnnonceChallengeeCommand(m mailer.Mailer, fromEmail, baseURL string) EnvoyerAnnonceChallengeeCommand {
	baseURL = strings.TrimRight(baseURL, "/")
	return func(annonce *domain.Annonce) error {
		titre := fmt.Sprintf("Votre annonce « %s » doit être corrigée", annonce.Titre())
		corps := fmt.Sprintf(`Bonjour %s,

Votre annonce « %s » n'a pas pu être mise en ligne en l'état. Motif de la modération :

%s

Corrigez-la ici : %s/annonces/%s/corriger
La correction vous demandera le mot de passe que vous avez choisi au dépôt.
`, annonce.Vendeur().Pseudo(), annonce.Titre(), annonce.Motif(), baseURL, annonce.ID())
		return m.SendSimpleEmail(annonce.Vendeur().Email(), fromEmail, titre, corps)
	}
}
