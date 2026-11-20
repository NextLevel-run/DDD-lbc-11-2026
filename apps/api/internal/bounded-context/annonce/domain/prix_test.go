package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrix_KeepsTheAmountInCentimes(t *testing.T) {
	prix, err := NewPrix(125050)

	require.NoError(t, err)
	assert.Equal(t, int64(125050), prix.Centimes())
	assert.False(t, prix.EstUnDon())
}

func TestNewPrix_AcceptsZeroAsADon(t *testing.T) {
	prix, err := NewPrix(0)

	require.NoError(t, err)
	assert.True(t, prix.EstUnDon())
}

func TestNewPrix_HasNoCeiling(t *testing.T) {
	_, err := NewPrix(1_000_000_000_00)

	require.NoError(t, err)
}

func TestNewPrix_RejectsANegativeAmount(t *testing.T) {
	_, err := NewPrix(-1)

	require.ErrorIs(t, err, ErrPrixNegatif)
}
