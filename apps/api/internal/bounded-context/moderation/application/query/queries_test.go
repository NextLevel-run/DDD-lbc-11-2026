package query

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func mettreEnFile(t *testing.T, repo domain.HistoriqueAnnonceRepository, id, titre string, le time.Time) {
	t.Helper()
	h, err := repo.Charger(id)
	require.NoError(t, err)
	h.MettreEnFile(domain.InstantaneAnnonce{
		AnnonceID: id, Titre: titre, Description: "Très bon état.", Prix: 1500, Categorie: "vehicules",
		Ville: "Lyon", CodePostal: "69003", Pseudo: "cyclo", DeposeeLe: le.Add(-time.Hour),
	}, le)
	require.NoError(t, repo.Enregistrer(h))
}

func verrouiller(t *testing.T, repo domain.HistoriqueAnnonceRepository, id, pseudo string, le time.Time) {
	t.Helper()
	h, err := repo.Charger(id)
	require.NoError(t, err)
	m, err := domain.NewModerateur(pseudo)
	require.NoError(t, err)
	require.NoError(t, h.Verrouiller(m, le))
	require.NoError(t, repo.Enregistrer(h))
}

func TestListerAnnoncesAModerer_ListsTheOldestFirstWithTheirLock(t *testing.T) {
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	mettreEnFile(t, repo, "recente", "Récente", maintenant.Add(-time.Minute))
	mettreEnFile(t, repo, "ancienne", "Ancienne", maintenant.Add(-time.Hour))
	verrouiller(t, repo, "ancienne", "alice", maintenant.Add(-30*time.Minute))

	views, err := BuildListerAnnoncesAModererQuery(repo, func() time.Time { return maintenant })()

	require.NoError(t, err)
	require.Len(t, views, 2)
	assert.Equal(t, "ancienne", views[0].ID)
	assert.Equal(t, "alice", views[0].VerrouillePar)
	assert.Equal(t, maintenant.Add(-30*time.Minute), views[0].VerrouilleDepuis)
	assert.Equal(t, "Ancienne", views[0].Titre)
	assert.Equal(t, int64(1500), views[0].Prix)
	assert.Equal(t, "cyclo", views[0].Pseudo)
	assert.Equal(t, "recente", views[1].ID)
	assert.Empty(t, views[1].VerrouillePar)
	assert.True(t, views[1].VerrouilleDepuis.IsZero())
}

func TestListerAnnoncesAModerer_CountsAnExpiredLockAsFree(t *testing.T) {
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	mettreEnFile(t, repo, "a1", "Vélo", maintenant.Add(-5*time.Hour))
	verrouiller(t, repo, "a1", "alice", maintenant.Add(-3*time.Hour))

	views, err := BuildListerAnnoncesAModererQuery(repo, func() time.Time { return maintenant })()

	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.Empty(t, views[0].VerrouillePar)
}

func TestListerAnnoncesAModerer_IsEmptyWithoutAnyAnnonce(t *testing.T) {
	views, err := BuildListerAnnoncesAModererQuery(inmemory.NewInMemoryHistoriqueAnnonceRepository(), func() time.Time { return maintenant })()

	require.NoError(t, err)
	assert.NotNil(t, views)
	assert.Empty(t, views)
}

func TestConsulterAnnonceAModerer_ReturnsTheSnapshotTheStateAndTheHistory(t *testing.T) {
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	mettreEnFile(t, repo, "a1", "Vélo", maintenant.Add(-time.Hour))
	verrouiller(t, repo, "a1", "alice", maintenant.Add(-time.Minute))

	view, err := BuildConsulterAnnonceAModererQuery(repo, func() time.Time { return maintenant })("a1")

	require.NoError(t, err)
	assert.Equal(t, "Vélo", view.Titre)
	assert.Equal(t, "Très bon état.", view.Description)
	assert.Equal(t, "69003", view.CodePostal)
	assert.Equal(t, "verrouillee", view.Statut)
	assert.Equal(t, "alice", view.VerrouillePar)
	require.Len(t, view.Historique, 2)
	assert.Equal(t, EvenementView{Type: "AnnonceMiseEnFile", Date: maintenant.Add(-time.Hour)}, view.Historique[0])
	assert.Equal(t, EvenementView{Type: "AnnonceVerrouillee", Date: maintenant.Add(-time.Minute), Moderateur: "alice"}, view.Historique[1])
}

