package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRaisonSuppression_AcceptsTheClosedList(t *testing.T) {
	for _, code := range []string{"vendue-sur-le-site", "vendue-ailleurs", "plus-envie-de-vendre", "autre"} {
		raison, err := NewRaisonSuppression(code)

		require.NoError(t, err)
		assert.Equal(t, code, raison.Code())
	}
}

func TestNewRaisonSuppression_RefusesAnEmptyOrUnknownCode(t *testing.T) {
	for _, code := range []string{"", "inconnue", "Autre"} {
		_, err := NewRaisonSuppression(code)

		assert.ErrorIs(t, err, ErrRaisonSuppressionInvalide)
	}
}
