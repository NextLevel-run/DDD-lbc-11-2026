package inmemory

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func enFile(id string) *domain.HistoriqueAnnonce {
	h := domain.NewHistoriqueAnnonce(id)
	h.MettreEnFile(domain.InstantaneAnnonce{AnnonceID: id, Titre: "Vélo"}, maintenant)
	return h
}

func TestCharger_AnswersAnEmptyHistoriqueForAnUnknownAnnonce(t *testing.T) {
	repo := NewInMemoryHistoriqueAnnonceRepository()

	h, err := repo.Charger("inconnue")

	require.NoError(t, err)
	assert.Equal(t, domain.EtatVide, h.Etat())
	assert.Equal(t, 0, h.Version())
}

func TestEnregistrer_AppendsTheNewEventsAndChargerReplaysThem(t *testing.T) {
	repo := NewInMemoryHistoriqueAnnonceRepository()
	require.NoError(t, repo.Enregistrer(enFile("a1")))
	h, err := repo.Charger("a1")
	require.NoError(t, err)
	moderateur, err := domain.NewModerateur("alice")
	require.NoError(t, err)
	require.NoError(t, h.Verrouiller(moderateur, maintenant))
	require.NoError(t, repo.Enregistrer(h))

	stored, err := repo.Charger("a1")

	require.NoError(t, err)
	assert.Equal(t, domain.EtatVerrouillee, stored.Etat())
	assert.Equal(t, 2, stored.Version())
	assert.Len(t, stored.Evenements(), 2)
}

func TestEnregistrer_RefusesAStaleVersion(t *testing.T) {
	repo := NewInMemoryHistoriqueAnnonceRepository()
	require.NoError(t, repo.Enregistrer(enFile("a1")))
	first, _ := repo.Charger("a1")
	second, _ := repo.Charger("a1")
	alice, _ := domain.NewModerateur("alice")
	bob, _ := domain.NewModerateur("bob")
	require.NoError(t, first.Verrouiller(alice, maintenant))
	require.NoError(t, second.Verrouiller(bob, maintenant))
	require.NoError(t, repo.Enregistrer(first))

	err := repo.Enregistrer(second)

	assert.ErrorIs(t, err, domain.ErrConflitVersion)
	stored, _ := repo.Charger("a1")
	assert.Len(t, stored.Evenements(), 2)
}

func TestEnregistrer_RefusesAFreshStreamOverAnExistingOne(t *testing.T) {
	repo := NewInMemoryHistoriqueAnnonceRepository()
	require.NoError(t, repo.Enregistrer(enFile("a1")))

	err := repo.Enregistrer(enFile("a1"))

	assert.ErrorIs(t, err, domain.ErrConflitVersion)
}

func TestTous_ReplaysEveryStream(t *testing.T) {
	repo := NewInMemoryHistoriqueAnnonceRepository()
	require.NoError(t, repo.Enregistrer(enFile("b")))
	require.NoError(t, repo.Enregistrer(enFile("a")))

	tous, err := repo.Tous()

	require.NoError(t, err)
	require.Len(t, tous, 2)
	assert.Equal(t, "a", tous[0].AnnonceID())
	assert.Equal(t, domain.EtatEnFile, tous[1].Etat())
}
