package domain

import (
	"strings"
	"unicode/utf8"
)

// FormulaireContact is what an acheteur writes to the vendeur of an annonce:
// their pseudo and email, under the same rules as the vendeur's, and a
// message. Nothing keeps it: it is emailed, then carried by an event.
type FormulaireContact struct {
	pseudo  string
	email   string
	message string
}

const (
	messageMin = 10
	messageMax = 2000
)

// NewFormulaireContact trims every field; the message counts runes.
func NewFormulaireContact(pseudo, email, message string) (FormulaireContact, error) {
	pseudo, err := validerPseudo(pseudo)
	if err != nil {
		return FormulaireContact{}, err
	}
	email, err = validerEmail(email)
	if err != nil {
		return FormulaireContact{}, err
	}
	message = strings.TrimSpace(message)
	if n := utf8.RuneCountInString(message); n < messageMin || n > messageMax {
		return FormulaireContact{}, ErrMessageInvalide
	}
	return FormulaireContact{pseudo: pseudo, email: email, message: message}, nil
}

func (f FormulaireContact) Pseudo() string  { return f.pseudo }
func (f FormulaireContact) Email() string   { return f.email }
func (f FormulaireContact) Message() string { return f.message }
