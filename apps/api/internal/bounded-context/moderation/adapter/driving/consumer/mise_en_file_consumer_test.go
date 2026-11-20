package consumer

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/application/command"
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"
	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func unInstantane(titre string) contracts.InstantaneAnnonce {
	return contracts.InstantaneAnnonce{
		AnnonceID: "a1", Titre: titre, Description: "Très bon état.", Prix: 15000,
		Categorie: "vehicules", Ville: "Lyon", CodePostal: "69003", Pseudo: "cyclo",
		DeposeeLe: maintenant.Add(-time.Hour),
	}
}

func uneAnnonceDeposee() contracts.AnnonceDeposeeEvent {
	return contracts.AnnonceDeposeeEvent{Annonce: unInstantane("Vélo de ville"), Date: maintenant}
}

func uneAnnonceCorrigee() contracts.AnnonceCorrigeeEvent {
	return contracts.AnnonceCorrigeeEvent{Annonce: unInstantane("Vélo de ville révisé"), Date: maintenant}
}

// otherEvent has the type of an AnnonceDeposee but is not the contract.
type otherEvent struct{}

func (otherEvent) EventType() string { return contracts.AnnonceDeposeeEventType }

func setup(t *testing.T) (*inmemory.InMemoryHistoriqueAnnonceRepository, *eventbus.InMemoryEventBus) {
	t.Helper()
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	require.NoError(t, SubscribeMiseEnFile(bus, command.BuildMettreEnFileCommand(repo, func() time.Time { return maintenant })))
	return repo, bus
}

func TestSubscribeMiseEnFile_QueuesADeposedAnnonce(t *testing.T) {
	repo, bus := setup(t)

	require.NoError(t, bus.Publish(uneAnnonceDeposee()))

	h, err := repo.Charger("a1")
	require.NoError(t, err)
	assert.Equal(t, domain.EtatEnFile, h.Etat())
	assert.Equal(t, domain.InstantaneAnnonce{
		AnnonceID: "a1", Titre: "Vélo de ville", Description: "Très bon état.", Prix: 15000,
		Categorie: "vehicules", Ville: "Lyon", CodePostal: "69003", Pseudo: "cyclo",
		DeposeeLe: maintenant.Add(-time.Hour),
	}, h.Annonce())
	assert.Equal(t, maintenant, h.MisEnFileLe())
}

func TestSubscribeMiseEnFile_IgnoresAnEventThatIsNotTheContract(t *testing.T) {
	repo, bus := setup(t)

	require.NoError(t, bus.Publish(otherEvent{}))

	tous, err := repo.Tous()
	require.NoError(t, err)
	assert.Empty(t, tous)
}

func TestSubscribeMiseEnFile_QueuesACorrectedAnnonceBackWithItsHistory(t *testing.T) {
	repo, bus := setup(t)
	require.NoError(t, bus.Publish(uneAnnonceDeposee()))
	h, err := repo.Charger("a1")
	require.NoError(t, err)
	moderateur, err := domain.NewModerateur("alice")
	require.NoError(t, err)
	motif, err := domain.NewMotif("Photos manquantes")
	require.NoError(t, err)
	require.NoError(t, h.Verrouiller(moderateur, maintenant))
	_, err = h.Challenger(moderateur, motif, maintenant)
	require.NoError(t, err)
	require.NoError(t, repo.Enregistrer(h))

	require.NoError(t, bus.Publish(uneAnnonceCorrigee()))

	h, err = repo.Charger("a1")
	require.NoError(t, err)
	assert.Equal(t, domain.EtatEnFile, h.Etat())
	assert.Equal(t, "Vélo de ville révisé", h.Annonce().Titre)
	assert.Len(t, h.Evenements(), 4)
}

func TestSubscribeMiseEnFile_IgnoresACorrectedAnnonceThatIsNotChallenged(t *testing.T) {
	repo, bus := setup(t)
	require.NoError(t, bus.Publish(uneAnnonceDeposee()))

	require.NoError(t, bus.Publish(uneAnnonceCorrigee()))

	h, err := repo.Charger("a1")
	require.NoError(t, err)
	assert.Len(t, h.Evenements(), 1)
	assert.Equal(t, "Vélo de ville", h.Annonce().Titre)
}
