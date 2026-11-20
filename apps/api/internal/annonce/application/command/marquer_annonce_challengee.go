package command

import "ddd-second-hand-marketplace/internal/annonce/domain"

// MarquerAnnonceChallengeeCommand sends an annonce back to its vendeur with
// the motif of the moderator, and returns it, for the email that tells them.
type MarquerAnnonceChallengeeCommand func(annonceID, motif string) (*domain.Annonce, error)

// BuildMarquerAnnonceChallengeeCommand loads the annonce, has it challenged
// (statut Challengee, motif kept) and persists it. It publishes nothing.
func BuildMarquerAnnonceChallengeeCommand(repo domain.AnnonceRepository) MarquerAnnonceChallengeeCommand {
	return func(annonceID, motif string) (*domain.Annonce, error) {
		annonce, err := repo.GetByID(annonceID)
		if err != nil {
			return nil, err
		}
		if err := annonce.Challenger(motif); err != nil {
			return nil, err
		}
		if err := repo.Save(annonce); err != nil {
			return nil, err
		}
		return annonce, nil
	}
}
