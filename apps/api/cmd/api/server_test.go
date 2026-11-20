package main

// Integration tests of the composition root: the real server of newServer,
// over HTTP, with the async event bus and a mailer spy. The handler tests cover
// each route; these check that main wires them together.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const frontendURL = "http://front.test"

func startServer(t *testing.T, demoData bool) (*httptest.Server, *mailertesting.MailerSpy, []string) {
	t.Helper()
	spy := mailertesting.NewMailerSpy()
	handler, demoIDs, err := newServer(serverConfig{Mailer: spy, DemoData: demoData, FrontendURL: frontendURL})
	require.NoError(t, err)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server, spy, demoIDs
}

func getJSON(t *testing.T, server *httptest.Server, path string, into any) int {
	t.Helper()
	resp, err := http.Get(server.URL + path)
	require.NoError(t, err)
	defer resp.Body.Close()
	if into != nil && resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(into))
	}
	return resp.StatusCode
}

func postJSON(t *testing.T, server *httptest.Server, path, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(server.URL+path, "application/json", strings.NewReader(body))
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

type recherche struct {
	Annonces []struct {
		ID    string `json:"id"`
		Titre string `json:"titre"`
	} `json:"annonces"`
	Total int `json:"total"`
}

func rechercher(t *testing.T, server *httptest.Server, params url.Values) recherche {
	t.Helper()
	var r recherche
	require.Equal(t, http.StatusOK, getJSON(t, server, "/annonces?"+params.Encode(), &r))
	return r
}

func titres(r recherche) []string {
	out := make([]string, 0, len(r.Annonces))
	for _, a := range r.Annonces {
		out = append(out, a.Titre)
	}
	return out
}

const depot = `{
	"titre": "Lampe de bureau articulée",
	"description": "Ampoule LED fournie, aucun défaut.",
	"prix": 2500,
	"categorie": "maison-jardin",
	"ville": "Lyon",
	"codePostal": "69003",
	"pseudo": "lyonnaise",
	"email": "lyonnaise@example.com",
	"motDePasse": "secret12"
}`

func TestServer_AnswersTheHealthCheck(t *testing.T) {
	server, _, _ := startServer(t, false)

	var body map[string]string
	require.Equal(t, http.StatusOK, getJSON(t, server, "/health", &body))
	assert.Equal(t, map[string]string{"status": "ok"}, body)
}

func TestServer_StartsEmptyWithoutDemoData(t *testing.T) {
	server, _, demoIDs := startServer(t, false)

	assert.Empty(t, demoIDs)
	assert.Zero(t, rechercher(t, server, nil).Total)
}

func TestServer_LoadsDemoAnnoncesThatCanBeConsultedAndSearched(t *testing.T) {
	server, _, demoIDs := startServer(t, true)

	require.NotEmpty(t, demoIDs)
	assert.Equal(t, len(demoIDs), rechercher(t, server, nil).Total)
	for _, id := range demoIDs {
		assert.Equal(t, http.StatusOK, getJSON(t, server, "/annonces/"+id, nil), id)
	}
	// The QA guides rely on these demo annonces: two homonyms and a don.
	assert.ElementsMatch(t,
		[]string{"Planche de surf 7'2", "Table de jardin en teck"},
		titres(rechercher(t, server, url.Values{"villeOuCodePostal": {"Saint-Denis"}})))
	assert.Equal(t,
		[]string{"Cartons de déménagement à donner"},
		titres(rechercher(t, server, url.Values{"prixMin": {"0"}, "prixMax": {"0"}})))
}

func TestServer_DemoAnnoncesAreSupprimeesWithTheDemoMotDePasse(t *testing.T) {
	server, _, demoIDs := startServer(t, true)

	resp := postJSON(t, server, "/annonces/"+demoIDs[0]+"/suppression", `{"motDePasse": "demo1234", "raison": "autre"}`)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+demoIDs[0], nil))
}

func TestServer_DepotSendsTheConfirmationEmailThroughTheEventBus(t *testing.T) {
	server, spy, _ := startServer(t, false)

	resp := postJSON(t, server, "/annonces", depot)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	require.Eventually(t, func() bool { return len(spy.GetSentEmails()) == 1 },
		2*time.Second, 10*time.Millisecond, "the async bus delivers AnnonceDeposee to the consumer")
	email := spy.GetSentEmails()[0]
	assert.Equal(t, "lyonnaise@example.com", email.To)
	assert.Equal(t, "Votre annonce « Lampe de bureau articulée » est en ligne", email.Title)
	assert.Contains(t, email.Body, frontendURL+"/annonces/"+created.ID)
	assert.Contains(t, email.Body, frontendURL+"/annonces/"+created.ID+"/supprimer")
	assert.NotContains(t, email.Body, "secret12")
}

func TestServer_DeposeeThenSupprimeeAnnonceGoesOffline(t *testing.T) {
	server, _, _ := startServer(t, false)
	resp := postJSON(t, server, "/annonces", depot)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	assert.Equal(t, http.StatusOK, getJSON(t, server, "/annonces/"+created.ID, nil))
	assert.Equal(t, 1, rechercher(t, server, url.Values{"motCle": {"LAMPE articulee"}}).Total)

	resp = postJSON(t, server, "/annonces/"+created.ID+"/suppression", `{"motDePasse": "secret12", "raison": "vendue-ailleurs"}`)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+created.ID, nil))
	assert.Zero(t, rechercher(t, server, url.Values{"motCle": {"lampe"}}).Total)
}

func TestServer_ListsTheRegions(t *testing.T) {
	server, _, _ := startServer(t, false)

	var regions []map[string]any
	require.Equal(t, http.StatusOK, getJSON(t, server, "/regions", &regions))
	assert.Len(t, regions, 18)
}

func TestServer_ContactEmailsTheVendeurWithTheAcheteurAsReplyTo(t *testing.T) {
	server, spy, demoIDs := startServer(t, true)
	require.NotEmpty(t, demoIDs)

	resp := postJSON(t, server, "/annonces/"+demoIDs[0]+"/contact",
		`{"pseudo": "acheteur42", "email": "acheteur@example.com", "message": "Est-il toujours disponible ?"}`)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	sent := spy.GetSentEmails()
	require.Len(t, sent, 1, "the contact email is sent synchronously")
	assert.Equal(t, "noreply@seconde-main.local", sent[0].From)
	assert.Equal(t, "acheteur@example.com", sent[0].ReplyTo)
	assert.Contains(t, sent[0].Body, frontendURL+"/annonces/"+demoIDs[0])
}
