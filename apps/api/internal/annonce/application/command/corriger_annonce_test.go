package command

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCorrigerAnnonce(t *testing.T) (CorrigerAnnonceCommand, *inmemory.InMemoryAnnonceRepository, *eventbustesting.EventCollector, string) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, bus.Subscribe(domain.AnnonceCorrigeeEventType, collector.EventHandler()))
	id, err := BuildDeposerAnnonceCommand(repo, bus, func() time.Time { return maintenant })(validCommandArgs())
	require.NoError(t, err)
	annonce, err := repo.GetByID(id)
	require.NoError(t, err)
	require.NoError(t, annonce.Challenger("Photos manquantes"))
	require.NoError(t, repo.Save(annonce))
	corrigeeLe := maintenant.Add(24 * time.Hour)
	return BuildCorrigerAnnonceCommand(repo, bus, func() time.Time { return corrigeeLe }), repo, collector, id
}

func uneCorrectionArgs(id string) CorrigerAnnonceCommandArgs {
	return CorrigerAnnonceCommandArgs{
		ID: id, MotDePasse: "demo1234", Titre: "Vélo de ville révisé", Description: "Un vélo, photos ajoutées.",
		Prix: 12000, Categorie: "vehicules", Ville: "Lyon", CodePostal: "69003",
	}
}

func TestCorrigerAnnonce_StoresTheCorrectionEnAttenteModeration(t *testing.T) {
	corriger, repo, _, id := setupCorrigerAnnonce(t)

	require.NoError(t, corriger(uneCorrectionArgs(id)))

	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, domain.EnAttenteModeration, stored.Statut())
	assert.Empty(t, stored.Motif())
	assert.Equal(t, "Vélo de ville révisé", stored.Titre())
	assert.Equal(t, int64(12000), stored.Prix().Centimes())
	assert.Equal(t, "Lyon", stored.Localisation().Ville())
}

func TestCorrigerAnnonce_PublishesAnnonceCorrigeeAfterPersisting(t *testing.T) {
	corriger, _, collector, id := setupCorrigerAnnonce(t)

	require.NoError(t, corriger(uneCorrectionArgs(id)))

	events := collector.GetEvents()
	require.Len(t, events, 1)
	event, ok := events[0].(*domain.AnnonceCorrigeeEvent)
	require.True(t, ok)
	assert.Equal(t, id, event.AnnonceID())
	assert.Equal(t, "Vélo de ville révisé", event.Titre())
	assert.Equal(t, maintenant.Add(24*time.Hour), event.OccurredAt())
}

func TestCorrigerAnnonce_FailsWithoutStoringNorPublishing(t *testing.T) {
	cases := map[string]struct {
		mutate func(*CorrigerAnnonceCommandArgs)
		want   error
	}{
		"malformed ID":   {func(a *CorrigerAnnonceCommandArgs) { a.ID = "pas-un-uuid" }, domain.ErrIDAnnonceInvalide},
		"unknown":        {func(a *CorrigerAnnonceCommandArgs) { a.ID = "7d1cbd8e-2a2c-4f1c-a8f2-6c1f0a3f4b11" }, domain.ErrAnnonceIntrouvable},
		"wrong password": {func(a *CorrigerAnnonceCommandArgs) { a.MotDePasse = "mauvais-mot" }, domain.ErrMotDePasseIncorrect},
		"invalid titre":  {func(a *CorrigerAnnonceCommandArgs) { a.Titre = "ab" }, domain.ErrTitreInvalide},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			corriger, repo, collector, id := setupCorrigerAnnonce(t)
			args := uneCorrectionArgs(id)
			tc.mutate(&args)

			assert.ErrorIs(t, corriger(args), tc.want)

			assert.Empty(t, collector.GetEvents())
			stored, err := repo.GetByID(id)
			require.NoError(t, err)
			assert.Equal(t, domain.Challengee, stored.Statut())
			assert.Equal(t, "Photos manquantes", stored.Motif())
		})
	}
}

func TestCorrigerAnnonce_AnswersIntrouvableWhenNotChallengee(t *testing.T) {
	corriger, repo, collector, id := setupCorrigerAnnonce(t)
	require.NoError(t, corriger(uneCorrectionArgs(id)))
	collector.Clear()

	err := corriger(uneCorrectionArgs(id))

	assert.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
	assert.Empty(t, collector.GetEvents())
	stored, _ := repo.GetByID(id)
	assert.Equal(t, domain.EnAttenteModeration, stored.Statut())
}
