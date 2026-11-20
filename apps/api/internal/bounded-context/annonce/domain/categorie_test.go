package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategories_AreTheClosedListOfEight(t *testing.T) {
	libelles := []string{}
	for _, categorie := range Categories() {
		libelles = append(libelles, categorie.Libelle())
	}

	assert.Equal(t, []string{
		"Véhicules", "Immobilier", "Multimédia", "Maison & Jardin",
		"Mode", "Loisirs", "Emploi & Services", "Autres",
	}, libelles)
}

func TestNewCategorie_FindsACategorieByItsCode(t *testing.T) {
	categorie, err := NewCategorie("maison-jardin")

	require.NoError(t, err)
	assert.Equal(t, CategorieMaisonJardin, categorie)
	assert.Equal(t, "maison-jardin", categorie.Code())
	assert.Equal(t, "Maison & Jardin", categorie.Libelle())
}

func TestNewCategorie_RejectsAnUnknownCode(t *testing.T) {
	for _, code := range []string{"", "Vehicules", "jardin"} {
		t.Run(code, func(t *testing.T) {
			_, err := NewCategorie(code)

			require.ErrorIs(t, err, ErrCategorieInconnue)
		})
	}
}
