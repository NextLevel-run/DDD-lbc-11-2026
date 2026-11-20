package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalisation_DeducesDepartementAndRegionFromCodePostal(t *testing.T) {
	tests := []struct {
		name        string
		codePostal  string
		departement string
		region      string
	}{
		{"first metropolitan departement", "01000", "01", "Auvergne-Rhône-Alpes"},
		{"Paris", "75015", "75", "Île-de-France"},
		{"last metropolitan departement", "95000", "95", "Île-de-France"},
		{"Corse-du-Sud from 200", "20000", "2A", "Corse"},
		{"Corse-du-Sud from 201", "20137", "2A", "Corse"},
		{"Haute-Corse from 202", "20200", "2B", "Corse"},
		{"Haute-Corse from 206", "20620", "2B", "Corse"},
		{"Guadeloupe", "97100", "971", "Guadeloupe"},
		{"Martinique", "97200", "972", "Martinique"},
		{"Guyane", "97300", "973", "Guyane"},
		{"La Reunion", "97400", "974", "La Réunion"},
		{"Mayotte", "97600", "976", "Mayotte"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localisation, err := NewLocalisation("Ville", tt.codePostal)

			require.NoError(t, err)
			assert.Equal(t, tt.codePostal, localisation.CodePostal())
			assert.Equal(t, tt.departement, localisation.Departement().Code())
			assert.Equal(t, tt.region, localisation.Region().Nom())
		})
	}
}

func TestNewLocalisation_NamesTheDepartement(t *testing.T) {
	localisation, err := NewLocalisation("Ajaccio", "20000")

	require.NoError(t, err)
	assert.Equal(t, "Ajaccio", localisation.Ville())
	assert.Equal(t, "Corse-du-Sud", localisation.Departement().Nom())
	assert.Equal(t, "94", localisation.Region().Code())
}

func TestNewLocalisation_RejectsACodePostalThatIsNotFiveDigits(t *testing.T) {
	for _, codePostal := range []string{"", "7500", "750011", "75a01", " 7500"} {
		t.Run(codePostal, func(t *testing.T) {
			_, err := NewLocalisation("Paris", codePostal)

			require.ErrorIs(t, err, ErrCodePostalInvalide)
		})
	}
}

func TestNewLocalisation_RejectsAPrefixThatDesignatesNoDepartement(t *testing.T) {
	for _, codePostal := range []string{"00100", "20700", "20900", "96000", "97000", "97500", "97700", "98000", "99000"} {
		t.Run(codePostal, func(t *testing.T) {
			_, err := NewLocalisation("Ville", codePostal)

			require.ErrorIs(t, err, ErrDepartementInconnu)
		})
	}
}

func TestNewLocalisation_RejectsAnEmptyVille(t *testing.T) {
	_, err := NewLocalisation("   ", "75001")

	require.ErrorIs(t, err, ErrVilleVide)
}

func TestNewLocalisation_TrimsTheVille(t *testing.T) {
	localisation, err := NewLocalisation("  Lyon ", "69001")

	require.NoError(t, err)
	assert.Equal(t, "Lyon", localisation.Ville())
}

func TestDepartements_CoverEveryDepartementOnce(t *testing.T) {
	departements := Departements()

	// 94 metropolitan codes from 01 to 95 minus 20, plus 2A, 2B and five overseas.
	assert.Len(t, departements, 101)
	seen := map[string]bool{}
	for _, departement := range departements {
		assert.False(t, seen[departement.Code()], "duplicate %s", departement.Code())
		seen[departement.Code()] = true
		assert.NotEmpty(t, departement.Nom())
		assert.NotEmpty(t, departement.Region().Nom())
	}
	assert.Len(t, Regions(), 18)
}

func TestNewRegion_FindsARegionByItsCode(t *testing.T) {
	region, err := NewRegion("94")
	require.NoError(t, err)
	assert.Equal(t, "Corse", region.Nom())

	_, err = NewRegion("99")
	require.ErrorIs(t, err, ErrRegionInconnue)
}

func TestNewDepartement_FindsADepartementByItsCode(t *testing.T) {
	departement, err := NewDepartement("2B")
	require.NoError(t, err)
	assert.Equal(t, "Haute-Corse", departement.Nom())
	assert.Equal(t, "Corse", departement.Region().Nom())

	_, err = NewDepartement("975")
	require.ErrorIs(t, err, ErrCodeDepartementInconnu)
}

func TestDepartementsDe_ListsTheDepartementsOfARegion(t *testing.T) {
	var codes []string
	for _, departement := range DepartementsDe(corse) {
		codes = append(codes, departement.Code())
	}
	assert.Equal(t, []string{"2A", "2B"}, codes)
}
