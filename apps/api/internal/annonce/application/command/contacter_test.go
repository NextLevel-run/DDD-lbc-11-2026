package command

import (
	"errors"
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/annonce/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/annonce/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
	"ddd-second-hand-marketplace/pkg/mailer"
	mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var soumisLe = maintenant.Add(24 * time.Hour)

type contacterSetup struct {
	repo      *inmemory.InMemoryAnnonceRepository
	bus       eventbus.Bus
	collector *eventbustesting.EventCollector
	id        string
}

func setupContacter(t *testing.T) contacterSetup {
	t.Helper()
	repo := inmemory.NewInMemoryAnnonceRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, bus.Subscribe(domain.FormulaireContactSoumisEventType, collector.EventHandler()))
	id, err := BuildDeposerAnnonceCommand(repo, bus, func() time.Time { return maintenant })(validCommandArgs())
	require.NoError(t, err)
	return contacterSetup{repo: repo, bus: bus, collector: collector, id: id}
}

func (s contacterSetup) contacter(m mailer.Mailer) ContacterCommand {
	return BuildContacterCommand(s.repo, m, s.bus, func() time.Time { return soumisLe }, "noreply@example.com", "http://front.test/")
}

func (s contacterSetup) validArgs() ContacterCommandArgs {
	return ContacterCommandArgs{ID: s.id, Pseudo: "acheteur42", Email: "acheteur@example.com", Message: "  Est-il toujours disponible ?  "}
}

type brokenMailer struct{}

func (brokenMailer) SendEmail(mailer.Email) error { return errors.New("smtp down") }
func (brokenMailer) SendTemplatedEmail(string, string, string, map[string]any) error {
	return errors.New("smtp down")
}

func TestContacter_EmailsTheVendeurWithTheAcheteurAsReplyTo(t *testing.T) {
	s := setupContacter(t)
	spy := mailertesting.NewMailerSpy()

	require.NoError(t, s.contacter(spy)(s.validArgs()))

	sent := spy.GetSentEmails()
	require.Len(t, sent, 1)
	email := sent[0]
	assert.Equal(t, "cycliste@example.com", email.To)
	assert.Equal(t, "noreply@example.com", email.From)
	assert.Equal(t, "acheteur@example.com", email.ReplyTo)
	assert.Equal(t, "Nouveau message pour votre annonce « Vélo de course »", email.Title)
	for _, want := range []string{
		"acheteur42 vous a écrit :",
		"Est-il toujours disponible ?",
		"Répondez directement à ce mail pour lui écrire.",
		"http://front.test/annonces/" + s.id,
	} {
		assert.Contains(t, email.Body, want)
	}
	assert.NotContains(t, email.Body, "acheteur@example.com", "the acheteur's email stays out of the body")
}

func TestContacter_PublishesFormulaireContactSoumis(t *testing.T) {
	s := setupContacter(t)

	require.NoError(t, s.contacter(mailertesting.NewMailerSpy())(s.validArgs()))

	events := s.collector.GetEvents()
	require.Len(t, events, 1)
	event, ok := events[0].(*domain.FormulaireContactSoumisEvent)
	require.True(t, ok)
	assert.Equal(t, "FormulaireContactSoumis", event.EventType())
	assert.Equal(t, s.id, event.AnnonceID)
	assert.Equal(t, "acheteur42", event.Pseudo)
	assert.Equal(t, "acheteur@example.com", event.Email)
	assert.Equal(t, "Est-il toujours disponible ?", event.Message)
	assert.Equal(t, soumisLe, event.SoumisLe)
}

func TestContacter_SendFailureAnswersErrEnvoiImpossibleWithoutPublishing(t *testing.T) {
	s := setupContacter(t)

	err := s.contacter(brokenMailer{})(s.validArgs())

	assert.ErrorIs(t, err, ErrEnvoiImpossible)
	assert.Empty(t, s.collector.GetEvents())
}

func TestContacter_RefusesAnInvalidFormulaireWithoutSending(t *testing.T) {
	tests := []struct {
		name string
		edit func(*ContacterCommandArgs)
		want error
	}{
		{"pseudo", func(a *ContacterCommandArgs) { a.Pseudo = "a" }, domain.ErrPseudoInvalide},
		{"email", func(a *ContacterCommandArgs) { a.Email = "pas-un-email" }, domain.ErrEmailInvalide},
		{"message", func(a *ContacterCommandArgs) { a.Message = strings.Repeat(" ", 20) + "court" }, domain.ErrMessageInvalide},
		{"malformed ID", func(a *ContacterCommandArgs) { a.ID = "pas-un-uuid" }, domain.ErrIDAnnonceInvalide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupContacter(t)
			spy := mailertesting.NewMailerSpy()
			args := s.validArgs()
			tt.edit(&args)

			err := s.contacter(spy)(args)

			assert.ErrorIs(t, err, tt.want)
			assert.Empty(t, spy.GetSentEmails())
			assert.Empty(t, s.collector.GetEvents())
		})
	}
}

func TestContacter_AnswersIntrouvableForAnUnknownOrSupprimeeAnnonce(t *testing.T) {
	s := setupContacter(t)
	spy := mailertesting.NewMailerSpy()
	require.NoError(t, BuildSupprimerAnnonceCommand(s.repo, s.bus, time.Now)(SupprimerAnnonceCommandArgs{ID: s.id, MotDePasse: "demo1234", Raison: "autre"}))

	supprimee := s.contacter(spy)(s.validArgs())
	inconnue := s.contacter(spy)(ContacterCommandArgs{ID: "00000000-0000-4000-8000-000000000000", Pseudo: "acheteur42", Email: "acheteur@example.com", Message: "Est-il toujours disponible ?"})

	assert.ErrorIs(t, supprimee, domain.ErrAnnonceIntrouvable)
	assert.ErrorIs(t, inconnue, domain.ErrAnnonceIntrouvable)
	assert.Empty(t, spy.GetSentEmails())
	assert.Empty(t, s.collector.GetEvents())
}
