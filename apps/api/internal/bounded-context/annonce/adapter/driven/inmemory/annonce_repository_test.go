package inmemory

import (
	"fmt"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAnnonce(t *testing.T) *domain.Annonce {
	t.Helper()
	annonce, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: "Canapé", Description: "Canapé trois places.", Prix: 20000,
		Categorie: "maison-jardin", Ville: "Nantes", CodePostal: "44000",
		Pseudo: "nantais", MotDePasse: "demo1234", Email: "nantais@example.com", DeposeeLe: time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	return annonce
}

func TestInMemoryAnnonceRepository_LoadsWhatItSaved(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()
	annonce := newAnnonce(t)

	require.NoError(t, repo.Save(annonce))
	loaded, err := repo.GetByID(annonce.ID())

	require.NoError(t, err)
	assert.Equal(t, annonce, loaded)
}

func TestInMemoryAnnonceRepository_ReportsAnUnknownID(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()

	_, err := repo.GetByID("3f1c8a52-8f0e-4d1b-9a77-0c2b8d6f5e41")

	require.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
}

func saveAnnonce(t *testing.T, repo *InMemoryAnnonceRepository, titre string, deposeeLe time.Time, statut domain.Statut) *domain.Annonce {
	t.Helper()
	annonce, err := domain.RestoreAnnonce(uuid.NewString(), statut, domain.NewAnnonceArgs{
		Titre: titre, Description: "Une description assez longue.", Prix: 1000,
		Categorie: "autres", Ville: "Lille", CodePostal: "59000",
		Pseudo: "lillois", MotDePasse: "demo1234", Email: "lillois@example.com", DeposeeLe: deposeeLe,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Save(annonce))
	return annonce
}

func rechercher(t *testing.T, repo *InMemoryAnnonceRepository, motCle string, numero int) domain.ResultatRecherche {
	t.Helper()
	criteres, err := domain.NewCriteresRecherche(domain.CriteresRechercheArgs{MotCle: motCle})
	require.NoError(t, err)
	page, err := domain.NewPage(numero)
	require.NoError(t, err)
	resultat, err := repo.Rechercher(criteres, page)
	require.NoError(t, err)
	return resultat
}

func titres(annonces []*domain.Annonce) []string {
	out := make([]string, len(annonces))
	for i, a := range annonces {
		out[i] = a.Titre()
	}
	return out
}

var jour = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

func TestInMemoryAnnonceRepository_RechercherSortsMostRecentFirst(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()
	saveAnnonce(t, repo, "Ancienne", jour, domain.Activee)
	saveAnnonce(t, repo, "Récente", jour.Add(48*time.Hour), domain.Activee)
	saveAnnonce(t, repo, "Moyenne", jour.Add(24*time.Hour), domain.Activee)

	resultat := rechercher(t, repo, "", 1)

	assert.Equal(t, []string{"Récente", "Moyenne", "Ancienne"}, titres(resultat.Annonces))
	assert.Equal(t, 3, resultat.Total)
}

func TestInMemoryAnnonceRepository_RechercherExcludesAnnoncesNotActivee(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()
	saveAnnonce(t, repo, "Vélo visible", jour, domain.Activee)
	saveAnnonce(t, repo, "Vélo en attente", jour, domain.EnAttenteModeration)
	saveAnnonce(t, repo, "Vélo challengé", jour, domain.Challengee)
	saveAnnonce(t, repo, "Vélo refusé", jour, domain.Refusee)

	resultat := rechercher(t, repo, "velo", 1)

	assert.Equal(t, []string{"Vélo visible"}, titres(resultat.Annonces))
	assert.Equal(t, 1, resultat.Total)
}

func TestInMemoryAnnonceRepository_RechercherAppliesTheMotCle(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()
	saveAnnonce(t, repo, "Vélo Électrique", jour, domain.Activee)
	saveAnnonce(t, repo, "Vélo de course", jour.Add(time.Hour), domain.Activee)
	saveAnnonce(t, repo, "Trottinette électrique", jour.Add(2*time.Hour), domain.Activee)

	resultat := rechercher(t, repo, "velo electrique", 1)

	assert.Equal(t, []string{"Vélo Électrique"}, titres(resultat.Annonces))
	assert.Equal(t, 1, resultat.Total)
}

func TestInMemoryAnnonceRepository_RechercherPaginatesByTwenty(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()
	for i := range 45 {
		saveAnnonce(t, repo, fmt.Sprintf("Annonce %02d", i), jour.Add(time.Duration(i)*time.Hour), domain.Activee)
	}

	premiere := rechercher(t, repo, "", 1)
	troisieme := rechercher(t, repo, "", 3)
	auDela := rechercher(t, repo, "", 4)

	require.Len(t, premiere.Annonces, 20)
	assert.Equal(t, "Annonce 44", premiere.Annonces[0].Titre())
	assert.Equal(t, "Annonce 25", premiere.Annonces[19].Titre())
	assert.Equal(t, []string{"Annonce 04", "Annonce 03", "Annonce 02", "Annonce 01", "Annonce 00"}, titres(troisieme.Annonces))
	assert.Empty(t, auDela.Annonces)
	for _, r := range []domain.ResultatRecherche{premiere, troisieme, auDela} {
		assert.Equal(t, 45, r.Total)
	}
}

func saveAnnonceA(t *testing.T, repo *InMemoryAnnonceRepository, titre, categorie string, prix int64, deposeeLe time.Time) {
	t.Helper()
	annonce, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: titre, Description: "Une description assez longue.", Prix: prix,
		Categorie: categorie, Ville: "Lille", CodePostal: "59000",
		Pseudo: "lillois", MotDePasse: "demo1234", Email: "lillois@example.com", DeposeeLe: deposeeLe,
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, repo.Save(annonce))
}

func centimes(n int64) *int64 { return &n }

func TestInMemoryAnnonceRepository_RechercherCombinesCategorieAndPrix(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()
	saveAnnonceA(t, repo, "Vélo donné", "loisirs", 0, jour)
	saveAnnonceA(t, repo, "Vélo de ville", "loisirs", 15000, jour.Add(time.Hour))
	saveAnnonceA(t, repo, "Vélo de course", "loisirs", 90000, jour.Add(2*time.Hour))
	saveAnnonceA(t, repo, "Vélo cargo", "vehicules", 15000, jour.Add(3*time.Hour))
	saveAnnonceA(t, repo, "Canapé donné", "maison-jardin", 0, jour.Add(4*time.Hour))

	tests := []struct {
		name   string
		args   domain.CriteresRechercheArgs
		titres []string
	}{
		{"categorie alone", domain.CriteresRechercheArgs{Categorie: "loisirs"},
			[]string{"Vélo de course", "Vélo de ville", "Vélo donné"}},
		{"minimum alone, included", domain.CriteresRechercheArgs{PrixMin: centimes(15000)},
			[]string{"Vélo cargo", "Vélo de course", "Vélo de ville"}},
		{"maximum alone, included", domain.CriteresRechercheArgs{PrixMax: centimes(15000)},
			[]string{"Canapé donné", "Vélo cargo", "Vélo de ville", "Vélo donné"}},
		{"0-0 finds only the dons", domain.CriteresRechercheArgs{PrixMin: centimes(0), PrixMax: centimes(0)},
			[]string{"Canapé donné", "Vélo donné"}},
		{"categorie and fourchette", domain.CriteresRechercheArgs{Categorie: "loisirs", PrixMin: centimes(1), PrixMax: centimes(20000)},
			[]string{"Vélo de ville"}},
		{"mot-cle, categorie and fourchette", domain.CriteresRechercheArgs{MotCle: "velo", Categorie: "loisirs", PrixMax: centimes(0)},
			[]string{"Vélo donné"}},
		{"mot-cle and categorie", domain.CriteresRechercheArgs{MotCle: "donne", Categorie: "maison-jardin"},
			[]string{"Canapé donné"}},
		{"nothing in the categorie and fourchette", domain.CriteresRechercheArgs{Categorie: "vehicules", PrixMin: centimes(20000)},
			[]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteres, err := domain.NewCriteresRecherche(tt.args)
			require.NoError(t, err)
			page, err := domain.NewPage(1)
			require.NoError(t, err)

			resultat, err := repo.Rechercher(criteres, page)

			require.NoError(t, err)
			assert.Equal(t, tt.titres, titres(resultat.Annonces))
			assert.Equal(t, len(tt.titres), resultat.Total)
		})
	}
}

