package http

import (
	"errors"
	"net/http"
	"testing"

	"ddd-second-hand-marketplace/pkg/mailer"
	mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const contactValide = `{"pseudo": "acheteur42", "email": "acheteur@example.com", "message": "Est-il toujours disponible ?"}`

type brokenMailer struct{}

func (brokenMailer) SendEmail(mailer.Email) error { return errors.New("smtp down") }
func (brokenMailer) SendTemplatedEmail(string, string, string, map[string]any) error {
	return errors.New("smtp down")
}

func TestContacter_AnswersNoContentAndEmailsTheVendeur(t *testing.T) {
	spy := mailertesting.NewMailerSpy()
	_, mux := setupHandlerWithMailer(t, spy)
	id := deposerUneAnnonce(t, mux)

	rec := post(mux, "/annonces/"+id+"/contact", contactValide)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	sent := spy.GetSentEmails()
	require.Len(t, sent, 1)
	assert.Equal(t, "acheteur@example.com", sent[0].ReplyTo)
}

func TestContacter_AnswersBadRequestNamingTheField(t *testing.T) {
	tests := []struct {
		name, body, champ string
	}{
		{"pseudo", `{"pseudo": "a", "email": "acheteur@example.com", "message": "Est-il toujours disponible ?"}`, "pseudo"},
		{"email", `{"pseudo": "acheteur42", "email": "pas-un-email", "message": "Est-il toujours disponible ?"}`, "email"},
		{"message", `{"pseudo": "acheteur42", "email": "acheteur@example.com", "message": "  court  "}`, "message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := mailertesting.NewMailerSpy()
			_, mux := setupHandlerWithMailer(t, spy)
			id := deposerUneAnnonce(t, mux)

			rec := post(mux, "/annonces/"+id+"/contact", tt.body)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), `"champ":"`+tt.champ+`"`)
			assert.Empty(t, spy.GetSentEmails())
		})
	}
}

func TestContacter_AnswersNotFoundOrBadRequestForTheAnnonce(t *testing.T) {
	_, mux := setupHandler(t)
	id := deposerUneAnnonce(t, mux)
	require.Equal(t, http.StatusNoContent, post(mux, "/annonces/"+id+"/suppression", `{"motDePasse": "demo1234", "raison": "autre"}`).Code)

	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/"+id+"/contact", contactValide).Code)
	assert.Equal(t, http.StatusNotFound, post(mux, "/annonces/00000000-0000-4000-8000-000000000000/contact", contactValide).Code)
	assert.Equal(t, http.StatusBadRequest, post(mux, "/annonces/pas-un-uuid/contact", contactValide).Code)
}

func TestContacter_AnswersBadGatewayWhenTheEmailCannotLeave(t *testing.T) {
	_, mux := setupHandlerWithMailer(t, brokenMailer{})
	id := deposerUneAnnonce(t, mux)

	rec := post(mux, "/annonces/"+id+"/contact", contactValide)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.JSONEq(t, `{"erreur": "Votre message n'a pas pu être envoyé, réessayez plus tard."}`, rec.Body.String())
}
