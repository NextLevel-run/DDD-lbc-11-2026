package domain

import (
	"strings"
	"unicode/utf8"
)

// Motif is why a moderator challenges an annonce: what the vendeur must fix.
type Motif struct{ texte string }

const (
	motifMin = 10
	motifMax = 1000
)

// NewMotif trims the text and checks it holds 10 to 1 000 characters.
func NewMotif(texte string) (Motif, error) {
	texte = strings.TrimSpace(texte)
	if n := utf8.RuneCountInString(texte); n < motifMin || n > motifMax {
		return Motif{}, ErrMotifInvalide
	}
	return Motif{texte: texte}, nil
}

func (m Motif) Texte() string { return m.texte }
