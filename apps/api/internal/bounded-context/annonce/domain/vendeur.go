package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Vendeur is the person who deposits an annonce, without an account. Only the
// pseudo is ever shown; the email and the telephone stay private.
type Vendeur struct {
	pseudo    string
	email     string
	telephone string
}

const (
	pseudoMin = 2
	pseudoMax = 30
)

// emailFormat accepts a bare address only (no display name): the SMTP mailer
// takes nothing else.
var (
	emailFormat     = regexp.MustCompile(`^[^\s@<>]+@[^\s@<>]+\.[^\s@<>]+$`)
	telephoneFormat = regexp.MustCompile(`^0[0-9]{9}$`)
	telephoneBlanks = strings.NewReplacer(" ", "", ".", "", "-", "")
)

// NewVendeur takes an optional telephone: an empty string means none. A
// telephone is French, ten digits starting with 0; spaces, dots and dashes
// between the digits are dropped.
func NewVendeur(pseudo, email, telephone string) (Vendeur, error) {
	pseudo = strings.TrimSpace(pseudo)
	if n := utf8.RuneCountInString(pseudo); n < pseudoMin || n > pseudoMax {
		return Vendeur{}, ErrPseudoInvalide
	}
	email = strings.TrimSpace(email)
	if !emailFormat.MatchString(email) {
		return Vendeur{}, ErrEmailInvalide
	}
	telephone = telephoneBlanks.Replace(strings.TrimSpace(telephone))
	if telephone != "" && !telephoneFormat.MatchString(telephone) {
		return Vendeur{}, ErrTelephoneInvalide
	}
	return Vendeur{pseudo: pseudo, email: email, telephone: telephone}, nil
}

func (v Vendeur) Pseudo() string    { return v.pseudo }
func (v Vendeur) Email() string     { return v.email }
func (v Vendeur) Telephone() string { return v.telephone }
