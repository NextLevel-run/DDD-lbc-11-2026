package command

import "ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

// MarquerAnnonceRefuseeCommand definitively rejects an annonce. The vendeur is
// told nothing: a refusal means fraud.
type MarquerAnnonceRefuseeCommand func(annonceID string) error

// BuildMarquerAnnonceRefuseeCommand loads the annonce, has it refused (statut
// Refusee) and persists it. It publishes nothing and sends no email.
func BuildMarquerAnnonceRefuseeCommand(repo domain.AnnonceRepository) MarquerAnnonceRefuseeCommand {
	return func(annonceID string) error {
		annonce, err := repo.GetByID(annonceID)
		if err != nil {
			return err
		}
		if err := annonce.Refuser(); err != nil {
			return err
		}
		return repo.Save(annonce)
	}
}
