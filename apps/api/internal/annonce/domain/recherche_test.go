package domain

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func annonceAvec(t *testing.T, titre, description string) *Annonce {
	t.Helper()
	annonce, err := NewAnnonce(NewAnnonceArgs{
		Titre: titre, Description: description, Prix: 50000,
		Categorie: "loisirs", Ville: "Lyon", CodePostal: "69003",
		Pseudo: "cycliste", MotDePasse: "demo1234", Email: "cycliste@example.com", DeposeeLe: time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	return annonce
}

func criteres(t *testing.T, motCle string) CriteresRecherche {
	t.Helper()
	c, err := NewCriteresRecherche(CriteresRechercheArgs{MotCle: motCle})
	require.NoError(t, err)
	return c
}

func TestCriteresRecherche_MotCle(t *testing.T) {
	velo := annonceAvec(t, "Vélo Électrique", "Autonomie 60 km, révisé.")
	tests := []struct {
		name   string
		motCle string
		retenu bool
	}{
		{"an empty mot-cle keeps every annonce", "", true},
		{"blanks are an empty mot-cle", "   ", true},
		{"ignores case and accents", "velo electrique", true},
		{"ignores accents typed by the acheteur", "ÉLECTRIQUE", true},
		{"searches the description too", "autonomie", true},
		{"words may come from titre and description", "velo revise", true},
		{"a word may be part of a longer one", "elec", true},
		{"every word must appear", "velo cargo", false},
		{"an absent word finds nothing", "voiture", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.retenu, criteres(t, tt.motCle).Retient(velo))
		})
	}
}

func TestCriteresRecherche_KeepsOnlyActiveeAnnonces(t *testing.T) {
	velo := annonceAvec(t, "Vélo Électrique", "Autonomie 60 km, révisé.")
	for _, statut := range []Statut{EnAttenteModeration, Challengee, Refusee, Supprimee} {
		cache, err := RestoreAnnonce(uuid.NewString(), statut, NewAnnonceArgs{
			Titre: velo.Titre(), Description: velo.Description(), Prix: 50000,
			Categorie: "loisirs", Ville: "Lyon", CodePostal: "69003",
			Pseudo: "cycliste", MotDePasse: "demo1234", Email: "cycliste@example.com", DeposeeLe: time.Now(),
		})
		require.NoError(t, err)

		assert.False(t, criteres(t, "").Retient(cache), statut)
		assert.False(t, criteres(t, "velo").Retient(cache), statut)
	}
}

func TestNewPage(t *testing.T) {
	page, err := NewPage(0)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Numero(), "zero means the first page")
	assert.Equal(t, 0, page.Offset())

	page, err = NewPage(3)
	require.NoError(t, err)
	assert.Equal(t, 3, page.Numero())
	assert.Equal(t, 40, page.Offset())

	_, err = NewPage(-1)
	require.ErrorIs(t, err, ErrPageInvalide)

	page, err = NewPage(math.MaxInt)
	require.NoError(t, err)
	assert.Equal(t, math.MaxInt, page.Offset(), "a page far past the last saturates instead of overflowing")
}

func centimes(n int64) *int64 { return &n }

