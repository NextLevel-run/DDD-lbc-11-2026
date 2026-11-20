package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func saveVelo(t *testing.T, repo *inmemory.InMemoryAnnonceRepository, titre string, prix int64, deposeeLe time.Time) *domain.Annonce {
	t.Helper()
	annonce, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: titre, Description: "Bon état général, à venir chercher.", Prix: prix,
		Categorie: "loisirs", Ville: "Saint-Étienne", CodePostal: "42000",
		Pseudo: "stephanois", MotDePasse: "demo1234", Email: "stephanois@example.com", DeposeeLe: deposeeLe,
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, repo.Save(annonce))
	return annonce
}

func TestRechercherAnnonces_AnswersAPageOfResultsInFrenchJSON(t *testing.T) {
	repo, mux := setupHandler(t)
	velo := saveVelo(t, repo, "Vélo Électrique", 0, time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC))
	saveVelo(t, repo, "Trottinette", 5000, time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC))

	rec := get(mux, "/annonces?motCle=velo+electrique")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		"annonces": [{
			"id": "`+velo.ID()+`",
			"titre": "Vélo Électrique",
			"prix": 0,
			"ville": "Saint-Étienne",
			"images": [],
			"deposeeLe": "2026-11-02T10:00:00Z"
		}],
		"total": 1,
		"page": 1,
		"parPage": 20
	}`, rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "stephanois")
}

func TestRechercherAnnonces_AnswersTheRequestedPage(t *testing.T) {
	repo, mux := setupHandler(t)
	for i := range 21 {
		saveVelo(t, repo, fmt.Sprintf("Vélo %02d", i), 1000, maintenant.Add(time.Duration(i)*time.Minute))
	}

	rec := get(mux, "/annonces?page=2")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"total": 21, "page": 2, "parPage": 20}`, withoutAnnonces(t, rec.Body.Bytes()))
	assert.Contains(t, rec.Body.String(), `"Vélo 00"`)
	assert.NotContains(t, rec.Body.String(), `"Vélo 01"`)
}

func TestRechercherAnnonces_AnswersAnEmptyListWhenNothingMatches(t *testing.T) {
	_, mux := setupHandler(t)

	rec := get(mux, "/annonces?motCle=licorne")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"annonces": [], "total": 0, "page": 1, "parPage": 20}`, rec.Body.String())
}

func TestRechercherAnnonces_AnswersAnEmptyListForAPageFarPastTheLast(t *testing.T) {
	repo, mux := setupHandler(t)
	saveVelo(t, repo, "Vélo de ville", 1000, maintenant)

	rec := get(mux, "/annonces?page=9223372036854775807")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"annonces": [], "total": 1, "page": 9223372036854775807, "parPage": 20}`, rec.Body.String())
}

func TestRechercherAnnonces_AnswersBadRequestForAnInvalidPage(t *testing.T) {
	for _, page := range []string{"-1", "deux", "1.5"} {
		t.Run(page, func(t *testing.T) {
			_, mux := setupHandler(t)

			rec := get(mux, "/annonces?page="+page)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			body := decodeErreur(t, rec)
			assert.Equal(t, "Le numéro de page doit être un entier positif", body.Erreur)
			assert.Equal(t, "page", body.Champ)
		})
	}
}

func withoutAnnonces(t *testing.T, body []byte) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	delete(m, "annonces")
	out, err := json.Marshal(m)
	require.NoError(t, err)
	return string(out)
}

