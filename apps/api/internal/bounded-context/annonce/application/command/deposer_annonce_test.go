package command

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func setupDeposerAnnonce(t *testing.T) (DeposerAnnonceCommand, *inmemory.InMemoryAnnonceRepository, *eventbustesting.EventCollector) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, bus.Subscribe(contracts.AnnonceDeposeeEventType, collector.EventHandler()))
	clock := func() time.Time { return maintenant }
	return BuildDeposerAnnonceCommand(repo, bus, clock), repo, collector
}

func validCommandArgs() DeposerAnnonceCommandArgs {
	return DeposerAnnonceCommandArgs{
		Titre:       "Vélo de course",
		Description: "Cadre carbone, très bon état.",
		Prix:        45000,
		Categorie:   "loisirs",
		Ville:       "Lyon",
		CodePostal:  "69003",
		Pseudo:      "cycliste69",
		MotDePasse:  "demo1234",
		Email:       "cycliste@example.com",
		Telephone:   "",
	}
}

func TestDeposerAnnonce_PersistsAnAnnonceEnAttenteModerationDatedNow(t *testing.T) {
	deposerAnnonce, repo, _ := setupDeposerAnnonce(t)

	id, err := deposerAnnonce(validCommandArgs())

	require.NoError(t, err)
	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, domain.EnAttenteModeration, stored.Statut())
	assert.False(t, stored.EstActivee())
	assert.Equal(t, maintenant, stored.DeposeeLe())
	assert.Equal(t, "Vélo de course", stored.Titre())
	assert.Equal(t, int64(45000), stored.Prix().Centimes())
	assert.Equal(t, domain.CategorieLoisirs, stored.Categorie())
	assert.Equal(t, "69003", stored.Localisation().CodePostal())
	assert.Equal(t, "cycliste69", stored.Vendeur().Pseudo())
	assert.Equal(t, "cycliste@example.com", stored.Vendeur().Email())
}

func TestDeposerAnnonce_PublishesAnnonceDeposee(t *testing.T) {
	deposerAnnonce, _, collector := setupDeposerAnnonce(t)

	id, err := deposerAnnonce(validCommandArgs())

	require.NoError(t, err)
	events := collector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, contracts.AnnonceDeposeeEvent{
		Annonce: contracts.InstantaneAnnonce{
			AnnonceID:   id,
			Titre:       "Vélo de course",
			Description: "Cadre carbone, très bon état.",
			Prix:        45000,
			Categorie:   "loisirs",
			Ville:       "Lyon",
			CodePostal:  "69003",
			Pseudo:      "cycliste69",
			DeposeeLe:   maintenant,
		},
		Date: maintenant,
	}, events[0])
}

func TestDeposerAnnonce_AcceptsADon(t *testing.T) {
	deposerAnnonce, repo, _ := setupDeposerAnnonce(t)
	args := validCommandArgs()
	args.Prix = 0

	id, err := deposerAnnonce(args)

	require.NoError(t, err)
	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.True(t, stored.Prix().EstUnDon())
}

func TestDeposerAnnonce_RejectsAnInvalidAnnonceWithoutPublishing(t *testing.T) {
	deposerAnnonce, _, collector := setupDeposerAnnonce(t)
	args := validCommandArgs()
	args.Titre = "ab"

	id, err := deposerAnnonce(args)

	require.ErrorIs(t, err, domain.ErrTitreInvalide)
	assert.Empty(t, id)
	assert.Empty(t, collector.GetEvents())
}

func TestDeposerAnnonce_StoresTheMotDePasseHashed(t *testing.T) {
	deposerAnnonce, repo, _ := setupDeposerAnnonce(t)

	id, err := deposerAnnonce(validCommandArgs())

	require.NoError(t, err)
	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.True(t, stored.MotDePasseCorrespond("demo1234"))
	assert.False(t, stored.MotDePasseCorrespond("autre-mot"))
}

func TestDeposerAnnonce_RefusesAMotDePasseTooShortOrMissing(t *testing.T) {
	for _, motDePasse := range []string{"", "1234567"} {
		deposerAnnonce, repo, collector := setupDeposerAnnonce(t)
		args := validCommandArgs()
		args.MotDePasse = motDePasse

		_, err := deposerAnnonce(args)

		assert.ErrorIs(t, err, domain.ErrMotDePasseTropCourt)
		assert.Empty(t, collector.GetEvents())
		page, err := domain.NewPage(0)
		require.NoError(t, err)
		resultat, err := repo.Rechercher(domain.CriteresRecherche{}, page)
		require.NoError(t, err)
		assert.Zero(t, resultat.Total)
	}
}
