package domain

import (
	"math"
	"strings"
)

// ParPage is the number of annonces on one page of a recherche.
const ParPage = 20

// CriteresRechercheArgs carries the primitives the criteres of a recherche
// are built from. Every criterion is optional: an empty Categorie and a nil
// bound are no criterion.
type CriteresRechercheArgs struct {
	MotCle    string
	Categorie string
	// Bounds of the fourchette de prix, in centimes, both included.
	PrixMin *int64
	PrixMax *int64
	// The lieu: an INSEE region or departement code, never both.
	Region      string
	Departement string
	// Five digits are a code postal, anything else a ville.
	VilleOuCodePostal string
}

// CriteresRecherche selects the annonces a recherche finds. It only ever
// keeps activee annonces; its criteria combine with a logical AND.
type CriteresRecherche struct {
	// Each word of the mot-clé, normalised; none means every annonce.
	mots []string
	// The empty Categorie means every catégorie.
	categorie Categorie
	prixMin   *Prix
	prixMax   *Prix
	// At most one of region and departement is set.
	region      *Region
	departement *Departement
	// An exact code postal, or a normalised ville; both empty for none.
	codePostal string
	ville      string
}

// NewCriteresRecherche builds criteres from what the acheteur typed. A blank
// mot-clé or ville is no criterion. An unknown catégorie, region or
// departement, both a region and a departement, a negative bound or a minimum
// above the maximum is refused.
func NewCriteresRecherche(args CriteresRechercheArgs) (CriteresRecherche, error) {
	criteres := CriteresRecherche{mots: strings.Fields(normaliser(args.MotCle))}
	if args.Categorie != "" {
		categorie, err := NewCategorie(args.Categorie)
		if err != nil {
			return CriteresRecherche{}, err
		}
		criteres.categorie = categorie
	}
	var err error
	if criteres.prixMin, err = borneDePrix(args.PrixMin, ErrPrixMinInvalide); err != nil {
		return CriteresRecherche{}, err
	}
	if criteres.prixMax, err = borneDePrix(args.PrixMax, ErrPrixMaxInvalide); err != nil {
		return CriteresRecherche{}, err
	}
	if criteres.prixMin != nil && criteres.prixMax != nil &&
		criteres.prixMin.Centimes() > criteres.prixMax.Centimes() {
		return CriteresRecherche{}, ErrFourchettePrixInvalide
	}
	if err := criteres.lireLieu(args); err != nil {
		return CriteresRecherche{}, err
	}
	return criteres, nil
}

// lireLieu sets the region or the departement, and the ville or the code
// postal.
func (c *CriteresRecherche) lireLieu(args CriteresRechercheArgs) error {
	if args.Region != "" && args.Departement != "" {
		return ErrLieuAmbigu
	}
	if args.Region != "" {
		region, err := NewRegion(args.Region)
		if err != nil {
			return err
		}
		c.region = &region
	}
	if args.Departement != "" {
		departement, err := NewDepartement(args.Departement)
		if err != nil {
			return err
		}
		c.departement = &departement
	}
	saisie := strings.TrimSpace(args.VilleOuCodePostal)
	if codePostalFormat.MatchString(saisie) {
		c.codePostal = saisie
	} else {
		c.ville = normaliser(saisie)
	}
	return nil
}

// borneDePrix reads an optional bound as a Prix, refusing it with errInvalide
// when it is negative.
func borneDePrix(centimes *int64, errInvalide error) (*Prix, error) {
	if centimes == nil {
		return nil, nil
	}
	prix, err := NewPrix(*centimes)
	if err != nil {
		return nil, errInvalide
	}
	return &prix, nil
}

// Retient tells whether the annonce matches: activee, of the catégorie, its
// prix within the fourchette, in the region or departement, of the code
// postal or the ville (identical once normalised), and every word of the
// mot-clé found in its titre or its description, ignoring case and accents.
func (c CriteresRecherche) Retient(annonce *Annonce) bool {
	if !annonce.EstActivee() {
		return false
	}
	if c.categorie != "" && annonce.Categorie() != c.categorie {
		return false
	}
	prix := annonce.Prix().Centimes()
	if c.prixMin != nil && prix < c.prixMin.Centimes() {
		return false
	}
	if c.prixMax != nil && prix > c.prixMax.Centimes() {
		return false
	}
	if !c.retientLieu(annonce.Localisation()) {
		return false
	}
	if len(c.mots) == 0 {
		return true
	}
	texte := normaliser(annonce.Titre() + " " + annonce.Description())
	for _, mot := range c.mots {
		if !strings.Contains(texte, mot) {
			return false
		}
	}
	return true
}

func (c CriteresRecherche) retientLieu(localisation Localisation) bool {
	if c.region != nil && localisation.Region() != *c.region {
		return false
	}
	if c.departement != nil && localisation.Departement() != *c.departement {
		return false
	}
	if c.codePostal != "" && localisation.CodePostal() != c.codePostal {
		return false
	}
	if c.ville != "" && normaliser(localisation.Ville()) != c.ville {
		return false
	}
	return true
}

// Page is the number of a page of results, counted from 1.
type Page struct {
	numero int
}

// NewPage takes zero as the first page, the default when none is asked for,
// and refuses a negative number.
func NewPage(numero int) (Page, error) {
	if numero < 0 {
		return Page{}, ErrPageInvalide
	}
	if numero == 0 {
		numero = 1
	}
	return Page{numero: numero}, nil
}

func (p Page) Numero() int { return p.numero }

// Offset is the number of annonces the pages before this one hold. A page
// so far past the last that the count would overflow saturates at MaxInt.
func (p Page) Offset() int {
	if p.numero-1 > math.MaxInt/ParPage {
		return math.MaxInt
	}
	return (p.numero - 1) * ParPage
}

// ResultatRecherche is one page of the annonces a recherche finds, most
// recent first, and how many it finds in all.
type ResultatRecherche struct {
	Annonces []*Annonce
	Total    int
}
