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

	require.Eventually(t, func() bool { return len(spy.GetSentSimpleEmails()) == 1 },
		2*time.Second, 10*time.Millisecond, "the async bus delivers AnnonceDeposee to the consumer")
	email := spy.GetSentSimpleEmails()[0]
	assert.Equal(t, "lyonnaise@example.com", email.To)
	assert.Equal(t, "Votre annonce « Lampe de bureau articulée » est en cours de modération", email.Title)
	assert.Contains(t, email.Body, frontendURL+"/annonces/"+created.ID)
	assert.Contains(t, email.Body, frontendURL+"/annonces/"+created.ID+"/supprimer")
	assert.NotContains(t, email.Body, "secret12")
}

func TestServer_DeposeeAnnonceIsInvisibleWhileEnAttenteModeration(t *testing.T) {
	server, _, _ := startServer(t, false)
	resp := postJSON(t, server, "/annonces", depot)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+created.ID, nil))
	assert.Zero(t, rechercher(t, server, url.Values{"motCle": {"lampe"}}).Total)
	resp = postJSON(t, server, "/annonces/"+created.ID+"/suppression", `{"motDePasse": "secret12", "raison": "vendue-ailleurs"}`)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestServer_ListsTheRegions(t *testing.T) {
	server, _, _ := startServer(t, false)

	var regions []map[string]any
	require.Equal(t, http.StatusOK, getJSON(t, server, "/regions", &regions))
	assert.Len(t, regions, 18)
}

type fileModeration struct {
	Annonces []struct {
		ID            string `json:"id"`
		Titre         string `json:"titre"`
		VerrouillePar string `json:"verrouillePar"`
	} `json:"annonces"`
}

