package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Statut of an annonce. Only Activee is visible to acheteurs.
type Statut string

const (
	EnAttenteModeration Statut = "EnAttenteModeration"
	Activee             Statut = "Activee"
	Challengee          Statut = "Challengee"
	Refusee             Statut = "Refusee"
	Supprimee           Statut = "Supprimee"
)

// Annonce is the aggregate root: a good a vendeur puts up for sale.
type Annonce struct {
	id           string
	titre        string
	description  string
	prix         Prix
	categorie    Categorie
	localisation Localisation
	vendeur      Vendeur
	motDePasse   MotDePasse
	images       []string
	deposeeLe    time.Time
	statut       Statut
	motif        string // why moderation challenged the annonce, empty otherwise
}

// Lengths of the titre and the description, counted in characters once
// trimmed.
const (
	titreMin       = 3
	titreMax       = 100
	descriptionMin = 10
	descriptionMax = 4000
)

// NewAnnonceArgs carries the primitives an annonce is built from.
type NewAnnonceArgs struct {
	Titre       string
	Description string
	Prix        int64
	Categorie   string
	Ville       string
	CodePostal  string
	Pseudo      string
	Email       string
	Telephone   string
	MotDePasse  string // clear text, hashed by the constructor and never kept
	DeposeeLe   time.Time
}

// NewAnnonce builds an annonce with a fresh ID, en attente de modération.
func NewAnnonce(args NewAnnonceArgs) (*Annonce, error) {
	return RestoreAnnonce(uuid.NewString(), EnAttenteModeration, args)
}

// RestoreAnnonce rebuilds an annonce whose ID and statut are already known,
// as a persistence adapter does when it loads one.
func RestoreAnnonce(id string, statut Statut, args NewAnnonceArgs) (*Annonce, error) {
	champs, err := validerChamps(CorrectionArgs{
		Titre: args.Titre, Description: args.Description, Prix: args.Prix,
		Categorie: args.Categorie, Ville: args.Ville, CodePostal: args.CodePostal,
	})
	if err != nil {
		return nil, err
	}
	vendeur, err := NewVendeur(args.Pseudo, args.Email, args.Telephone)
	if err != nil {
		return nil, err
	}
	motDePasse, err := NewMotDePasse(args.MotDePasse)
	if err != nil {
		return nil, err
	}
	return &Annonce{
		id:           id,
		titre:        champs.titre,
		description:  champs.description,
		prix:         champs.prix,
		categorie:    champs.categorie,
		localisation: champs.localisation,
		vendeur:      vendeur,
		motDePasse:   motDePasse,
		images:       []string{},
		deposeeLe:    args.DeposeeLe,
		statut:       statut,
	}, nil
}

// CorrectionArgs carries the fields a vendeur may change when correcting a
// challenged annonce: the ones of a dépôt, but the vendeur and the mot de
// passe.
type CorrectionArgs struct {
	Titre       string
	Description string
	Prix        int64
	Categorie   string
	Ville       string
	CodePostal  string
}

// champsAnnonce are the validated fields shared by a dépôt and a correction.
type champsAnnonce struct {
	titre        string
	description  string
	prix         Prix
	categorie    Categorie
	localisation Localisation
}

// validerChamps applies the rules of a dépôt, in the order of the form.
func validerChamps(args CorrectionArgs) (champsAnnonce, error) {
	titre := strings.TrimSpace(args.Titre)
	if n := utf8.RuneCountInString(titre); n < titreMin || n > titreMax {
		return champsAnnonce{}, ErrTitreInvalide
	}
	description := strings.TrimSpace(args.Description)
	if n := utf8.RuneCountInString(description); n < descriptionMin || n > descriptionMax {
		return champsAnnonce{}, ErrDescriptionInvalide
	}
	prix, err := NewPrix(args.Prix)
	if err != nil {
		return champsAnnonce{}, err
	}
	categorie, err := NewCategorie(args.Categorie)
	if err != nil {
		return champsAnnonce{}, err
	}
	localisation, err := NewLocalisation(args.Ville, args.CodePostal)
	if err != nil {
		return champsAnnonce{}, err
	}
	return champsAnnonce{titre, description, prix, categorie, localisation}, nil
}

