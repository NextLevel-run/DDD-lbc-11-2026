package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ddd-second-hand-marketplace/internal/annonce/adapter/driven/inmemory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deposerUneAnnonce deposits an annonce and activates it, as moderation will.
func deposerUneAnnonce(t *testing.T, repo *inmemory.InMemoryAnnonceRepository, mux *http.ServeMux) string {
	t.Helper()
	var body DeposerAnnonceResponse
	require.NoError(t, json.NewDecoder(post(mux, "/annonces", depotValide).Body).Decode(&body))
	annonce, err := repo.GetByID(body.ID)
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	return body.ID
}

func TestSupprimerAnnonce_AnswersNoContentAndTakesTheAnnonceOffline(t *testing.T) {
	repo, mux := setupHandler(t)
	id := deposerUneAnnonce(t, repo, mux)

	rec := post(mux, "/annonces/"+id+"/suppression", `{"motDePasse": "demo1234", "raison": "vendue-sur-le-site"}`)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, http.StatusNotFound, get(mux, "/annonces/"+id).Code)
	var recherche RechercherAnnoncesResponse
	require.NoError(t, json.NewDecoder(get(mux, "/annonces").Body).Decode(&recherche))
	assert.Zero(t, recherche.Total)
}

func TestSupprimerAnnonce_AnswersBadRequestNamingTheField(t *testing.T) {
	tests := []struct {
		name, body, champ string
	}{
		{"wrong mot de passe", `{"motDePasse": "mauvais-mot", "raison": "autre"}`, "motDePasse"},
		{"missing raison", `{"motDePasse": "demo1234"}`, "raison"},
		{"unknown raison", `{"motDePasse": "demo1234", "raison": "inconnue"}`, "raison"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mux := setupHandler(t)
			id := deposerUneAnnonce(t, repo, mux)

			rec := post(mux, "/annonces/"+id+"/suppression", tt.body)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), `"champ":"`+tt.champ+`"`)
			assert.Equal(t, http.StatusOK, get(mux, "/annonces/"+id).Code)
		})
	}
}

func TestSupprimerAnnonce_AnswersNotFoundForAnUnknownOrAlreadySupprimeeAnnonce(t *testing.T) {
	repo, mux := setupHandler(t)
	id := deposerUneAnnonce(t, repo, mux)
	body := `{"motDePasse": "demo1234", "raison": "autre"}`
	require.Equal(t, http.StatusNoContent, post(mux, "/annonces/"+id+"/suppression", body).Code)

	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/"+id+"/suppression", body).Code)
	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/6f1c1c2e-8a43-4b7e-9d52-0a1b2c3d4e5f/suppression", body).Code)
}

func TestSupprimerAnnonce_AnswersBadRequestForAMalformedID(t *testing.T) {
	_, mux := setupHandler(t)

	rec := post(mux, "/annonces/nope/suppression", `{"motDePasse": "demo1234", "raison": "autre"}`)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.False(t, strings.Contains(rec.Body.String(), "champ"))
}

func TestSupprimerAnnonce_AnswersNotFoundForAnAnnonceStillEnAttenteModeration(t *testing.T) {
	_, mux := setupHandler(t)
	var body DeposerAnnonceResponse
	require.NoError(t, json.NewDecoder(post(mux, "/annonces", depotValide).Body).Decode(&body))

	rec := post(mux, "/annonces/"+body.ID+"/suppression", `{"motDePasse": "demo1234", "raison": "autre"}`)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