func TestServer_DeposedAnnonceReachesTheModerationQueueAndCanBeLockedThenReleased(t *testing.T) {
	server, _, _ := startServer(t, false)
	resp := postJSON(t, server, "/annonces", depot)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	var file fileModeration
	require.Eventually(t, func() bool {
		file = fileModeration{}
		getJSON(t, server, "/moderation/annonces", &file)
		return len(file.Annonces) == 1
	}, 2*time.Second, 10*time.Millisecond, "the async bus delivers AnnonceDeposee to the moderation consumer")
	assert.Equal(t, created.ID, file.Annonces[0].ID)
	assert.Equal(t, "Lampe de bureau articulée", file.Annonces[0].Titre)
	assert.Empty(t, file.Annonces[0].VerrouillePar)

	chemin := "/moderation/annonces/" + created.ID
	assert.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"alice"}`).StatusCode)
	assert.Equal(t, http.StatusConflict, postJSON(t, server, chemin+"/verrou", `{"moderateur":"bob"}`).StatusCode)
	var detail struct {
		Statut        string `json:"statut"`
		VerrouillePar string `json:"verrouillePar"`
		Historique    []struct {
			Type string `json:"type"`
		} `json:"historique"`
	}
	require.Equal(t, http.StatusOK, getJSON(t, server, chemin, &detail))
	assert.Equal(t, "verrouillee", detail.Statut)
	assert.Equal(t, "alice", detail.VerrouillePar)
	assert.Len(t, detail.Historique, 2)

	assert.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/relachement", `{"moderateur":"alice"}`).StatusCode)
	file = fileModeration{}
	require.Equal(t, http.StatusOK, getJSON(t, server, "/moderation/annonces", &file))
	require.Len(t, file.Annonces, 1)
	assert.Empty(t, file.Annonces[0].VerrouillePar)
	// Still invisible to acheteurs: moderation did not activate it.
	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+created.ID, nil))
}

func TestServer_ModerationAnswers404ForAnUnknownAnnonceAnd400ForABadModerateur(t *testing.T) {
	server, _, _ := startServer(t, false)

	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/moderation/annonces/inconnue", nil))
	assert.Equal(t, http.StatusNotFound, postJSON(t, server, "/moderation/annonces/inconnue/verrou", `{"moderateur":"alice"}`).StatusCode)
	assert.Equal(t, http.StatusBadRequest, postJSON(t, server, "/moderation/annonces/inconnue/verrou", `{"moderateur":"a"}`).StatusCode)
}

// deposeEtMetEnFile deposits an annonce and waits for it to reach the
// moderation queue, then returns its ID.
func deposeEtMetEnFile(t *testing.T, server *httptest.Server, spy *mailertesting.MailerSpy) string {
	t.Helper()
	resp := postJSON(t, server, "/annonces", depot)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.Eventually(t, func() bool {
		var file fileModeration
		getJSON(t, server, "/moderation/annonces", &file)
		return len(file.Annonces) == 1 && len(spy.GetSentSimpleEmails()) == 1
	}, 2*time.Second, 10*time.Millisecond, "the confirmation email and the queue")
	return created.ID
}

func statutModeration(t *testing.T, server *httptest.Server, id string) string {
	t.Helper()
	var detail struct {
		Statut string `json:"statut"`
	}
	require.Equal(t, http.StatusOK, getJSON(t, server, "/moderation/annonces/"+id, &detail))
	return detail.Statut
}

func TestServer_ValidationPutsTheAnnonceOnlineAndEmailsTheVendeur(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := deposeEtMetEnFile(t, server, spy)
	chemin := "/moderation/annonces/" + id
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"alice"}`).StatusCode)
	require.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+id, nil), "invisible before")

	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/validation", `{"moderateur":"alice"}`).StatusCode)

	require.Eventually(t, func() bool { return getJSON(t, server, "/annonces/"+id, nil) == http.StatusOK },
		2*time.Second, 10*time.Millisecond, "visible after")
	assert.Equal(t, 1, rechercher(t, server, url.Values{"motCle": {"lampe"}}).Total)
	require.Eventually(t, func() bool { return len(spy.GetSentSimpleEmails()) == 2 },
		2*time.Second, 10*time.Millisecond, "the validation email")
	email := spy.GetSentSimpleEmails()[1]
	assert.Equal(t, "lyonnaise@example.com", email.To)
	assert.Equal(t, "Votre annonce « Lampe de bureau articulée » est validée", email.Title)
	assert.Contains(t, email.Body, frontendURL+"/annonces/"+id)
	assert.Equal(t, "validee", statutModeration(t, server, id))
	var file fileModeration
	require.Equal(t, http.StatusOK, getJSON(t, server, "/moderation/annonces", &file))
	assert.Empty(t, file.Annonces, "out of the queue")
	assert.Equal(t, http.StatusConflict, postJSON(t, server, chemin+"/validation", `{"moderateur":"alice"}`).StatusCode, "no second decision")
}

func TestServer_RefusalKeepsTheAnnonceInvisibleAndSendsNoEmail(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := deposeEtMetEnFile(t, server, spy)
	chemin := "/moderation/annonces/" + id
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"alice"}`).StatusCode)

	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/refus", `{"moderateur":"alice"}`).StatusCode)

	assert.Equal(t, "refusee", statutModeration(t, server, id))
	// The consumer runs on the async bus: a refusal that came through has
	// already marked the annonce, which a later deletion attempt cannot reach.
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+id, nil))
	assert.Zero(t, rechercher(t, server, nil).Total)
	assert.Len(t, spy.GetSentSimpleEmails(), 1, "only the confirmation de dépôt")
}

func TestServer_ChallengeEmailsTheMotifAndTheCorrectionLinkAndKeepsTheAnnonceInvisible(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := deposeEtMetEnFile(t, server, spy)
	chemin := "/moderation/annonces/" + id
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"alice"}`).StatusCode)

	require.Equal(t, http.StatusNoContent,
		postJSON(t, server, chemin+"/challenge", `{"moderateur":"alice","motif":"Photos manquantes"}`).StatusCode)

	require.Eventually(t, func() bool { return len(spy.GetSentSimpleEmails()) == 2 },
		2*time.Second, 10*time.Millisecond, "the challenge email")
	email := spy.GetSentSimpleEmails()[1]
	assert.Equal(t, "lyonnaise@example.com", email.To)
	assert.Contains(t, email.Body, "Photos manquantes")
	assert.Contains(t, email.Body, frontendURL+"/annonces/"+id+"/corriger")
	assert.Equal(t, "challengee", statutModeration(t, server, id))
	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+id, nil))
	assert.Zero(t, rechercher(t, server, nil).Total)
	var file fileModeration
	require.Equal(t, http.StatusOK, getJSON(t, server, "/moderation/annonces", &file))
	assert.Empty(t, file.Annonces)
}

