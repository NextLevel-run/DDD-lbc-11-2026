package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormulaireContact_TrimsEveryField(t *testing.T) {
	formulaire, err := NewFormulaireContact(" acheteur42 ", " acheteur@example.com ", "  Toujours disponible ?  ")

	require.NoError(t, err)
	assert.Equal(t, "acheteur42", formulaire.Pseudo())
	assert.Equal(t, "acheteur@example.com", formulaire.Email())
	assert.Equal(t, "Toujours disponible ?", formulaire.Message())
}

func TestNewFormulaireContact_AcceptsMessageLengthBounds(t *testing.T) {
	for _, message := range []string{strings.Repeat("é", 10), strings.Repeat("é", 2000)} {
		_, err := NewFormulaireContact("acheteur42", "a@example.com", message)

		assert.NoError(t, err, len(message))
	}
}

func TestNewFormulaireContact_RejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name, pseudo, email, message string
		want                         error
	}{
		{"pseudo too short", "a", "a@example.com", "Toujours disponible ?", ErrPseudoInvalide},
		{"pseudo too long", strings.Repeat("a", 31), "a@example.com", "Toujours disponible ?", ErrPseudoInvalide},
		{"email with a display name", "acheteur42", "Bob <a@example.com>", "Toujours disponible ?", ErrEmailInvalide},
		{"email missing", "acheteur42", "", "Toujours disponible ?", ErrEmailInvalide},
		{"message too short once trimmed", "acheteur42", "a@example.com", "   court    ", ErrMessageInvalide},
		{"message too long", "acheteur42", "a@example.com", strings.Repeat("a", 2001), ErrMessageInvalide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFormulaireContact(tt.pseudo, tt.email, tt.message)

			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestNewFormulaireContactSoumisEvent_CarriesTheAcheteurAndTheDate(t *testing.T) {
	formulaire, err := NewFormulaireContact("acheteur42", "acheteur@example.com", "Toujours disponible ?")
	require.NoError(t, err)
	occurredAt := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)

	event := NewFormulaireContactSoumisEvent(occurredAt, "abc", formulaire)

	assert.Equal(t, "FormulaireContactSoumis", event.EventType())
	assert.NotEmpty(t, event.ID())
	assert.Equal(t, "abc", event.AnnonceID)
	assert.Equal(t, "acheteur42", event.Pseudo)
	assert.Equal(t, "acheteur@example.com", event.Email)
	assert.Equal(t, "Toujours disponible ?", event.Message)
	assert.Equal(t, occurredAt, event.SoumisLe)
}
