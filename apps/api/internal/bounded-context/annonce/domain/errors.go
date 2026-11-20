package domain

import "errors"

// Domain errors of the annonce bounded context. Their messages are French:
// the HTTP adapter may show them as they are.
var (
	ErrAnnonceIntrouvable        = errors.New("annonce introuvable")
	ErrTransitionStatutInvalide  = errors.New("cette annonce ne peut pas changer de statut")
	ErrIDAnnonceInvalide         = errors.New("identifiant d'annonce invalide")
	ErrTitreInvalide             = errors.New("le titre doit comporter de 3 à 100 caractères")
	ErrDescriptionInvalide       = errors.New("la description doit comporter de 10 à 4 000 caractères")
	ErrPseudoInvalide            = errors.New("le pseudo doit comporter de 2 à 30 caractères")
	ErrEmailInvalide             = errors.New("l'email n'est pas valide")
	ErrTelephoneInvalide         = errors.New("le téléphone doit comporter 10 chiffres et commencer par 0")
	ErrMotDePasseTropCourt       = errors.New("le mot de passe de l'annonce doit comporter au moins 8 caractères")
	ErrMotDePasseIncorrect       = errors.New("le mot de passe de l'annonce est incorrect")
	ErrRaisonSuppressionInvalide = errors.New("la raison de suppression est obligatoire et doit être connue")
	ErrPrixNegatif               = errors.New("le prix ne peut pas être négatif")
	ErrCategorieInconnue         = errors.New("catégorie inconnue")
	ErrVilleVide                 = errors.New("la ville est obligatoire")
	ErrCodePostalInvalide        = errors.New("le code postal doit comporter 5 chiffres")
	ErrDepartementInconnu        = errors.New("le code postal ne désigne aucun département")
	ErrPageInvalide              = errors.New("le numéro de page doit être positif")
	ErrPrixMinInvalide           = errors.New("le prix minimum doit être un montant positif ou nul")
	ErrPrixMaxInvalide           = errors.New("le prix maximum doit être un montant positif ou nul")
	ErrFourchettePrixInvalide    = errors.New("le prix minimum ne peut pas dépasser le prix maximum")
	ErrRegionInconnue            = errors.New("région inconnue")
	ErrCodeDepartementInconnu    = errors.New("département inconnu")
	ErrLieuAmbigu                = errors.New("choisir une région ou un département, pas les deux")
)
