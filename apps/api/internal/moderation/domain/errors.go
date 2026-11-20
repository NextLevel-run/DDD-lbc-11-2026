package domain

import "errors"

// Domain errors of the moderation bounded context. Their messages are French:
// the HTTP adapter may show them as they are.
var (
	ErrAnnonceIntrouvable                 = errors.New("annonce introuvable")
	ErrModerateurInvalide                 = errors.New("le pseudo du modérateur doit comporter de 2 à 30 caractères")
	ErrAnnonceDejaVerrouillee             = errors.New("cette annonce est déjà verrouillée par un autre modérateur")
	ErrAnnoncePasEnFile                   = errors.New("cette annonce n'est pas dans la file de modération")
	ErrAnnoncePasVerrouilleeParModerateur = errors.New("cette annonce n'est pas verrouillée par ce modérateur")
	ErrMotifInvalide                      = errors.New("le motif doit comporter de 10 à 1 000 caractères")
	ErrVerrouExpire                       = errors.New("votre verrou a expiré")
	ErrConflitVersion                     = errors.New("l'historique de l'annonce a changé entre-temps")
)