func TestServer_DecisionsAnswerLockErrorsAndAnInvalidMotif(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := deposeEtMetEnFile(t, server, spy)
	chemin := "/moderation/annonces/" + id

	for _, decision := range []struct{ route, corps string }{
		{"/validation", `{"moderateur":"alice"}`},
		{"/refus", `{"moderateur":"alice"}`},
		{"/challenge", `{"moderateur":"alice","motif":"Photos manquantes"}`},
	} {
		assert.Equal(t, http.StatusConflict, postJSON(t, server, chemin+decision.route, decision.corps).StatusCode, decision.route+" without lock")
	}
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"alice"}`).StatusCode)
	assert.Equal(t, http.StatusConflict, postJSON(t, server, chemin+"/validation", `{"moderateur":"bob"}`).StatusCode, "another moderator")
	resp := postJSON(t, server, chemin+"/challenge", `{"moderateur":"alice","motif":"court"}`)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var erreur struct {
		Champ string `json:"champ"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&erreur))
	assert.Equal(t, "motif", erreur.Champ)
	assert.Equal(t, "verrouillee", statutModeration(t, server, id), "nothing was decided")
	assert.Equal(t, http.StatusNotFound, postJSON(t, server, "/moderation/annonces/inconnue/validation", `{"moderateur":"alice"}`).StatusCode)
	assert.Len(t, spy.GetSentSimpleEmails(), 1)
}

const correction = `{
	"motDePasse": "secret12",
	"titre": "Lampe de bureau révisée",
	"description": "Ampoule LED fournie, photos ajoutées.",
	"prix": 2000,
	"categorie": "maison-jardin",
	"ville": "Lyon",
	"codePostal": "69003"
}`

