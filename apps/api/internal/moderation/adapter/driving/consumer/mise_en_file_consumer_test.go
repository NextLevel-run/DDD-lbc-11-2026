package consumer

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

// fakeAnnonceDeposee stands for the event of the annonce context: only its
// primitive accessors count.
type fakeAnnonceDeposee struct{}

func (fakeAnnonceDeposee) EventType() string    { return "AnnonceDeposee" }
func (fakeAnnonceDeposee) AnnonceID() string    { return "a1" }
func (fakeAnnonceDeposee) Titre() string        { return "Vélo de ville" }
func (fakeAnnonceDeposee) Description() string  { return "Très bon état." }
func (fakeAnnonceDeposee) Prix() int64          { return 15000 }
func (fakeAnnonceDeposee) Categorie() string    { return "vehicules" }
func (fakeAnnonceDeposee) Ville() string        { return "Lyon" }
func (fakeAnnonceDeposee) CodePostal() string   { return "69003" }
func (fakeAnnonceDeposee) Pseudo() string       { return "cyclo" }
func (fakeAnnonceDeposee) DeposeeLe() time.Time { return maintenant.Add(-time.Hour) }

// fakeAnnonceCorrigee is the corrected annonce, with the same accessors and
// its own type.
type fakeAnnonceCorrigee struct{ fakeAnnonceDeposee }

func (fakeAnnonceCorrigee) EventType() string { return "AnnonceCorrigee" }
func (fakeAnnonceCorrigee) Titre() string     { return "Vélo de ville révisé" }

type otherEvent struct{}

func (otherEvent) EventType() string { return "AnnonceDeposee" }

func setup(t *testing.T) (*inmemory.InMemoryHistoriqueAnnonceRepository, *eventbus.InMemoryEventBus) {
	t.Helper()
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	require.NoError(t, SubscribeMiseEnFile(bus, command.BuildMettreEnFileCommand(repo, func() time.Time { return maintenant })))
	return repo, bus
}

func TestSubscribeMiseEnFile_QueuesADeposedAnnonce(t *testing.T) {
	repo, bus := setup(t)

	require.NoError(t, bus.Publish(fakeAnnonceDeposee{}))

	h, err := repo.Charger("a1")
	require.NoError(t, err)
	assert.Equal(t, domain.EtatEnFile, h.Etat())
	assert.Equal(t, "Vélo de ville", h.Annonce().Titre)
	assert.Equal(t, "cyclo", h.Annonce().Pseudo)
	assert.Equal(t, maintenant, h.MisEnFileLe())
}

func TestSubscribeMiseEnFile_IgnoresAnEventWithoutTheAccessors(t *testing.T) {
	repo, bus := setup(t)

	require.NoError(t, bus.Publish(otherEvent{}))

	tous, err := repo.Tous()
	require.NoError(t, err)
	assert.Empty(t, tous)
}

func TestSubscribeMiseEnFile_QueuesACorrectedAnnonceBackWithItsHistory(t *testing.T) {
	repo, bus := setup(t)
	require.NoError(t, bus.Publish(fakeAnnonceDeposee{}))
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

	require.NoError(t, bus.Publish(fakeAnnonceCorrigee{}))

	h, err = repo.Charger("a1")
	require.NoError(t, err)
	assert.Equal(t, domain.EtatEnFile, h.Etat())
	assert.Equal(t, "Vélo de ville révisé", h.Annonce().Titre)
	assert.Len(t, h.Evenements(), 4)
}

func TestSubscribeMiseEnFile_IgnoresACorrectedAnnonceThatIsNotChallenged(t *testing.T) {
	repo, bus := setup(t)
	require.NoError(t, bus.Publish(fakeAnnonceDeposee{}))

	require.NoError(t, bus.Publish(fakeAnnonceCorrigee{}))

	h, err := repo.Charger("a1")
	require.NoError(t, err)
	assert.Len(t, h.Evenements(), 1)
	assert.Equal(t, "Vélo de ville", h.Annonce().Titre)
}
