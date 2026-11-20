package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/application/command"
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/application/query"
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func setupHandler(t *testing.T) (*inmemory.InMemoryHistoriqueAnnonceRepository, *http.ServeMux, *time.Time) {
	t.Helper()
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	now := maintenant
	clock := func() time.Time { return now }
	bus := eventbus.NewSyncInMemoryEventBus()
	mux := http.NewServeMux()
	NewHandler(
		query.BuildListerAnnoncesAModererQuery(repo, clock),
		query.BuildConsulterAnnonceAModererQuery(repo, clock),
		command.BuildVerrouillerAnnonceCommand(repo, clock),
		command.BuildRelacherVerrouCommand(repo, clock),
		command.BuildValiderAnnonceCommand(repo, bus, clock),
		command.BuildRefuserAnnonceCommand(repo, bus, clock),
		command.BuildChallengerAnnonceCommand(repo, bus, clock),
	).RegisterRoutes(mux)
	h := domain.NewHistoriqueAnnonce("a1")
	h.MettreEnFile(domain.InstantaneAnnonce{
		AnnonceID: "a1", Titre: "Vélo de ville", Description: "Très bon état.", Prix: 15000, Categorie: "vehicules",
		Ville: "Lyon", CodePostal: "69003", Pseudo: "cyclo", DeposeeLe: maintenant.Add(-time.Hour),
	}, maintenant)
	require.NoError(t, repo.Enregistrer(h))
	return repo, mux, &now
}

func post(mux *http.ServeMux, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	return rec
}

func get(mux *http.ServeMux, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

func TestListerAnnonces_AnswersTheQueueInFrenchJSON(t *testing.T) {
	_, mux, _ := setupHandler(t)

	rec := get(mux, "/moderation/annonces")

	require.Equal(t, http.StatusOK, rec.Code)
	annonces := decode(t, rec)["annonces"].([]any)
	require.Len(t, annonces, 1)
	annonce := annonces[0].(map[string]any)
	assert.Equal(t, "a1", annonce["id"])
	assert.Equal(t, "Vélo de ville", annonce["titre"])
	assert.EqualValues(t, 15000, annonce["prix"])
	assert.Equal(t, "Lyon", annonce["ville"])
	assert.Equal(t, "cyclo", annonce["pseudo"])
	assert.Equal(t, "", annonce["verrouillePar"])
	assert.NotContains(t, annonce, "verrouilleDepuis")
	assert.Contains(t, annonce, "misEnFileLe")
}

func TestVerrouiller_LocksThenTheListAndTheDetailShowIt(t *testing.T) {
	_, mux, _ := setupHandler(t)

	rec := post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`)

	require.Equal(t, http.StatusNoContent, rec.Code)
	liste := decode(t, get(mux, "/moderation/annonces"))["annonces"].([]any)[0].(map[string]any)
	assert.Equal(t, "alice", liste["verrouillePar"])
	assert.Contains(t, liste, "verrouilleDepuis")
	detail := decode(t, get(mux, "/moderation/annonces/a1"))
	assert.Equal(t, "verrouillee", detail["statut"])
	assert.Equal(t, "alice", detail["verrouillePar"])
	assert.Equal(t, "69003", detail["codePostal"])
	assert.Equal(t, "vehicules", detail["categorie"])
	historique := detail["historique"].([]any)
	require.Len(t, historique, 2)
	assert.Equal(t, "AnnonceVerrouillee", historique[1].(map[string]any)["type"])
	assert.Equal(t, "alice", historique[1].(map[string]any)["moderateur"])
}

func TestVerrouiller_AnswersConflictWhenAnotherModerateurHoldsTheLock(t *testing.T) {
	_, mux, _ := setupHandler(t)
	require.Equal(t, http.StatusNoContent, post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`).Code)

	rec := post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"bob"}`)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "Cette annonce est déjà verrouillée par un autre modérateur", decode(t, rec)["erreur"])
}

func TestVerrouiller_TakesOverAnExpiredLock(t *testing.T) {
	_, mux, now := setupHandler(t)
	require.Equal(t, http.StatusNoContent, post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`).Code)
	*now = now.Add(2*time.Hour + time.Second)

	assert.Equal(t, http.StatusNoContent, post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"bob"}`).Code)
}

func TestVerrouiller_NamesTheModerateurAtFault(t *testing.T) {
	_, mux, _ := setupHandler(t)

	rec := post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"a"}`)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "moderateur", decode(t, rec)["champ"])
	assert.Equal(t, http.StatusBadRequest, post(mux, "/moderation/annonces/a1/verrou", `pas du json`).Code)
}

func TestRelacher_PutsTheAnnonceBackInTheQueue(t *testing.T) {
	_, mux, _ := setupHandler(t)
	post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`)

	rec := post(mux, "/moderation/annonces/a1/relachement", `{"moderateur":"alice"}`)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "en-file", decode(t, get(mux, "/moderation/annonces/a1"))["statut"])
}

func TestRelacher_AnswersConflictWithoutAValidLockOfTheModerateur(t *testing.T) {
	_, mux, now := setupHandler(t)

	assert.Equal(t, http.StatusConflict, post(mux, "/moderation/annonces/a1/relachement", `{"moderateur":"alice"}`).Code)
	post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`)
	assert.Equal(t, http.StatusConflict, post(mux, "/moderation/annonces/a1/relachement", `{"moderateur":"bob"}`).Code)
	*now = now.Add(3 * time.Hour)
	rec := post(mux, "/moderation/annonces/a1/relachement", `{"moderateur":"alice"}`)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "Votre verrou a expiré", decode(t, rec)["erreur"])
}

