package domain

import (
	"strings"
	"unicode/utf8"
)

// Moderateur is the pseudo a moderator types in: no account, no aggregate.
type Moderateur struct{ pseudo string }

const (
	moderateurMin = 2
	moderateurMax = 30
)

// NewModerateur trims the pseudo and checks it holds 2 to 30 characters.
func NewModerateur(pseudo string) (Moderateur, error) {
	pseudo = strings.TrimSpace(pseudo)
	if n := utf8.RuneCountInString(pseudo); n < moderateurMin || n > moderateurMax {
		return Moderateur{}, ErrModerateurInvalide
	}
	return Moderateur{pseudo: pseudo}, nil
}

func (m Moderateur) Pseudo() string { return m.pseudo }
