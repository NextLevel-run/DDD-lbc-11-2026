package query

import (
	"testing"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAEditer(t *testing.T, statut domain.Statut) (ConsulterAnnonceAEditerQuery, string) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	annonce, err := domain.RestoreAnnonce("7d1cbd8e-2a2c-4f1c-a8f2-6c1f0a3f4b11", domain.EnAttenteModeration, annonceArgs())
	require.NoError(t, err)
	switch statut {
	case domain.Challengee:
		require.NoError(t, annonce.Challenger("Photos manquantes"))
	case domain.Activee:
		require.NoError(t, annonce.Activer())
	}
	require.NoError(t, repo.Save(annonce))
	return BuildConsulterAnnonceAEditerQuery(repo), annonce.ID()
}

func TestConsulterAnnonceAEditer_ReturnsTheEditableFieldsAndTheMotif(t *testing.T) {
	consulter, id := setupAEditer(t, domain.Challengee)

	view, err := consulter(id, "demo1234")

	require.NoError(t, err)
	assert.Equal(t, AnnonceAEditerView{
		Titre: "Peugeot 208", Description: "Première main, 45 000 km.", Prix: 1250000,
		Categorie: "vehicules", Ville: "Ajaccio", CodePostal: "20000", Motif: "Photos manquantes",
	}, *view)
}

func TestConsulterAnnonceAEditer_RefusesAWrongMotDePasse(t *testing.T) {
	consulter, id := setupAEditer(t, domain.Challengee)

	_, err := consulter(id, "mauvais-mot")

	assert.ErrorIs(t, err, domain.ErrMotDePasseIncorrect)
}

func TestConsulterAnnonceAEditer_AnswersIntrouvableUnlessChallengee(t *testing.T) {
	for name, statut := range map[string]domain.Statut{"activee": domain.Activee, "en attente": domain.EnAttenteModeration} {
		t.Run(name, func(t *testing.T) {
			consulter, id := setupAEditer(t, statut)

			_, err := consulter(id, "demo1234")

			assert.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
		})
	}
}

func TestConsulterAnnonceAEditer_RejectsAMalformedIDAndAnUnknownAnnonce(t *testing.T) {
	consulter, _ := setupAEditer(t, domain.Challengee)

	_, err := consulter("pas-un-uuid", "demo1234")
	assert.ErrorIs(t, err, domain.ErrIDAnnonceInvalide)

	_, err = consulter("11111111-2a2c-4f1c-a8f2-6c1f0a3f4b11", "demo1234")
	assert.ErrorIs(t, err, domain.ErrAnnonceIntrouvable)
}
