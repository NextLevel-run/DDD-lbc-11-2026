package command

import "ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

// ActiverAnnonceCommand puts a validated annonce online and returns it, for
// the email that tells its vendeur.
type ActiverAnnonceCommand func(annonceID string) (*domain.Annonce, error)

// BuildActiverAnnonceCommand loads the annonce, has it activated and persists
// it. It publishes nothing: the decision belongs to moderation.
func BuildActiverAnnonceCommand(repo domain.AnnonceRepository) ActiverAnnonceCommand {
	return func(annonceID string) (*domain.Annonce, error) {
		annonce, err := repo.GetByID(annonceID)
		if err != nil {
			return nil, err
		}
		if err := annonce.Activer(); err != nil {
			return nil, err
		}
		if err := repo.Save(annonce); err != nil {
			return nil, err
		}
		return annonce, nil
	}
}