func annonceAPrix(t *testing.T, categorie string, prix int64) *Annonce {
	t.Helper()
	annonce, err := NewAnnonce(NewAnnonceArgs{
		Titre: "Vélo de ville", Description: "Bon état général, révisé.", Prix: prix,
		Categorie: categorie, Ville: "Lyon", CodePostal: "69003",
		Pseudo: "cycliste", MotDePasse: "demo1234", Email: "cycliste@example.com", DeposeeLe: time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	return annonce
}

func TestCriteresRecherche_Categorie(t *testing.T) {
	loisirs := annonceAPrix(t, "loisirs", 50000)
	vehicules := annonceAPrix(t, "vehicules", 50000)

	c, err := NewCriteresRecherche(CriteresRechercheArgs{Categorie: "loisirs"})
	require.NoError(t, err)

	assert.True(t, c.Retient(loisirs))
	assert.False(t, c.Retient(vehicules))
}

func TestCriteresRecherche_RejectsAnUnknownCategorie(t *testing.T) {
	_, err := NewCriteresRecherche(CriteresRechercheArgs{Categorie: "licornes"})

	require.ErrorIs(t, err, ErrCategorieInconnue)
}

func TestCriteresRecherche_FourchetteDePrix(t *testing.T) {
	tests := []struct {
		name    string
		min     *int64
		max     *int64
		prix    int64
		retenue bool
	}{
		{"no bound keeps every prix", nil, nil, 123, true},
		{"the minimum is included", centimes(1000), nil, 1000, true},
		{"below the minimum is excluded", centimes(1000), nil, 999, false},
		{"the maximum is included", nil, centimes(1000), 1000, true},
		{"above the maximum is excluded", nil, centimes(1000), 1001, false},
		{"inside both bounds is kept", centimes(1000), centimes(2000), 1500, true},
		{"above both bounds is excluded", centimes(1000), centimes(2000), 2001, false},
		{"0-0 keeps a don", centimes(0), centimes(0), 0, true},
		{"0-0 excludes a priced annonce", centimes(0), centimes(0), 1, false},
		{"equal bounds keep that exact prix", centimes(500), centimes(500), 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCriteresRecherche(CriteresRechercheArgs{PrixMin: tt.min, PrixMax: tt.max})
			require.NoError(t, err)

			assert.Equal(t, tt.retenue, c.Retient(annonceAPrix(t, "loisirs", tt.prix)))
		})
	}
}

func TestCriteresRecherche_RejectsAnInvalidFourchetteDePrix(t *testing.T) {
	tests := []struct {
		name string
		args CriteresRechercheArgs
		err  error
	}{
		{"a minimum above the maximum", CriteresRechercheArgs{PrixMin: centimes(2000), PrixMax: centimes(1000)}, ErrFourchettePrixInvalide},
		{"a negative minimum", CriteresRechercheArgs{PrixMin: centimes(-1)}, ErrPrixMinInvalide},
		{"a negative maximum", CriteresRechercheArgs{PrixMax: centimes(-1)}, ErrPrixMaxInvalide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCriteresRecherche(tt.args)

			require.ErrorIs(t, err, tt.err)
		})
	}
}

func TestCriteresRecherche_CombinesEveryCriterionWithAnd(t *testing.T) {
	c, err := NewCriteresRecherche(CriteresRechercheArgs{
		MotCle: "velo", Categorie: "loisirs", PrixMin: centimes(10000), PrixMax: centimes(60000),
	})
	require.NoError(t, err)

	assert.True(t, c.Retient(annonceAPrix(t, "loisirs", 50000)))
	assert.False(t, c.Retient(annonceAPrix(t, "vehicules", 50000)), "wrong categorie")
	assert.False(t, c.Retient(annonceAPrix(t, "loisirs", 70000)), "above the maximum")
	assert.False(t, c.Retient(annonceAvec(t, "Canapé trois places", "Bon état général.")), "no mot-cle")
}

