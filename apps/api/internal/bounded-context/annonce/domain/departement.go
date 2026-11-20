package domain

// Region is a French region, identified by its INSEE code.
type Region struct {
	code string
	nom  string
}

func (r Region) Code() string { return r.code }
func (r Region) Nom() string  { return r.nom }

// Departement is a French departement, identified by its INSEE code (01, 2A,
// 974…), and the region it belongs to.
type Departement struct {
	code   string
	nom    string
	region Region
}

func (d Departement) Code() string   { return d.code }
func (d Departement) Nom() string    { return d.nom }
func (d Departement) Region() Region { return d.region }

var (
	auvergneRhoneAlpes     = Region{"84", "Auvergne-Rhône-Alpes"}
	bourgogneFrancheComte  = Region{"27", "Bourgogne-Franche-Comté"}
	bretagne               = Region{"53", "Bretagne"}
	centreValDeLoire       = Region{"24", "Centre-Val de Loire"}
	corse                  = Region{"94", "Corse"}
	grandEst               = Region{"44", "Grand Est"}
	hautsDeFrance          = Region{"32", "Hauts-de-France"}
	ileDeFrance            = Region{"11", "Île-de-France"}
	normandie              = Region{"28", "Normandie"}
	nouvelleAquitaine      = Region{"75", "Nouvelle-Aquitaine"}
	occitanie              = Region{"76", "Occitanie"}
	paysDeLaLoire          = Region{"52", "Pays de la Loire"}
	provenceAlpesCoteDAzur = Region{"93", "Provence-Alpes-Côte d'Azur"}
	guadeloupe             = Region{"01", "Guadeloupe"}
	martinique             = Region{"02", "Martinique"}
	guyane                 = Region{"03", "Guyane"}
	laReunion              = Region{"04", "La Réunion"}
	mayotte                = Region{"06", "Mayotte"}
)

var regions = []Region{
	auvergneRhoneAlpes, bourgogneFrancheComte, bretagne, centreValDeLoire, corse,
	grandEst, hautsDeFrance, ileDeFrance, normandie, nouvelleAquitaine, occitanie,
	paysDeLaLoire, provenceAlpesCoteDAzur, guadeloupe, martinique, guyane,
	laReunion, mayotte,
}