func TestConsulterAnnonceAModerer_ShowsAnExpiredLockAsFreeButKeepsTheHistory(t *testing.T) {
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	mettreEnFile(t, repo, "a1", "Vélo", maintenant.Add(-5*time.Hour))
	verrouiller(t, repo, "a1", "alice", maintenant.Add(-3*time.Hour))

	view, err := BuildConsulterAnnonceAModererQuery(repo, func() time.Time { return maintenant })("a1")

	require.NoError(t, err)
	assert.Equal(t, "en-file", view.Statut)
	assert.Empty(t, view.VerrouillePar)
	assert.Len(t, view.Historique, 2)
}

func TestConsulterAnnonceAModerer_AnswersIntrouvableForAnUnknownAnnonce(t *testing.T) {
	_, err := BuildConsulterAnnonceAModererQuery(inmemory.NewInMemoryHistoriqueAnnonceRepository(), func() time.Time { return maintenant })("inconnue")

	assert.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
}

func decider(t *testing.T, repo domain.HistoriqueAnnonceRepository, id string, f func(h *domain.HistoriqueAnnonce, m domain.Moderateur) error) {
	t.Helper()
	h, err := repo.Charger(id)
	require.NoError(t, err)
	m, err := domain.NewModerateur("alice")
	require.NoError(t, err)
	require.NoError(t, f(h, m))
	require.NoError(t, repo.Enregistrer(h))
}

func TestDecidedAnnonces_LeaveTheQueueAndShowTheirDecisionInTheHistory(t *testing.T) {
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	for _, id := range []string{"valide", "refusee", "challengee", "en-attente"} {
		mettreEnFile(t, repo, id, id, maintenant.Add(-time.Hour))
	}
	for _, id := range []string{"valide", "refusee", "challengee"} {
		verrouiller(t, repo, id, "alice", maintenant.Add(-30*time.Minute))
	}
	decider(t, repo, "valide", func(h *domain.HistoriqueAnnonce, m domain.Moderateur) error {
		_, err := h.Valider(m, maintenant)
		return err
	})
	decider(t, repo, "refusee", func(h *domain.HistoriqueAnnonce, m domain.Moderateur) error {
		_, err := h.Refuser(m, maintenant)
		return err
	})
	decider(t, repo, "challengee", func(h *domain.HistoriqueAnnonce, m domain.Moderateur) error {
		motif, _ := domain.NewMotif("Photos manquantes")
		_, err := h.Challenger(m, motif, maintenant)
		return err
	})
	clock := func() time.Time { return maintenant }

	file, err := BuildListerAnnoncesAModererQuery(repo, clock)()
	require.NoError(t, err)
	require.Len(t, file, 1)
	assert.Equal(t, "en-attente", file[0].ID)

	consulter := BuildConsulterAnnonceAModererQuery(repo, clock)
	for id, statut := range map[string]string{"valide": "validee", "refusee": "refusee", "challengee": "challengee"} {
		view, err := consulter(id)
		require.NoError(t, err)
		assert.Equal(t, statut, view.Statut, id)
		assert.Empty(t, view.VerrouillePar, id)
	}
	view, _ := consulter("challengee")
	derniere := view.Historique[len(view.Historique)-1]
	assert.Equal(t, "AnnonceChallengee", derniere.Type)
	assert.Equal(t, "alice", derniere.Moderateur)
	assert.Equal(t, "Photos manquantes", derniere.Motif)
}
