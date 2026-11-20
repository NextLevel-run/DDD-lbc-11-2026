package domain

import (
	"regexp"
	"strings"
)

// Localisation is where the good is: the ville and the code postal the vendeur
// typed, and the departement (with its region) deduced from the code postal.
type Localisation struct {
	ville       string
	codePostal  string
	departement Departement
}

var codePostalFormat = regexp.MustCompile(`^[0-9]{5}$`)

func NewLocalisation(ville, codePostal string) (Localisation, error) {
	ville = strings.TrimSpace(ville)
	if ville == "" {
		return Localisation{}, ErrVilleVide
	}
	if !codePostalFormat.MatchString(codePostal) {
		return Localisation{}, ErrCodePostalInvalide
	}
	departement, ok := departementDuCodePostal(codePostal)
	if !ok {
		return Localisation{}, ErrDepartementInconnu
	}
	return Localisation{ville: ville, codePostal: codePostal, departement: departement}, nil
}

func (l Localisation) Ville() string            { return l.ville }
func (l Localisation) CodePostal() string       { return l.codePostal }
func (l Localisation) Departement() Departement { return l.departement }
func (l Localisation) Region() Region           { return l.departement.region }

// departementDuCodePostal reads the departement off the code postal prefix:
// two digits in metropolitan France, the third digit splitting Corse (200-201
// is 2A, 202-206 is 2B), three digits overseas (971 to 976).
func departementDuCodePostal(codePostal string) (Departement, bool) {
	code := codePostal[:2]
	switch code {
	case "20":
		switch codePostal[2] {
		case '0', '1':
			code = "2A"
		case '2', '3', '4', '5', '6':
			code = "2B"
		default:
			return Departement{}, false
		}
	case "97":
		code = codePostal[:3]
	}
	departement, ok := departementsParCode[code]
	return departement, ok
}
