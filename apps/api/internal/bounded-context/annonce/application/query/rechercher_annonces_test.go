package query

import (
	"fmt"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRechercherAnnonces(t *testing.T) (*inmemory.InMemoryAnnonceRepository, RechercherAnnoncesQuery) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	return repo, BuildRechercherAnnoncesQuery(repo)
}

func saveAnnonce(t *testing.T, repo *inmemory.InMemoryAnnonceRepository, statut domain.Statut, edit func(*domain.NewAnnonceArgs)) *domain.Annonce {
	t.Helper()
	args := annonceArgs()
	edit(&args)
	annonce, err := domain.RestoreAnnonce(uuid.NewString(), statut, args)
	require.NoError(t, err)
	require.NoError(t, repo.Save(annonce))
	return annonce
}

func TestRechercherAnnonces_ReturnsAPageOfSummaries(t *testing.T) {
	repo, rechercher := setupRechercherAnnonces(t)
	annonce := saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) {})

	view, err := rechercher(RechercherAnnoncesArgs{})

	require.NoError(t, err)
	assert.Equal(t, RechercherAnnoncesView{
		Annonces: []AnnonceResumeView{{
			ID:        annonce.ID(),
			Titre:     "Peugeot 208",
			Prix:      1250000,
			Ville:     "Ajaccio",
			Images:    []string{},
			DeposeeLe: deposeeLe,
		}},
		Total:   1,
		Page:    1,
		ParPage: 20,
	}, view)
}

func TestRechercherAnnonces_FindsEveryWordIgnoringCaseAndAccents(t *testing.T) {
	repo, rechercher := setupRechercherAnnonces(t)
	velo := saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) { a.Titre = "Vélo Électrique" })
	saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) { a.Titre = "Vélo de course" })
	saveAnnonce(t, repo, domain.Statut(""), func(a *domain.NewAnnonceArgs) { a.Titre = "Vélo électrique caché" })

	view, err := rechercher(RechercherAnnoncesArgs{MotCle: "velo electrique"})

	require.NoError(t, err)
	require.Len(t, view.Annonces, 1)
	assert.Equal(t, velo.ID(), view.Annonces[0].ID)
	assert.Equal(t, 1, view.Total)
}

func TestRechercherAnnonces_ServesTheRequestedPageMostRecentFirst(t *testing.T) {
	repo, rechercher := setupRechercherAnnonces(t)
	for i := range 25 {
		saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) {
			a.Titre = fmt.Sprintf("Annonce %02d", i)
			a.DeposeeLe = deposeeLe.Add(time.Duration(i) * time.Minute)
		})
	}

	view, err := rechercher(RechercherAnnoncesArgs{Page: 2})

	require.NoError(t, err)
	assert.Equal(t, 2, view.Page)
	assert.Equal(t, 25, view.Total)
	require.Len(t, view.Annonces, 5)
	assert.Equal(t, "Annonce 04", view.Annonces[0].Titre)
	assert.Equal(t, "Annonce 00", view.Annonces[4].Titre)
}

func TestRechercherAnnonces_AnswersAnEmptyListWhenNothingMatches(t *testing.T) {
	_, rechercher := setupRechercherAnnonces(t)

	view, err := rechercher(RechercherAnnoncesArgs{MotCle: "licorne"})

	require.NoError(t, err)
	assert.NotNil(t, view.Annonces)
	assert.Empty(t, view.Annonces)
	assert.Equal(t, 0, view.Total)
}

func TestRechercherAnnonces_RejectsANegativePage(t *testing.T) {
	_, rechercher := setupRechercherAnnonces(t)

	_, err := rechercher(RechercherAnnoncesArgs{Page: -1})

	require.ErrorIs(t, err, domain.ErrPageInvalide)
}

func centimes(n int64) *int64 { return &n }

func TestRechercherAnnonces_FiltersByCategorieAndFourchetteDePrix(t *testing.T) {
	repo, rechercher := setupRechercherAnnonces(t)
	retenue := saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) { a.Prix = 1000000 })
	saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) { a.Prix = 2000000 })
	saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) { a.Categorie = "loisirs"; a.Prix = 1000000 })

	view, err := rechercher(RechercherAnnoncesArgs{
		Categorie: "vehicules", PrixMin: centimes(500000), PrixMax: centimes(1500000),
	})

	require.NoError(t, err)
	require.Len(t, view.Annonces, 1)
	assert.Equal(t, retenue.ID(), view.Annonces[0].ID)
	assert.Equal(t, 1, view.Total)
}

func TestRechercherAnnonces_RejectsInvalidCriteria(t *testing.T) {
	_, rechercher := setupRechercherAnnonces(t)

	_, err := rechercher(RechercherAnnoncesArgs{Categorie: "licornes"})
	require.ErrorIs(t, err, domain.ErrCategorieInconnue)

	_, err = rechercher(RechercherAnnoncesArgs{PrixMin: centimes(2000), PrixMax: centimes(1000)})
	require.ErrorIs(t, err, domain.ErrFourchettePrixInvalide)
}

func TestRechercherAnnonces_FiltersByLieu(t *testing.T) {
	repo, rechercher := setupRechercherAnnonces(t)
	reunion := saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) {
		a.Ville, a.CodePostal = "Saint-Denis", "97400"
	})
	saveAnnonce(t, repo, domain.Activee, func(a *domain.NewAnnonceArgs) {
		a.Ville, a.CodePostal = "Saint-Denis", "93200"
	})

	view, err := rechercher(RechercherAnnoncesArgs{Region: "04", VilleOuCodePostal: "saint denis"})

	require.NoError(t, err)
	require.Len(t, view.Annonces, 1)
	assert.Equal(t, reunion.ID(), view.Annonces[0].ID)
}

func TestRechercherAnnonces_RejectsARegionAndADepartementTogether(t *testing.T) {
	_, rechercher := setupRechercherAnnonces(t)

	_, err := rechercher(RechercherAnnoncesArgs{Region: "04", Departement: "974"})

	require.ErrorIs(t, err, domain.ErrLieuAmbigu)
}