func TestRechercherAnnonces_FiltersByCategorieAndPrixFromTheQueryString(t *testing.T) {
	repo, mux := setupHandler(t)
	saveVelo(t, repo, "Vélo donné", 0, maintenant)
	saveVelo(t, repo, "Vélo de ville", 15000, maintenant.Add(time.Minute))
	saveVelo(t, repo, "Vélo de course", 90000, maintenant.Add(2*time.Minute))

	rec := get(mux, "/annonces?motCle=velo&categorie=loisirs&prixMin=10000&prixMax=20000")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"Vélo de ville"`)
	assert.JSONEq(t, `{"total": 1, "page": 1, "parPage": 20}`, withoutAnnonces(t, rec.Body.Bytes()))

	rec = get(mux, "/annonces?prixMin=0&prixMax=0")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"Vélo donné"`)
	assert.JSONEq(t, `{"total": 1, "page": 1, "parPage": 20}`, withoutAnnonces(t, rec.Body.Bytes()))

	rec = get(mux, "/annonces?categorie=vehicules")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"annonces": [], "total": 0, "page": 1, "parPage": 20}`, rec.Body.String())
}

func TestRechercherAnnonces_AnswersBadRequestForInvalidCriteria(t *testing.T) {
	tests := []struct {
		query   string
		champ   string
		message string
	}{
		{"categorie=licornes", "categorie", "Catégorie inconnue"},
		{"prixMin=2000&prixMax=1000", "prixMin", "Le prix minimum ne peut pas dépasser le prix maximum"},
		{"prixMin=-1", "prixMin", "Le prix minimum doit être un montant positif ou nul"},
		{"prixMin=dix", "prixMin", "Le prix minimum doit être un montant positif ou nul"},
		{"prixMax=-1", "prixMax", "Le prix maximum doit être un montant positif ou nul"},
		{"prixMax=12.5", "prixMax", "Le prix maximum doit être un montant positif ou nul"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			_, mux := setupHandler(t)

			rec := get(mux, "/annonces?"+tt.query)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			body := decodeErreur(t, rec)
			assert.Equal(t, tt.message, body.Erreur)
			assert.Equal(t, tt.champ, body.Champ)
		})
	}
}

func saveAnnonceA(t *testing.T, repo *inmemory.InMemoryAnnonceRepository, titre, ville, codePostal string) {
	t.Helper()
	annonce, err := domain.NewAnnonce(domain.NewAnnonceArgs{
		Titre: titre, Description: "Bon état général, à venir chercher.", Prix: 1000,
		Categorie: "maison-jardin", Ville: ville, CodePostal: codePostal,
		Pseudo: "voisin", MotDePasse: "demo1234", Email: "voisin@example.com", DeposeeLe: maintenant,
	})
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, repo.Save(annonce))
}

func TestRechercherAnnonces_FiltersByLieuFromTheQueryString(t *testing.T) {
	repo, mux := setupHandler(t)
	saveAnnonceA(t, repo, "Lampe dionysienne", "Saint-Denis", "93200")
	saveAnnonceA(t, repo, "Lampe réunionnaise", "Saint-Denis", "97400")
	saveAnnonceA(t, repo, "Lampe ajaccienne", "Ajaccio", "20000")

	tests := []struct {
		query string
		titre string
	}{
		{"region=04&villeOuCodePostal=saint+denis", "Lampe réunionnaise"},
		{"departement=93&villeOuCodePostal=Saint-Denis", "Lampe dionysienne"},
		{"departement=2A", "Lampe ajaccienne"},
		{"villeOuCodePostal=97400", "Lampe réunionnaise"},
		{"motCle=lampe&categorie=maison-jardin&region=94", "Lampe ajaccienne"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			rec := get(mux, "/annonces?"+tt.query)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"`+tt.titre+`"`)
			assert.JSONEq(t, `{"total": 1, "page": 1, "parPage": 20}`, withoutAnnonces(t, rec.Body.Bytes()))
		})
	}

	rec := get(mux, "/annonces?departement=2A&villeOuCodePostal=Saint-Denis")

	require.Equal(t, http.StatusOK, rec.Code, "an incoherent lieu is no error")
	assert.JSONEq(t, `{"annonces": [], "total": 0, "page": 1, "parPage": 20}`, rec.Body.String())
}

func TestRechercherAnnonces_AnswersBadRequestForAnInvalidLieu(t *testing.T) {
	tests := []struct {
		query   string
		champ   string
		message string
	}{
		{"region=99", "region", "Région inconnue"},
		{"departement=975", "departement", "Département inconnu"},
		{"region=04&departement=974", "region", "Choisissez une région ou un département, pas les deux"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			_, mux := setupHandler(t)

			rec := get(mux, "/annonces?"+tt.query)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			body := decodeErreur(t, rec)
			assert.Equal(t, tt.message, body.Erreur)
			assert.Equal(t, tt.champ, body.Champ)
		})
	}
}

func TestListerRegions_AnswersEveryRegionWithItsDepartements(t *testing.T) {
	_, mux := setupHandler(t)

	rec := get(mux, "/regions")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var regions []RegionAvecDepartementsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &regions))
	require.Len(t, regions, 18)
	assert.Contains(t, regions, RegionAvecDepartementsResponse{
		Code: "94", Nom: "Corse",
		Departements: []DepartementResponse{{Code: "2A", Nom: "Corse-du-Sud"}, {Code: "2B", Nom: "Haute-Corse"}},
	})
	assert.Contains(t, rec.Body.String(), `"departements":[{"code":"971","nom":"Guadeloupe"}]`)
}