func annonceA(t *testing.T, ville, codePostal string) *Annonce {
	t.Helper()
	annonce, err := NewAnnonce(NewAnnonceArgs{
		Titre: "Vélo de ville", Description: "Bon état général, révisé.", Prix: 50000,
		Categorie: "loisirs", Ville: ville, CodePostal: codePostal,
		Pseudo: "cycliste", MotDePasse: "demo1234", Email: "cycliste@example.com", DeposeeLe: time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	return annonce
}

func TestCriteresRecherche_RegionOrDepartement(t *testing.T) {
	saintDenis93 := annonceA(t, "Saint-Denis", "93200")
	saintDenis974 := annonceA(t, "Saint-Denis", "97400")
	ajaccio := annonceA(t, "Ajaccio", "20000")
	bastia := annonceA(t, "Bastia", "20200")
	tests := []struct {
		name    string
		args    CriteresRechercheArgs
		annonce *Annonce
		retenue bool
	}{
		{"no lieu keeps every annonce", CriteresRechercheArgs{}, saintDenis974, true},
		{"a region keeps its annonces", CriteresRechercheArgs{Region: "11"}, saintDenis93, true},
		{"a region excludes the others", CriteresRechercheArgs{Region: "11"}, saintDenis974, false},
		{"an overseas region", CriteresRechercheArgs{Region: "04"}, saintDenis974, true},
		{"a departement keeps its annonces", CriteresRechercheArgs{Departement: "93"}, saintDenis93, true},
		{"a departement excludes the others", CriteresRechercheArgs{Departement: "93"}, saintDenis974, false},
		{"the Corse region holds 2A", CriteresRechercheArgs{Region: "94"}, ajaccio, true},
		{"the Corse region holds 2B", CriteresRechercheArgs{Region: "94"}, bastia, true},
		{"2A is Corse-du-Sud", CriteresRechercheArgs{Departement: "2A"}, ajaccio, true},
		{"2A excludes Haute-Corse", CriteresRechercheArgs{Departement: "2A"}, bastia, false},
		{"2B is Haute-Corse", CriteresRechercheArgs{Departement: "2B"}, bastia, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCriteresRecherche(tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.retenue, c.Retient(tt.annonce))
		})
	}
}

func TestCriteresRecherche_VilleOuCodePostal(t *testing.T) {
	saintEtienne := annonceA(t, "Saint-Étienne", "42000")
	paris15 := annonceA(t, "Paris 15e", "75015")
	lHay := annonceA(t, "L'Haÿ-les-Roses", "94240")
	tests := []struct {
		name    string
		saisie  string
		annonce *Annonce
		retenue bool
	}{
		{"a blank keeps every annonce", "  ", saintEtienne, true},
		{"five digits match the code postal exactly", "42000", saintEtienne, true},
		{"five digits never match a prefix", "42100", saintEtienne, false},
		{"five digits are no ville", "75015", saintEtienne, false},
		{"fewer digits are compared to the ville", "4200", saintEtienne, false},
		{"the ville ignores case, accents and hyphens", "saint etienne", saintEtienne, true},
		{"the ville as the vendeur typed it", "Saint-Étienne", saintEtienne, true},
		{"surrounding spaces are dropped", "  SAINT-ETIENNE ", saintEtienne, true},
		{"apostrophes count as spaces", "l hay les roses", lHay, true},
		{"the ville must be identical", "Paris", paris15, false},
		{"the whole ville matches", "paris 15e", paris15, true},
		{"a part of the ville does not match", "saint", saintEtienne, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCriteresRecherche(CriteresRechercheArgs{VilleOuCodePostal: tt.saisie})
			require.NoError(t, err)

			assert.Equal(t, tt.retenue, c.Retient(tt.annonce))
		})
	}
}

func TestCriteresRecherche_CombinesTheLieuWithAnd(t *testing.T) {
	saintDenis93 := annonceA(t, "Saint-Denis", "93200")
	saintDenis974 := annonceA(t, "Saint-Denis", "97400")

	reunion, err := NewCriteresRecherche(CriteresRechercheArgs{Region: "04", VilleOuCodePostal: "saint denis"})
	require.NoError(t, err)
	assert.True(t, reunion.Retient(saintDenis974))
	assert.False(t, reunion.Retient(saintDenis93), "homonym in another region")

	incoherent, err := NewCriteresRecherche(CriteresRechercheArgs{Departement: "2A", VilleOuCodePostal: "93200"})
	require.NoError(t, err, "an incoherent combination is no error")
	assert.False(t, incoherent.Retient(saintDenis93))

	tout, err := NewCriteresRecherche(CriteresRechercheArgs{
		MotCle: "velo", Categorie: "loisirs", PrixMax: centimes(60000),
		Departement: "974", VilleOuCodePostal: "97400",
	})
	require.NoError(t, err)
	assert.True(t, tout.Retient(saintDenis974))
	assert.False(t, tout.Retient(saintDenis93))
}

func TestCriteresRecherche_RejectsAnInvalidLieu(t *testing.T) {
	tests := []struct {
		name string
		args CriteresRechercheArgs
		err  error
	}{
		{"an unknown region", CriteresRechercheArgs{Region: "99"}, ErrRegionInconnue},
		{"an unknown departement", CriteresRechercheArgs{Departement: "975"}, ErrCodeDepartementInconnu},
		{"Corse is no departement code", CriteresRechercheArgs{Departement: "20"}, ErrCodeDepartementInconnu},
		{"both a region and a departement", CriteresRechercheArgs{Region: "11", Departement: "93"}, ErrLieuAmbigu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCriteresRecherche(tt.args)

			require.ErrorIs(t, err, tt.err)
		})
	}
}
