package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVendeur_AcceptsAValidVendeur(t *testing.T) {
	vendeur, err := NewVendeur("  cycliste69 ", " cycliste@example.com ", "06 12 34 56 78")

	require.NoError(t, err)
	assert.Equal(t, "cycliste69", vendeur.Pseudo())
	assert.Equal(t, "cycliste@example.com", vendeur.Email())
	assert.Equal(t, "0612345678", vendeur.Telephone())
}

func TestNewVendeur_TelephoneIsOptional(t *testing.T) {
	vendeur, err := NewVendeur("cycliste69", "cycliste@example.com", "  ")

	require.NoError(t, err)
	assert.Empty(t, vendeur.Telephone())
}

func TestNewVendeur_AcceptsPseudoLengthBounds(t *testing.T) {
	for _, pseudo := range []string{"ab", strings.Repeat("é", 30)} {
		_, err := NewVendeur(pseudo, "a@example.com", "")

		assert.NoError(t, err, pseudo)
	}
}

func TestNewVendeur_RejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name      string
		pseudo    string
		email     string
		telephone string
		wantErr   error
	}{
		{"pseudo too short", "a", "a@example.com", "", ErrPseudoInvalide},
		{"pseudo blank", "   ", "a@example.com", "", ErrPseudoInvalide},
		{"pseudo too long", strings.Repeat("a", 31), "a@example.com", "", ErrPseudoInvalide},
		{"email empty", "cycliste", "", "", ErrEmailInvalide},
		{"email without at", "cycliste", "cycliste.example.com", "", ErrEmailInvalide},
		{"email without domain dot", "cycliste", "cycliste@example", "", ErrEmailInvalide},
		{"email with display name", "cycliste", "Cycliste <c@example.com>", "", ErrEmailInvalide},
		{"telephone too short", "cycliste", "a@example.com", "061234567", ErrTelephoneInvalide},
		{"telephone not starting with 0", "cycliste", "a@example.com", "6123456789", ErrTelephoneInvalide},
		{"telephone international", "cycliste", "a@example.com", "+33612345678", ErrTelephoneInvalide},
		{"telephone with letters", "cycliste", "a@example.com", "06123456ab", ErrTelephoneInvalide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewVendeur(tt.pseudo, tt.email, tt.telephone)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