func TestUnknownAnnonce_IsA404OnEveryRoute(t *testing.T) {
	_, mux, _ := setupHandler(t)

	assert.Equal(t, http.StatusNotFound, get(mux, "/moderation/annonces/inconnue").Code)
	assert.Equal(t, http.StatusNotFound, post(mux, "/moderation/annonces/inconnue/verrou", `{"moderateur":"alice"}`).Code)
	assert.Equal(t, http.StatusNotFound, post(mux, "/moderation/annonces/inconnue/relachement", `{"moderateur":"alice"}`).Code)
}

func TestDecisions_AnswerNoContentAndTakeTheAnnonceOutOfTheQueue(t *testing.T) {
	cas := []struct{ chemin, corps, statut string }{
		{"validation", `{"moderateur":"alice"}`, "validee"},
		{"refus", `{"moderateur":"alice"}`, "refusee"},
		{"challenge", `{"moderateur":"alice","motif":"Photos manquantes"}`, "challengee"},
	}
	for _, c := range cas {
		t.Run(c.chemin, func(t *testing.T) {
			_, mux, _ := setupHandler(t)
			require.Equal(t, http.StatusNoContent, post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`).Code)

			rec := post(mux, "/moderation/annonces/a1/"+c.chemin, c.corps)

			require.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, decode(t, get(mux, "/moderation/annonces"))["annonces"])
			detail := decode(t, get(mux, "/moderation/annonces/a1"))
			assert.Equal(t, c.statut, detail["statut"])
			assert.Equal(t, "", detail["verrouillePar"])
			historique := detail["historique"].([]any)
			derniere := historique[len(historique)-1].(map[string]any)
			assert.Equal(t, "alice", derniere["moderateur"])
		})
	}
}

func TestChallenger_ShowsTheMotifInTheHistory(t *testing.T) {
	_, mux, _ := setupHandler(t)
	post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`)
	post(mux, "/moderation/annonces/a1/challenge", `{"moderateur":"alice","motif":"  Photos manquantes  "}`)

	historique := decode(t, get(mux, "/moderation/annonces/a1"))["historique"].([]any)

	derniere := historique[len(historique)-1].(map[string]any)
	assert.Equal(t, "AnnonceChallengee", derniere["type"])
	assert.Equal(t, "Photos manquantes", derniere["motif"])
}

func TestChallenger_NamesTheMotifAtFault(t *testing.T) {
	_, mux, _ := setupHandler(t)
	post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`)

	for _, corps := range []string{
		`{"moderateur":"alice"}`,
		`{"moderateur":"alice","motif":"court"}`,
		`{"moderateur":"alice","motif":"` + strings.Repeat("x", 1001) + `"}`,
	} {
		rec := post(mux, "/moderation/annonces/a1/challenge", corps)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "motif", decode(t, rec)["champ"])
	}
	assert.Equal(t, http.StatusBadRequest, post(mux, "/moderation/annonces/a1/challenge", `{"moderateur":"a","motif":"Photos manquantes"}`).Code)
	assert.Equal(t, http.StatusBadRequest, post(mux, "/moderation/annonces/a1/challenge", `pas du json`).Code)
	assert.Equal(t, "verrouillee", decode(t, get(mux, "/moderation/annonces/a1"))["statut"])
}

func TestDecisions_AnswerConflictWithoutAValidLockOfTheModerateur(t *testing.T) {
	corps := map[string]string{
		"validation": `{"moderateur":"alice"}`,
		"refus":      `{"moderateur":"alice"}`,
		"challenge":  `{"moderateur":"alice","motif":"Photos manquantes"}`,
	}
	for chemin, body := range corps {
		t.Run(chemin, func(t *testing.T) {
			_, mux, now := setupHandler(t)
			url := "/moderation/annonces/a1/" + chemin

			rec := post(mux, url, body)
			assert.Equal(t, http.StatusConflict, rec.Code, "not locked")
			assert.Equal(t, "Cette annonce n'est pas verrouillée par vous", decode(t, rec)["erreur"])

			post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"bob"}`)
			assert.Equal(t, http.StatusConflict, post(mux, url, body).Code, "locked by another")

			*now = now.Add(2*time.Hour + time.Second)
			post(mux, "/moderation/annonces/a1/verrou", `{"moderateur":"alice"}`)
			*now = now.Add(2*time.Hour + time.Second)
			rec = post(mux, url, body)
			assert.Equal(t, http.StatusConflict, rec.Code, "expired")
			assert.Equal(t, "Votre verrou a expiré", decode(t, rec)["erreur"])
			assert.Equal(t, http.StatusNotFound, post(mux, "/moderation/annonces/inconnue/"+chemin, body).Code)
		})
	}
}
