package domain

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"unicode/utf8"
)

// MotDePasse is the mot de passe de l'annonce a vendeur chooses at the
// dépôt: it will let him delete the annonce later, since he has no account.
// Only a salted pbkdf2 hash is kept, never the clear text.
type MotDePasse struct {
	sel  []byte
	hash []byte
}

const (
	motDePasseMin        = 8
	motDePasseSelOctets  = 16
	motDePasseIterations = 100_000
	motDePasseCle        = 32
)

// NewMotDePasse hashes a clear mot de passe of at least 8 characters, with a
// random salt. No other composition rule applies.
func NewMotDePasse(clair string) (MotDePasse, error) {
	if utf8.RuneCountInString(clair) < motDePasseMin {
		return MotDePasse{}, ErrMotDePasseTropCourt
	}
	sel := make([]byte, motDePasseSelOctets)
	if _, err := rand.Read(sel); err != nil {
		return MotDePasse{}, err
	}
	hash, err := hacher(clair, sel)
	if err != nil {
		return MotDePasse{}, err
	}
	return MotDePasse{sel: sel, hash: hash}, nil
}

// Correspond tells whether saisie is the mot de passe, in constant time.
func (m MotDePasse) Correspond(saisie string) bool {
	hash, err := hacher(saisie, m.sel)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(hash, m.hash) == 1
}

func hacher(clair string, sel []byte) ([]byte, error) {
	return pbkdf2.Key(sha256.New, clair, sel, motDePasseIterations, motDePasseCle)
}
