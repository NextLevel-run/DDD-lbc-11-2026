package domain

// Categorie is one value of a closed list. Its code is stable (it travels in
// JSON and URLs); its libelle is what a user reads.
type Categorie string

const (
	CategorieVehicules      Categorie = "vehicules"
	CategorieImmobilier     Categorie = "immobilier"
	CategorieMultimedia     Categorie = "multimedia"
	CategorieMaisonJardin   Categorie = "maison-jardin"
	CategorieMode           Categorie = "mode"
	CategorieLoisirs        Categorie = "loisirs"
	CategorieEmploiServices Categorie = "emploi-services"
	CategorieAutres         Categorie = "autres"
)

var categories = []Categorie{
	CategorieVehicules,
	CategorieImmobilier,
	CategorieMultimedia,
	CategorieMaisonJardin,
	CategorieMode,
	CategorieLoisirs,
	CategorieEmploiServices,
	CategorieAutres,
}

var libelles = map[Categorie]string{
	CategorieVehicules:      "Véhicules",
	CategorieImmobilier:     "Immobilier",
	CategorieMultimedia:     "Multimédia",
	CategorieMaisonJardin:   "Maison & Jardin",
	CategorieMode:           "Mode",
	CategorieLoisirs:        "Loisirs",
	CategorieEmploiServices: "Emploi & Services",
	CategorieAutres:         "Autres",
}

func NewCategorie(code string) (Categorie, error) {
	categorie := Categorie(code)
	if _, ok := libelles[categorie]; !ok {
		return "", ErrCategorieInconnue
	}
	return categorie, nil
}

// Categories returns the closed list, in display order.
func Categories() []Categorie {
	return append([]Categorie(nil), categories...)
}

func (c Categorie) Code() string    { return string(c) }
func (c Categorie) Libelle() string { return libelles[c] }
