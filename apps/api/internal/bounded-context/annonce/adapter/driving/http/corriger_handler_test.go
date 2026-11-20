package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deposerUneAnnonceChallengee deposits an annonce and has moderation
// challenge it.
func deposerUneAnnonceChallengee(t *testing.T, repo *inmemory.InMemoryAnnonceRepository, mux *http.ServeMux) string {
	t.Helper()
	var body DeposerAnnonceResponse
	require.NoError(t, json.NewDecoder(post(mux, "/annonces", depotValide).Body).Decode(&body))
	annonce, err := repo.GetByID(body.ID)
	require.NoError(t, err)
	require.NoError(t, annonce.Challenger("Photos manquantes"))
	return body.ID
}

const correctionValide = `{
	"motDePasse": "demo1234",
	"titre": "Vélo de course révisé",
	"description": "Cadre carbone, photos ajoutées.",
	"prix": 42000,
	"categorie": "loisirs",
	"ville": "Lyon",
	"codePostal": "69003"
}`

func TestEditerAnnonce_AnswersTheEditableFieldsAndTheMotif(t *testing.T) {
	repo, mux := setupHandler(t)
	id := deposerUneAnnonceChallengee(t, repo, mux)

	rec := post(mux, "/annonces/"+id+"/edition", `{"motDePasse": "demo1234"}`)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{
		"titre": "Vélo de course", "description": "Cadre carbone, très bon état.", "prix": 45000,
		"categorie": "loisirs", "ville": "Lyon", "codePostal": "69003", "motif": "Photos manquantes"
	}`, rec.Body.String())
}

func TestEditerAnnonce_AnswersBadRequestForAWrongMotDePasse(t *testing.T) {
	repo, mux := setupHandler(t)
	id := deposerUneAnnonceChallengee(t, repo, mux)

	rec := post(mux, "/annonces/"+id+"/edition", `{"motDePasse": "mauvais-mot"}`)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "motDePasse", decodeErreur(t, rec).Champ)
}

func TestEditerAnnonce_AnswersNotFoundUnlessTheAnnonceIsChallengee(t *testing.T) {
	repo, mux := setupHandler(t)
	activee := deposerUneAnnonce(t, repo, mux)
	body := `{"motDePasse": "demo1234"}`

	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/"+activee+"/edition", body).Code)
	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/6f1c1c2e-8a43-4b7e-9d52-0a1b2c3d4e5f/edition", body).Code)
	assert.Equal(t, http.StatusBadRequest, post(mux, "/annonces/nope/edition", body).Code)
	assert.Equal(t, http.StatusBadRequest, post(mux, "/annonces/"+activee+"/edition", "pas du json").Code)
}

func TestCorrigerAnnonce_AnswersNoContentAndSendsTheAnnonceBackToModeration(t *testing.T) {
	repo, mux := setupHandler(t)
	id := deposerUneAnnonceChallengee(t, repo, mux)

	rec := post(mux, "/annonces/"+id+"/correction", correctionValide)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	stored, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, domain.EnAttenteModeration, stored.Statut())
	assert.Equal(t, "Vélo de course révisé", stored.Titre())
	assert.Equal(t, http.StatusNotFound, get(mux, "/annonces/"+id).Code)
}

func TestCorrigerAnnonce_AnswersBadRequestNamingTheField(t *testing.T) {
	tests := []struct {
		name, body, champ string
	}{
		{"wrong mot de passe", `{"motDePasse": "mauvais-mot", "titre": "Vélo", "description": "Cadre carbone, photos.", "prix": 1, "categorie": "loisirs", "ville": "Lyon", "codePostal": "69003"}`, "motDePasse"},
		{"titre", `{"motDePasse": "demo1234", "titre": "ab", "description": "Cadre carbone, photos.", "prix": 1, "categorie": "loisirs", "ville": "Lyon", "codePostal": "69003"}`, "titre"},
		{"description", `{"motDePasse": "demo1234", "titre": "Vélo", "description": "court", "prix": 1, "categorie": "loisirs", "ville": "Lyon", "codePostal": "69003"}`, "description"},
		{"missing prix", `{"motDePasse": "demo1234", "titre": "Vélo", "description": "Cadre carbone, photos.", "categorie": "loisirs", "ville": "Lyon", "codePostal": "69003"}`, "prix"},
		{"negative prix", `{"motDePasse": "demo1234", "titre": "Vélo", "description": "Cadre carbone, photos.", "prix": -1, "categorie": "loisirs", "ville": "Lyon", "codePostal": "69003"}`, "prix"},
		{"categorie", `{"motDePasse": "demo1234", "titre": "Vélo", "description": "Cadre carbone, photos.", "prix": 1, "categorie": "x", "ville": "Lyon", "codePostal": "69003"}`, "categorie"},
		{"ville", `{"motDePasse": "demo1234", "titre": "Vélo", "description": "Cadre carbone, photos.", "prix": 1, "categorie": "loisirs", "ville": "", "codePostal": "69003"}`, "ville"},
		{"codePostal", `{"motDePasse": "demo1234", "titre": "Vélo", "description": "Cadre carbone, photos.", "prix": 1, "categorie": "loisirs", "ville": "Lyon", "codePostal": "6"}`, "codePostal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mux := setupHandler(t)
			id := deposerUneAnnonceChallengee(t, repo, mux)

			rec := post(mux, "/annonces/"+id+"/correction", tt.body)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, tt.champ, decodeErreur(t, rec).Champ)
			stored, _ := repo.GetByID(id)
			assert.Equal(t, domain.Challengee, stored.Statut())
		})
	}
}

func TestCorrigerAnnonce_AnswersNotFoundUnlessTheAnnonceIsChallengee(t *testing.T) {
	repo, mux := setupHandler(t)
	activee := deposerUneAnnonce(t, repo, mux)

	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/"+activee+"/correction", correctionValide).Code)
	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/6f1c1c2e-8a43-4b7e-9d52-0a1b2c3d4e5f/correction", correctionValide).Code)
	assert.Equal(t, http.StatusBadRequest, post(mux, "/annonces/nope/correction", correctionValide).Code)
	assert.Equal(t, http.StatusBadRequest, post(mux, "/annonces/"+activee+"/correction", "pas du json").Code)
}