// ParseAnnonceID checks that id is a UUID, the shape of every annonce ID.
func ParseAnnonceID(id string) (string, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", ErrIDAnnonceInvalide
	}
	return parsed.String(), nil
}

func (a *Annonce) ID() string                 { return a.id }
func (a *Annonce) Titre() string              { return a.titre }
func (a *Annonce) Description() string        { return a.description }
func (a *Annonce) Prix() Prix                 { return a.prix }
func (a *Annonce) Categorie() Categorie       { return a.categorie }
func (a *Annonce) Localisation() Localisation { return a.localisation }
func (a *Annonce) Vendeur() Vendeur           { return a.vendeur }

// MotDePasseCorrespond tells whether saisie is the mot de passe de l'annonce.
func (a *Annonce) MotDePasseCorrespond(saisie string) bool { return a.motDePasse.Correspond(saisie) }
func (a *Annonce) Images() []string                        { return append([]string{}, a.images...) }
func (a *Annonce) DeposeeLe() time.Time                    { return a.deposeeLe }
func (a *Annonce) Statut() Statut                          { return a.statut }
func (a *Annonce) Motif() string                           { return a.motif }
func (a *Annonce) EstActivee() bool                        { return a.statut == Activee }

// Supprimer withdraws the annonce when saisie is its mot de passe. A
// non-activee annonce answers ErrAnnonceIntrouvable, so that its existence
// does not leak; a wrong mot de passe answers ErrMotDePasseIncorrect and
// leaves the annonce en ligne.
func (a *Annonce) Supprimer(saisie string) error {
	if !a.EstActivee() {
		return ErrAnnonceIntrouvable
	}
	if !a.MotDePasseCorrespond(saisie) {
		return ErrMotDePasseIncorrect
	}
	a.statut = Supprimee
	return nil
}

// Activer puts a validated annonce online.
func (a *Annonce) Activer() error {
	if a.statut != EnAttenteModeration {
		return ErrTransitionStatutInvalide
	}
	a.statut = Activee
	return nil
}

// Challenger sends the annonce back to its vendeur, who must correct it; the
// motif says why.
func (a *Annonce) Challenger(motif string) error {
	if a.statut != EnAttenteModeration {
		return ErrTransitionStatutInvalide
	}
	a.statut = Challengee
	a.motif = motif
	return nil
}

// Refuser definitively rejects the annonce.
func (a *Annonce) Refuser() error {
	if a.statut != EnAttenteModeration {
		return ErrTransitionStatutInvalide
	}
	a.statut = Refusee
	return nil
}

// AutoriserCorrection tells whether saisie lets the vendeur correct the
// annonce. Only a challengee annonce is correctable: any other answers
// ErrAnnonceIntrouvable, whatever the mot de passe, so that its existence does
// not leak. A wrong mot de passe answers ErrMotDePasseIncorrect.
func (a *Annonce) AutoriserCorrection(saisie string) error {
	if a.statut != Challengee {
		return ErrAnnonceIntrouvable
	}
	if !a.MotDePasseCorrespond(saisie) {
		return ErrMotDePasseIncorrect
	}
	return nil
}

// Corriger replaces the fields of a challengee annonce with the vendeur's
// correction, clears the motif and sends it back to moderation. Vendeur, mot
// de passe and date de dépôt do not change. The authorisation comes first,
// then the validations of a dépôt; on any error nothing changes.
func (a *Annonce) Corriger(saisie string, correction CorrectionArgs) error {
	if err := a.AutoriserCorrection(saisie); err != nil {
		return err
	}
	champs, err := validerChamps(correction)
	if err != nil {
		return err
	}
	a.titre, a.description, a.prix = champs.titre, champs.description, champs.prix
	a.categorie, a.localisation = champs.categorie, champs.localisation
	a.motif = ""
	a.statut = EnAttenteModeration
	return nil
}
