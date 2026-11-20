package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMotDePasse_RefusesLessThanEightCharacters(t *testing.T) {
	_, err := NewMotDePasse("1234567")

	assert.ErrorIs(t, err, ErrMotDePasseTropCourt)
}

func TestNewMotDePasse_CountsCharactersNotBytes(t *testing.T) {
	_, err := NewMotDePasse("éééééééé")

	assert.NoError(t, err)
}

func TestNewMotDePasse_AcceptsEightCharactersWithoutOtherRule(t *testing.T) {
	_, err := NewMotDePasse("aaaaaaaa")

	assert.NoError(t, err)
}

func TestMotDePasse_CorrespondsOnlyToTheSameSaisie(t *testing.T) {
	motDePasse, err := NewMotDePasse("demo1234")
	require.NoError(t, err)

	assert.True(t, motDePasse.Correspond("demo1234"))
	assert.False(t, motDePasse.Correspond("demo12345"))
	assert.False(t, motDePasse.Correspond("DEMO1234"))
	assert.False(t, motDePasse.Correspond(""))
}

func TestMotDePasse_IsNeverKeptInClear(t *testing.T) {
	motDePasse, err := NewMotDePasse("demo1234")
	require.NoError(t, err)

	assert.NotContains(t, string(motDePasse.hash), "demo1234")
	assert.NotContains(t, string(motDePasse.sel), "demo1234")
	assert.NotEmpty(t, motDePasse.hash)
}

func TestNewMotDePasse_SaltsEachHash(t *testing.T) {
	first, _ := NewMotDePasse("demo1234")
	second, _ := NewMotDePasse("demo1234")

	assert.NotEqual(t, first.sel, second.sel)
	assert.NotEqual(t, first.hash, second.hash)
	assert.False(t, strings.Contains(string(first.hash), "demo1234"))
}
