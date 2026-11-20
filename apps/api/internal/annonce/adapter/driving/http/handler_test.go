package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/annonce/application/command"
	"ddd-second-hand-marketplace/internal/annonce/application/query"
	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func setupHandler(t *testing.T) (*inmemory.InMemoryAnnonceRepository, *http.ServeMux) {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	mux := http.NewServeMux()
	NewHandler(
		query.BuildConsulterAnnonceQuery(repo),
		query.BuildRechercherAnnoncesQuery(repo),
		query.BuildListerRegionsQuery(),
		command.BuildDeposerAnnonceCommand(repo, bus, func() time.Time { return maintenant }),
		command.BuildSupprimerAnnonceCommand(repo, bus, func() time.Time { return maintenant }),
		query.BuildConsulterAnnonceAEditerQuery(repo),
		command.BuildCorrigerAnnonceCommand(repo, bus, func() time.Time { return maintenant }),
	).RegisterRoutes(mux)
	return repo, mux
}

func post(mux *http.ServeMux, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	return rec
}

func get(mux *http.ServeMux, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestConsulterAnnonce_AnswersTheAnnonceInFrenchJSON(t *testing.T) {
	repo, mux := setupHandler(t)
	annonce, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: "Lot de livres", Description: "Romans de poche, à donner.", Prix: 0,
		Categorie: "loisirs", Ville: "Saint-Denis", CodePostal: "97400",
		Pseudo: "lectrice", MotDePasse: "demo1234", Email: "lectrice@example.com", Telephone: "0692000000",
		DeposeeLe: time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, repo.Save(annonce))

	rec := get(mux, "/annonces/"+annonce.ID())

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		"id": "`+annonce.ID()+`",
		"titre": "Lot de livres",
		"description": "Romans de poche, à donner.",
		"prix": 0,
		"categorie": {"code": "loisirs", "libelle": "Loisirs"},
		"localisation": {
			"ville": "Saint-Denis",
			"codePostal": "97400",
			"departement": {"code": "974", "nom": "La Réunion"},
			"region": {"code": "04", "nom": "La Réunion"}
		},
		"vendeur": {"pseudo": "lectrice"},
		"images": [],
		"deposeeLe": "2026-11-02T10:00:00Z"
	}`, rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "lectrice@example.com")
	assert.NotContains(t, rec.Body.String(), "0692000000")
}

func TestConsulterAnnonce_AnswersNotFoundForAnUnknownAnnonce(t *testing.T) {
	_, mux := setupHandler(t)

	rec := get(mux, "/annonces/"+uuid.NewString())

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assertErreur(t, rec, "Annonce introuvable")
}

func TestConsulterAnnonce_AnswersBadRequestForAMalformedID(t *testing.T) {
	_, mux := setupHandler(t)

	rec := get(mux, "/annonces/pas-un-uuid")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErreur(t, rec, "Identifiant d'annonce invalide")
}

const depotValide = `{
	"titre": "Vélo de course",
	"description": "Cadre carbone, très bon état.",
	"prix": 45000,
	"categorie": "loisirs",
	"ville": "Lyon",
	"codePostal": "69003",
	"pseudo": "cycliste69",
	"email": "cycliste@example.com",
	"telephone": "0612345678",
	"motDePasse": "demo1234"
}`

func TestDeposerAnnonce_AnswersCreatedWithTheNewID(t *testing.T) {
	repo, mux := setupHandler(t)

	rec := post(mux, "/annonces", depotValide)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var body DeposerAnnonceResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	stored, err := repo.GetByID(body.ID)
	require.NoError(t, err)
	assert.Equal(t, "Vélo de course", stored.Titre())
	assert.Equal(t, maintenant, stored.DeposeeLe())
	assert.Equal(t, "0612345678", stored.Vendeur().Telephone())
	assert.True(t, stored.MotDePasseCorrespond("demo1234"))
}

func TestDeposerAnnonce_NeverEchoesTheMotDePasse(t *testing.T) {
	_, mux := setupHandler(t)
	var body DeposerAnnonceResponse
	created := post(mux, "/annonces", depotValide)
	require.NoError(t, json.NewDecoder(created.Body).Decode(&body))

	rec := get(mux, "/annonces/"+body.ID)

	assert.NotContains(t, strings.ToLower(rec.Body.String()), "demo1234")
	assert.NotContains(t, strings.ToLower(rec.Body.String()), "motdepasse")
}

func TestDeposerAnnonce_TheNewAnnonceIsInvisibleUntilModerated(t *testing.T) {
	_, mux := setupHandler(t)
	var body DeposerAnnonceResponse
	require.NoError(t, json.NewDecoder(post(mux, "/annonces", depotValide).Body).Decode(&body))

	assert.Equal(t, http.StatusNotFound, get(mux, "/annonces/"+body.ID).Code)
	var recherche RechercherAnnoncesResponse
	require.NoError(t, json.NewDecoder(get(mux, "/annonces").Body).Decode(&recherche))
	assert.Zero(t, recherche.Total)
}

func TestDeposerAnnonce_AcceptsAMissingTelephone(t *testing.T) {
	_, mux := setupHandler(t)

	rec := post(mux, "/annonces", strings.Replace(depotValide, `"telephone": "0612345678"`, `"telephone": ""`, 1))

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestDeposerAnnonce_AnswersBadRequestNamingTheField(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		champ   string
		message string
	}{
		{"titre", `"titre": "Vélo de course"`, `"titre": "ab"`, "titre", "Le titre doit comporter de 3 à 100 caractères"},
		{"description", `"description": "Cadre carbone, très bon état."`, `"description": "court"`, "description", "La description doit comporter de 10 à 4 000 caractères"},
		{"prix", `"prix": 45000`, `"prix": -1`, "prix", "Le prix ne peut pas être négatif"},
		{"missing prix", `"prix": 45000,`, ``, "prix", "Le prix est obligatoire"},
		{"categorie", `"categorie": "loisirs"`, `"categorie": "bijoux"`, "categorie", "Catégorie inconnue"},
		{"ville", `"ville": "Lyon"`, `"ville": " "`, "ville", "La ville est obligatoire"},
		{"code postal", `"codePostal": "69003"`, `"codePostal": "690"`, "codePostal", "Le code postal doit comporter 5 chiffres"},
		{"departement", `"codePostal": "69003"`, `"codePostal": "96000"`, "codePostal", "Le code postal ne désigne aucun département"},
		{"pseudo", `"pseudo": "cycliste69"`, `"pseudo": "c"`, "pseudo", "Le pseudo doit comporter de 2 à 30 caractères"},
		{"email", `"email": "cycliste@example.com"`, `"email": "cycliste"`, "email", "L'email n'est pas valide"},
		{"mot de passe trop court", `"motDePasse": "demo1234"`, `"motDePasse": "court"`, "motDePasse", "Le mot de passe de l'annonce doit comporter au moins 8 caractères"},
		{"mot de passe manquant", `,
	"motDePasse": "demo1234"`, ``, "motDePasse", "Le mot de passe de l'annonce doit comporter au moins 8 caractères"},
		{"telephone", `"telephone": "0612345678"`, `"telephone": "12"`, "telephone", "Le téléphone doit comporter 10 chiffres et commencer par 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mux := setupHandler(t)

			rec := post(mux, "/annonces", strings.Replace(depotValide, tt.from, tt.to, 1))

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			body := decodeErreur(t, rec)
			assert.Equal(t, tt.message, body.Erreur)
			assert.Equal(t, tt.champ, body.Champ)
		})
	}
}

func TestDeposerAnnonce_AnswersBadRequestForAMalformedBody(t *testing.T) {
	_, mux := setupHandler(t)

	rec := post(mux, "/annonces", `{"prix": "cher"}`)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	body := decodeErreur(t, rec)
	assert.Equal(t, "Requête invalide", body.Erreur)
	assert.Empty(t, body.Champ)
}

func assertErreur(t *testing.T, rec *httptest.ResponseRecorder, message string) {
	t.Helper()
	assert.Equal(t, message, decodeErreur(t, rec).Erreur)
}

func decodeErreur(t *testing.T, rec *httptest.ResponseRecorder) ErreurResponse {
	t.Helper()
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var body ErreurResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}