func saveAnnonceAVille(t *testing.T, repo *InMemoryAnnonceRepository, titre, ville, codePostal string, deposeeLe time.Time) {
	t.Helper()
	annonce, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: titre, Description: "Une description assez longue.", Prix: 1000,
		Categorie: "autres", Ville: ville, CodePostal: codePostal,
		Pseudo: "voisin", MotDePasse: "demo1234", Email: "voisin@example.com", DeposeeLe: deposeeLe,
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, repo.Save(annonce))
}

func TestInMemoryAnnonceRepository_RechercherFiltersByLieu(t *testing.T) {
	repo := NewInMemoryAnnonceRepository()
	saveAnnonceAVille(t, repo, "Lampe dionysienne", "Saint-Denis", "93200", jour)
	saveAnnonceAVille(t, repo, "Lampe réunionnaise", "Saint-Denis", "97400", jour.Add(time.Hour))
	saveAnnonceAVille(t, repo, "Lampe ajaccienne", "Ajaccio", "20000", jour.Add(2*time.Hour))
	saveAnnonceAVille(t, repo, "Lampe bastiaise", "Bastia", "20200", jour.Add(3*time.Hour))
	saveAnnonceAVille(t, repo, "Lampe stéphanoise", "Saint-Étienne", "42000", jour.Add(4*time.Hour))
	saveAnnonceAVille(t, repo, "Lampe parisienne", "Paris 15e", "75015", jour.Add(5*time.Hour))

	tests := []struct {
		name   string
		args   domain.CriteresRechercheArgs
		titres []string
	}{
		{"the homonyms share a ville", domain.CriteresRechercheArgs{VilleOuCodePostal: "saint-denis"},
			[]string{"Lampe réunionnaise", "Lampe dionysienne"}},
		{"Saint-Denis in La Réunion", domain.CriteresRechercheArgs{Region: "04", VilleOuCodePostal: "Saint-Denis"},
			[]string{"Lampe réunionnaise"}},
		{"Saint-Denis in Seine-Saint-Denis", domain.CriteresRechercheArgs{Departement: "93", VilleOuCodePostal: "saint denis"},
			[]string{"Lampe dionysienne"}},
		{"a code postal tells the homonyms apart", domain.CriteresRechercheArgs{VilleOuCodePostal: "97400"},
			[]string{"Lampe réunionnaise"}},
		{"the Corse region holds both departements", domain.CriteresRechercheArgs{Region: "94"},
			[]string{"Lampe bastiaise", "Lampe ajaccienne"}},
		{"Corse-du-Sud alone", domain.CriteresRechercheArgs{Departement: "2A"},
			[]string{"Lampe ajaccienne"}},
		{"Haute-Corse alone", domain.CriteresRechercheArgs{Departement: "2B"},
			[]string{"Lampe bastiaise"}},
		{"the ville without accent nor hyphen", domain.CriteresRechercheArgs{VilleOuCodePostal: "saint etienne"},
			[]string{"Lampe stéphanoise"}},
		{"Paris does not find Paris 15e", domain.CriteresRechercheArgs{VilleOuCodePostal: "Paris"},
			[]string{}},
		{"lieu and mot-cle", domain.CriteresRechercheArgs{MotCle: "lampe", Region: "84"},
			[]string{"Lampe stéphanoise"}},
		{"an incoherent lieu finds nothing", domain.CriteresRechercheArgs{Departement: "2A", VilleOuCodePostal: "Saint-Denis"},
			[]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteres, err := domain.NewCriteresRecherche(tt.args)
			require.NoError(t, err)
			page, err := domain.NewPage(1)
			require.NoError(t, err)

			resultat, err := repo.Rechercher(criteres, page)

			require.NoError(t, err)
			assert.Equal(t, tt.titres, titres(resultat.Annonces))
			assert.Equal(t, len(tt.titres), resultat.Total)
		})
	}
}