// challengeEtAttend deposits an annonce, has alice challenge it and waits for
// the annonce context to have marked it challengee.
func challengeEtAttend(t *testing.T, server *httptest.Server, spy *mailertesting.MailerSpy) string {
	t.Helper()
	id := deposeEtMetEnFile(t, server, spy)
	chemin := "/moderation/annonces/" + id
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"alice"}`).StatusCode)
	require.Equal(t, http.StatusNoContent,
		postJSON(t, server, chemin+"/challenge", `{"moderateur":"alice","motif":"Photos manquantes"}`).StatusCode)
	require.Eventually(t, func() bool {
		return postJSON(t, server, "/annonces/"+id+"/edition", `{"motDePasse":"secret12"}`).StatusCode == http.StatusOK
	}, 2*time.Second, 10*time.Millisecond, "the annonce marked challengee")
	return id
}

func TestServer_EditionPrefillsTheFormWithTheMotifOnlyForAChallengeeAnnonce(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := challengeEtAttend(t, server, spy)

	resp := postJSON(t, server, "/annonces/"+id+"/edition", `{"motDePasse":"secret12"}`)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var edition map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&edition))
	assert.Equal(t, "Lampe de bureau articulée", edition["titre"])
	assert.Equal(t, "Photos manquantes", edition["motif"])
	assert.EqualValues(t, 2500, edition["prix"])
	assert.NotContains(t, edition, "email")
	wrong := postJSON(t, server, "/annonces/"+id+"/edition", `{"motDePasse":"mauvais-mot"}`)
	assert.Equal(t, http.StatusBadRequest, wrong.StatusCode)
	assert.Equal(t, http.StatusNotFound, postJSON(t, server, "/annonces/6f1c1c2e-8a43-4b7e-9d52-0a1b2c3d4e5f/edition", `{"motDePasse":"secret12"}`).StatusCode)
}

func TestServer_CorrectionSendsTheAnnonceBackToTheQueueWithItsFullHistoryThenARevalidationPutsItOnline(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := challengeEtAttend(t, server, spy)

	require.Equal(t, http.StatusNoContent, postJSON(t, server, "/annonces/"+id+"/correction", correction).StatusCode)

	var file fileModeration
	require.Eventually(t, func() bool {
		getJSON(t, server, "/moderation/annonces", &file)
		return len(file.Annonces) == 1
	}, 2*time.Second, 10*time.Millisecond, "back in the queue")
	assert.Equal(t, id, file.Annonces[0].ID)
	assert.Equal(t, "Lampe de bureau révisée", file.Annonces[0].Titre)
	var detail struct {
		Statut     string `json:"statut"`
		Historique []struct {
			Type string `json:"type"`
		} `json:"historique"`
	}
	require.Equal(t, http.StatusOK, getJSON(t, server, "/moderation/annonces/"+id, &detail))
	assert.Equal(t, "en-file", detail.Statut)
	var types []string
	for _, e := range detail.Historique {
		types = append(types, e.Type)
	}
	assert.Equal(t, []string{"AnnonceMiseEnFile", "AnnonceVerrouillee", "AnnonceChallengee", "AnnonceMiseEnFile"}, types)
	assert.Equal(t, http.StatusNotFound, getJSON(t, server, "/annonces/"+id, nil), "still invisible")
	assert.Equal(t, http.StatusNotFound, postJSON(t, server, "/annonces/"+id+"/edition", `{"motDePasse":"secret12"}`).StatusCode, "no longer challengee")
	assert.Equal(t, http.StatusNotFound, postJSON(t, server, "/annonces/"+id+"/correction", correction).StatusCode, "no second correction")

	chemin := "/moderation/annonces/" + id
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"bob"}`).StatusCode)
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/validation", `{"moderateur":"bob"}`).StatusCode)
	require.Eventually(t, func() bool { return getJSON(t, server, "/annonces/"+id, nil) == http.StatusOK },
		2*time.Second, 10*time.Millisecond, "online after the second validation")
	var annonce struct {
		Titre string `json:"titre"`
		Prix  int64  `json:"prix"`
	}
	getJSON(t, server, "/annonces/"+id, &annonce)
	assert.Equal(t, "Lampe de bureau révisée", annonce.Titre)
	assert.EqualValues(t, 2000, annonce.Prix)
}

func TestServer_CorrectionAnswersTheDepotErrorsAndNotFoundForAnAnnonceThatIsNotChallengee(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := challengeEtAttend(t, server, spy)

	resp := postJSON(t, server, "/annonces/"+id+"/correction", strings.Replace(correction, "Lampe de bureau révisée", "ab", 1))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var erreur struct {
		Champ string `json:"champ"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&erreur))
	assert.Equal(t, "titre", erreur.Champ)
	assert.Equal(t, http.StatusBadRequest, postJSON(t, server, "/annonces/"+id+"/correction", strings.Replace(correction, "secret12", "mauvais-mot", 1)).StatusCode)
	assert.Equal(t, http.StatusBadRequest, postJSON(t, server, "/annonces/nope/correction", correction).StatusCode)
	assert.Equal(t, http.StatusNotFound, postJSON(t, server, "/annonces/6f1c1c2e-8a43-4b7e-9d52-0a1b2c3d4e5f/correction", correction).StatusCode)
	var file fileModeration
	require.Equal(t, http.StatusOK, getJSON(t, server, "/moderation/annonces", &file))
	assert.Empty(t, file.Annonces, "nothing went back to the queue")
}

func TestServer_AValidatedOrRefusedAnnonceCannotBeCorrected(t *testing.T) {
	server, spy, _ := startServer(t, false)
	id := deposeEtMetEnFile(t, server, spy)
	chemin := "/moderation/annonces/" + id
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/verrou", `{"moderateur":"alice"}`).StatusCode)
	require.Equal(t, http.StatusNoContent, postJSON(t, server, chemin+"/refus", `{"moderateur":"alice"}`).StatusCode)
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, http.StatusNotFound, postJSON(t, server, "/annonces/"+id+"/correction", correction).StatusCode)

	assert.Equal(t, "refusee", statutModeration(t, server, id))
}
