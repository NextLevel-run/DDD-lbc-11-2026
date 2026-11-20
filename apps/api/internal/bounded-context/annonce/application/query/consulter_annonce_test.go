package query

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deposeeLe = time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)

func annonceArgs() domain.NewAnnonceArgs {
	return domain.NewAnnonceArgs{
		Titre:       "Peugeot 208",
		Description: "Première main, 45 000 km.",
		Prix:        1250000,
		Categorie:   "vehicules",
		Ville:       "Ajaccio",
		CodePostal:  "20000",
		Pseudo:      "corsica",
		MotDePasse:  "demo1234",
		Email:       "corsica@example.com",
		Telephone:   "0612345678",
		DeposeeLe:   deposeeLe,
	}
}

func setupConsulterAnnonce(t *testing.T) (*inmemory.InMemoryAnnonceRepository, ConsulterAnnonceQuery) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	return repo, BuildConsulterAnnonceQuery(repo)
}

func TestConsulterAnnonce_ReturnsTheDetailOfAnActiveeAnnonce(t *testing.T) {
	repo, consulter := setupConsulterAnnonce(t)
	annonce, err := domain.NewAnnonce(annonceArgs())
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, repo.Save(annonce))

	view, err := consulter(annonce.ID())

	require.NoError(t, err)
	assert.Equal(t, AnnonceView{
		ID:          annonce.ID(),
		Titre:       "Peugeot 208",
		Description: "Première main, 45 000 km.",
		Prix:        1250000,
		Categorie:   CategorieView{Code: "vehicules", Libelle: "Véhicules"},
		Localisation: LocalisationView{
			Ville:       "Ajaccio",
			CodePostal:  "20000",
			Departement: DepartementView{Code: "2A", Nom: "Corse-du-Sud"},
			Region:      RegionView{Code: "94", Nom: "Corse"},
		},
		Vendeur:   VendeurView{Pseudo: "corsica"},
		Images:    []string{},
		DeposeeLe: deposeeLe,
	}, *view)
}

func TestConsulterAnnonce_RejectsAMalformedID(t *testing.T) {
	_, consulter := setupConsulterAnnonce(t)

	view, err := consulter("42")

	require.ErrorIs(t, err, domain.ErrIDAnnonceInvalide)
	assert.Nil(t, view)
}

func TestConsulterAnnonce_ReportsAnUnknownAnnonce(t *testing.T) {
	_, consulter := setupConsulterAnnonce(t)

	view, err := consulter(uuid.NewString())

	require.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
	assert.Nil(t, view)
}

func TestConsulterAnnonce_HidesAnAnnonceThatIsNotActivee(t *testing.T) {
	for _, statut := range []domain.Statut{domain.EnAttenteModeration, domain.Challengee, domain.Refusee, domain.Supprimee} {
		repo, consulter := setupConsulterAnnonce(t)
		annonce, err := domain.RestoreAnnonce(uuid.NewString(), statut, annonceArgs())
		require.NoError(t, err)
		require.NoError(t, repo.Save(annonce))

		view, err := consulter(annonce.ID())

		require.ErrorIs(t, err, domain.ErrAnnonceIntrouvable, statut)
		assert.Nil(t, view)
	}
}
