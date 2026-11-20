package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListerRegions_ListsEveryRegionWithItsDepartements(t *testing.T) {
	regions := BuildListerRegionsQuery()()

	require.Len(t, regions, 18)
	assert.Equal(t, "84", regions[0].Code)
	assert.Equal(t, "Auvergne-Rhône-Alpes", regions[0].Nom)
	departements := 0
	for _, region := range regions {
		require.NotEmpty(t, region.Departements, region.Nom)
		departements += len(region.Departements)
		if region.Code == "94" {
			assert.Equal(t, []DepartementView{
				{Code: "2A", Nom: "Corse-du-Sud"},
				{Code: "2B", Nom: "Haute-Corse"},
			}, region.Departements)
		}
	}
	assert.Equal(t, 101, departements)
}