var departements = []Departement{
	{"01", "Ain", auvergneRhoneAlpes},
	{"02", "Aisne", hautsDeFrance},
	{"03", "Allier", auvergneRhoneAlpes},
	{"04", "Alpes-de-Haute-Provence", provenceAlpesCoteDAzur},
	{"05", "Hautes-Alpes", provenceAlpesCoteDAzur},
	{"06", "Alpes-Maritimes", provenceAlpesCoteDAzur},
	{"07", "Ardèche", auvergneRhoneAlpes},
	{"08", "Ardennes", grandEst},
	{"09", "Ariège", occitanie},
	{"10", "Aube", grandEst},
	{"11", "Aude", occitanie},
	{"12", "Aveyron", occitanie},
	{"13", "Bouches-du-Rhône", provenceAlpesCoteDAzur},
	{"14", "Calvados", normandie},
	{"15", "Cantal", auvergneRhoneAlpes},
	{"16", "Charente", nouvelleAquitaine},
	{"17", "Charente-Maritime", nouvelleAquitaine},
	{"18", "Cher", centreValDeLoire},
	{"19", "Corrèze", nouvelleAquitaine},
	{"2A", "Corse-du-Sud", corse},
	{"2B", "Haute-Corse", corse},
	{"21", "Côte-d'Or", bourgogneFrancheComte},
	{"22", "Côtes-d'Armor", bretagne},
	{"23", "Creuse", nouvelleAquitaine},
	{"24", "Dordogne", nouvelleAquitaine},
	{"25", "Doubs", bourgogneFrancheComte},
	{"26", "Drôme", auvergneRhoneAlpes},
	{"27", "Eure", normandie},
	{"28", "Eure-et-Loir", centreValDeLoire},
	{"29", "Finistère", bretagne},
	{"30", "Gard", occitanie},
	{"31", "Haute-Garonne", occitanie},
	{"32", "Gers", occitanie},
	{"33", "Gironde", nouvelleAquitaine},
	{"34", "Hérault", occitanie},
	{"35", "Ille-et-Vilaine", bretagne},
	{"36", "Indre", centreValDeLoire},
	{"37", "Indre-et-Loire", centreValDeLoire},
	{"38", "Isère", auvergneRhoneAlpes},
	{"39", "Jura", bourgogneFrancheComte},
	{"40", "Landes", nouvelleAquitaine},
	{"41", "Loir-et-Cher", centreValDeLoire},
	{"42", "Loire", auvergneRhoneAlpes},
	{"43", "Haute-Loire", auvergneRhoneAlpes},
	{"44", "Loire-Atlantique", paysDeLaLoire},
	{"45", "Loiret", centreValDeLoire},
	{"46", "Lot", occitanie},
	{"47", "Lot-et-Garonne", nouvelleAquitaine},
	{"48", "Lozère", occitanie},
	{"49", "Maine-et-Loire", paysDeLaLoire},
	{"50", "Manche", normandie},
	{"51", "Marne", grandEst},
	{"52", "Haute-Marne", grandEst},
	{"53", "Mayenne", paysDeLaLoire},
	{"54", "Meurthe-et-Moselle", grandEst},
	{"55", "Meuse", grandEst},
	{"56", "Morbihan", bretagne},
	{"57", "Moselle", grandEst},
	{"58", "Nièvre", bourgogneFrancheComte},
	{"59", "Nord", hautsDeFrance},
	{"60", "Oise", hautsDeFrance},
	{"61", "Orne", normandie},
	{"62", "Pas-de-Calais", hautsDeFrance},
	{"63", "Puy-de-Dôme", auvergneRhoneAlpes},
	{"64", "Pyrénées-Atlantiques", nouvelleAquitaine},
	{"65", "Hautes-Pyrénées", occitanie},
	{"66", "Pyrénées-Orientales", occitanie},
	{"67", "Bas-Rhin", grandEst},
	{"68", "Haut-Rhin", grandEst},
	{"69", "Rhône", auvergneRhoneAlpes},
	{"70", "Haute-Saône", bourgogneFrancheComte},
	{"71", "Saône-et-Loire", bourgogneFrancheComte},
	{"72", "Sarthe", paysDeLaLoire},
	{"73", "Savoie", auvergneRhoneAlpes},
	{"74", "Haute-Savoie", auvergneRhoneAlpes},
	{"75", "Paris", ileDeFrance},
	{"76", "Seine-Maritime", normandie},
	{"77", "Seine-et-Marne", ileDeFrance},
	{"78", "Yvelines", ileDeFrance},
	{"79", "Deux-Sèvres", nouvelleAquitaine},
	{"80", "Somme", hautsDeFrance},
	{"81", "Tarn", occitanie},
	{"82", "Tarn-et-Garonne", occitanie},
	{"83", "Var", provenceAlpesCoteDAzur},
	{"84", "Vaucluse", provenceAlpesCoteDAzur},
	{"85", "Vendée", paysDeLaLoire},
	{"86", "Vienne", nouvelleAquitaine},
	{"87", "Haute-Vienne", nouvelleAquitaine},
	{"88", "Vosges", grandEst},
	{"89", "Yonne", bourgogneFrancheComte},
	{"90", "Territoire de Belfort", bourgogneFrancheComte},
	{"91", "Essonne", ileDeFrance},
	{"92", "Hauts-de-Seine", ileDeFrance},
	{"93", "Seine-Saint-Denis", ileDeFrance},
	{"94", "Val-de-Marne", ileDeFrance},
	{"95", "Val-d'Oise", ileDeFrance},
	{"971", "Guadeloupe", guadeloupe},
	{"972", "Martinique", martinique},
	{"973", "Guyane", guyane},
	{"974", "La Réunion", laReunion},
	{"976", "Mayotte", mayotte},
}

var departementsParCode = func() map[string]Departement {
	parCode := make(map[string]Departement, len(departements))
	for _, departement := range departements {
		parCode[departement.code] = departement
	}
	return parCode
}()

var regionsParCode = func() map[string]Region {
	parCode := make(map[string]Region, len(regions))
	for _, region := range regions {
		parCode[region.code] = region
	}
	return parCode
}()

// NewRegion finds a region by its INSEE code.
func NewRegion(code string) (Region, error) {
	region, ok := regionsParCode[code]
	if !ok {
		return Region{}, ErrRegionInconnue
	}
	return region, nil
}

// NewDepartement finds a departement by its INSEE code (2A, 974…).
func NewDepartement(code string) (Departement, error) {
	departement, ok := departementsParCode[code]
	if !ok {
		return Departement{}, ErrCodeDepartementInconnu
	}
	return departement, nil
}

// DepartementsDe returns the departements of a region, in code order.
func DepartementsDe(region Region) []Departement {
	var out []Departement
	for _, departement := range departements {
		if departement.region == region {
			out = append(out, departement)
		}
	}
	return out
}

// Regions returns every region, metropolitan first, then overseas.
func Regions() []Region {
	return append([]Region(nil), regions...)
}

// Departements returns every departement, in code order.
func Departements() []Departement {
	return append([]Departement(nil), departements...)
}
