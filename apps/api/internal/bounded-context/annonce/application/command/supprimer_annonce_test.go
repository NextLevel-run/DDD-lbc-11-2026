package command

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSupprimerAnnonce(t *testing.T) (SupprimerAnnonceCommand, *inmemory.InMemoryAnnonceRepository, *eventbustesting.EventCollector, string) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, bus.Subscribe(domain.AnnonceSupprimeeEventType, collector.EventHandler()))
	id, err := BuildDeposerAnnonceCommand(repo, bus, func() time.Time { return maintenant })(validCommandArgs())
	require.NoError(t, err)
	annonce, err := repo.GetByID(id)
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, repo.Save(annonce))
	supprimeeLe := maintenant.Add(48 * time.Hour)
	return BuildSupprimerAnnonceCommand(repo, bus, func() time.Time { return supprimeeLe }), repo, collector, id
}

func TestSupprimerAnnonce_MarksTheAnnonceSupprimeeAndKeepsItStored(t *testing.T) {
	supprimer, repo, _, id := setupSupprimerAnnonce(t)

	err := supprimer(SupprimerAnnonceCommandArgs{ID: id, MotDePasse: "demo1234", Raison: "vendue-ailleurs"})

	require.NoError(t, err)
	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, domain.Supprimee, stored.Statut())
}

func TestSupprimerAnnonce_PublishesAnnonceSupprimeeWithTheRaison(t *testing.T) {
	supprimer, _, collector, id := setupSupprimerAnnonce(t)

	require.NoError(t, supprimer(SupprimerAnnonceCommandArgs{ID: id, MotDePasse: "demo1234", Raison: "autre"}))

	events := collector.GetEvents()
	require.Len(t, events, 1)
	event, ok := events[0].(*domain.AnnonceSupprimeeEvent)
	require.True(t, ok)
	assert.Equal(t, "AnnonceSupprimee", event.EventType())
	assert.Equal(t, id, event.AnnonceID)
	assert.Equal(t, "autre", event.Raison.Code())
	assert.Equal(t, maintenant.Add(48*time.Hour), event.SupprimeeLe)
}

func TestSupprimerAnnonce_RefusesAWrongMotDePasseWithoutPublishing(t *testing.T) {
	supprimer, repo, collector, id := setupSupprimerAnnonce(t)

	err := supprimer(SupprimerAnnonceCommandArgs{ID: id, MotDePasse: "mauvais-mot", Raison: "autre"})

	assert.ErrorIs(t, err, domain.ErrMotDePasseIncorrect)
	assert.Empty(t, collector.GetEvents())
	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.True(t, stored.EstActivee())
}

func TestSupprimerAnnonce_AnswersIntrouvableWhenAlreadySupprimee(t *testing.T) {
	supprimer, _, collector, id := setupSupprimerAnnonce(t)
	args := SupprimerAnnonceCommandArgs{ID: id, MotDePasse: "demo1234", Raison: "autre"}
	require.NoError(t, supprimer(args))
	collector.Clear()

	err := supprimer(args)

	assert.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
	assert.Empty(t, collector.GetEvents())
}

func TestSupprimerAnnonce_RefusesAMissingOrUnknownRaison(t *testing.T) {
	for _, raison := range []string{"", "inconnue"} {
		supprimer, repo, collector, id := setupSupprimerAnnonce(t)

		err := supprimer(SupprimerAnnonceCommandArgs{ID: id, MotDePasse: "demo1234", Raison: raison})

		assert.ErrorIs(t, err, domain.ErrRaisonSuppressionInvalide)
		assert.Empty(t, collector.GetEvents())
		stored, _ := repo.GetByID(id)
		assert.True(t, stored.EstActivee())
	}
}

func TestSupprimerAnnonce_RejectsAMalformedOrUnknownID(t *testing.T) {
	supprimer, _, _, _ := setupSupprimerAnnonce(t)

	assert.ErrorIs(t, supprimer(SupprimerAnnonceCommandArgs{ID: "pas-un-uuid", MotDePasse: "demo1234", Raison: "autre"}), domain.ErrIDAnnonceInvalide)
	assert.ErrorIs(t, supprimer(SupprimerAnnonceCommandArgs{ID: "6f1c1c2e-8a43-4b7e-9d52-0a1b2c3d4e5f", MotDePasse: "demo1234", Raison: "autre"}), domain.ErrAnnonceIntrouvable)
}

func TestSupprimerAnnonce_AnswersIntrouvableForAnAnnonceNotActivee(t *testing.T) {
	repo := inmemory.NewInMemoryAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, bus.Subscribe(domain.AnnonceSupprimeeEventType, collector.EventHandler()))
	id, err := BuildDeposerAnnonceCommand(repo, bus, func() time.Time { return maintenant })(validCommandArgs())
	require.NoError(t, err)
	supprimer := BuildSupprimerAnnonceCommand(repo, bus, func() time.Time { return maintenant })

	err = supprimer(SupprimerAnnonceCommandArgs{ID: id, MotDePasse: "demo1234", Raison: "autre"})

	assert.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
	assert.Empty(t, collector.GetEvents())
	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, domain.EnAttenteModeration, stored.Statut())
}
