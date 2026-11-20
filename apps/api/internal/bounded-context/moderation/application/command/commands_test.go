package command

import (
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/bounded-context/moderation/domain"
	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func horloge(t *time.Time) func() time.Time { return func() time.Time { return *t } }

func unDepot() MettreEnFileCommandArgs {
	return MettreEnFileCommandArgs{
		AnnonceID: "a1", Titre: "Vélo de ville", Description: "Très bon état.", Prix: 15000,
		Categorie: "vehicules", Ville: "Lyon", CodePostal: "69003", Pseudo: "cyclo",
		DeposeeLe: maintenant.Add(-time.Hour),
	}
}

func setup(t *testing.T) (*inmemory.InMemoryHistoriqueAnnonceRepository, *time.Time) {
	t.Helper()
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	now := maintenant
	require.NoError(t, BuildMettreEnFileCommand(repo, horloge(&now))(unDepot()))
	return repo, &now
}

func TestMettreEnFile_StoresTheSnapshotInANewStream(t *testing.T) {
	repo, _ := setup(t)

	h, err := repo.Charger("a1")

	require.NoError(t, err)
	assert.Equal(t, domain.EtatEnFile, h.Etat())
	assert.Equal(t, "Vélo de ville", h.Annonce().Titre)
	assert.Equal(t, int64(15000), h.Annonce().Prix)
	assert.Equal(t, maintenant, h.MisEnFileLe())
}

func TestMettreEnFile_IsIdempotent(t *testing.T) {
	repo, now := setup(t)

	require.NoError(t, BuildMettreEnFileCommand(repo, horloge(now))(unDepot()))

	h, _ := repo.Charger("a1")
	assert.Len(t, h.Evenements(), 1)
}

func TestVerrouillerAnnonce_LocksForTheModerateur(t *testing.T) {
	repo, now := setup(t)

	err := BuildVerrouillerAnnonceCommand(repo, horloge(now))(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "alice"})

	require.NoError(t, err)
	h, _ := repo.Charger("a1")
	moderateur, _, ok := h.Verrou(*now)
	assert.True(t, ok)
	assert.Equal(t, "alice", moderateur)
}

func TestVerrouillerAnnonce_RefusesAnotherModerateurThenTakesOverAfterTwoHours(t *testing.T) {
	repo, now := setup(t)
	verrouiller := BuildVerrouillerAnnonceCommand(repo, horloge(now))
	require.NoError(t, verrouiller(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "alice"}))

	err := verrouiller(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "bob"})
	assert.ErrorIs(t, err, domain.ErrAnnonceDejaVerrouillee)

	*now = now.Add(2*time.Hour + time.Minute)
	require.NoError(t, verrouiller(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "bob"}))
	h, _ := repo.Charger("a1")
	moderateur, _, _ := h.Verrou(*now)
	assert.Equal(t, "bob", moderateur)
	assert.Len(t, h.Evenements(), 4) // mise en file, lock, expiration, lock
}

func TestVerrouillerAnnonce_RejectsAnInvalidModerateurAndAnUnknownAnnonce(t *testing.T) {
	repo, now := setup(t)
	verrouiller := BuildVerrouillerAnnonceCommand(repo, horloge(now))

	assert.ErrorIs(t, verrouiller(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "x"}), domain.ErrModerateurInvalide)
	assert.ErrorIs(t, verrouiller(VerrouillerAnnonceCommandArgs{AnnonceID: "inconnue", Moderateur: "alice"}), domain.ErrAnnonceIntrouvable)
	h, _ := repo.Charger("a1")
	assert.Len(t, h.Evenements(), 1)
}

func TestRelacherVerrou_PutsTheAnnonceBackInTheQueue(t *testing.T) {
	repo, now := setup(t)
	require.NoError(t, BuildVerrouillerAnnonceCommand(repo, horloge(now))(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "alice"}))

	err := BuildRelacherVerrouCommand(repo, horloge(now))(RelacherVerrouCommandArgs{AnnonceID: "a1", Moderateur: "alice"})

	require.NoError(t, err)
	h, _ := repo.Charger("a1")
	assert.Equal(t, domain.EtatEnFile, h.Etat())
}

func TestRelacherVerrou_IsRefusedToAnyoneButTheHolderOfAValidLock(t *testing.T) {
	repo, now := setup(t)
	relacher := BuildRelacherVerrouCommand(repo, horloge(now))
	require.NoError(t, BuildVerrouillerAnnonceCommand(repo, horloge(now))(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "alice"}))

	assert.ErrorIs(t, relacher(RelacherVerrouCommandArgs{AnnonceID: "a1", Moderateur: "bob"}), domain.ErrAnnoncePasVerrouilleeParModerateur)
	assert.ErrorIs(t, relacher(RelacherVerrouCommandArgs{AnnonceID: "a1", Moderateur: "x"}), domain.ErrModerateurInvalide)
	*now = now.Add(3 * time.Hour)
	assert.ErrorIs(t, relacher(RelacherVerrouCommandArgs{AnnonceID: "a1", Moderateur: "alice"}), domain.ErrVerrouExpire)
	assert.ErrorIs(t, relacher(RelacherVerrouCommandArgs{AnnonceID: "inconnue", Moderateur: "alice"}), domain.ErrAnnonceIntrouvable)
}

type decisionCase struct {
	nom       string
	eventType string
	// contract is the event the decision publishes, decided at date.
	contract func(date time.Time) eventbus.DomainEvent
	decider  func(repo domain.HistoriqueAnnonceRepository, bus eventbus.Bus, now func() time.Time, moderateur string) error
}

func decisions() []decisionCase {
	return []decisionCase{
		{"valider", "AnnonceValidee",
			func(date time.Time) eventbus.DomainEvent {
				return contracts.AnnonceValideeEvent{AnnonceID: "a1", Date: date}
			},
			func(repo domain.HistoriqueAnnonceRepository, bus eventbus.Bus, now func() time.Time, m string) error {
				return BuildValiderAnnonceCommand(repo, bus, now)(ValiderAnnonceCommandArgs{AnnonceID: "a1", Moderateur: m})
			}},
		{"refuser", "AnnonceRefusee",
			func(date time.Time) eventbus.DomainEvent {
				return contracts.AnnonceRefuseeEvent{AnnonceID: "a1", Date: date}
			},
			func(repo domain.HistoriqueAnnonceRepository, bus eventbus.Bus, now func() time.Time, m string) error {
				return BuildRefuserAnnonceCommand(repo, bus, now)(RefuserAnnonceCommandArgs{AnnonceID: "a1", Moderateur: m})
			}},
		{"challenger", "AnnonceChallengee",
			func(date time.Time) eventbus.DomainEvent {
				return contracts.AnnonceChallengeeEvent{AnnonceID: "a1", Motif: "Photos manquantes", Date: date}
			},
			func(repo domain.HistoriqueAnnonceRepository, bus eventbus.Bus, now func() time.Time, m string) error {
				return BuildChallengerAnnonceCommand(repo, bus, now)(ChallengerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: m, Motif: "Photos manquantes"})
			}},
	}
}

func collecte(t *testing.T) (eventbus.Bus, *eventbustesting.EventCollector) {
	t.Helper()
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	for _, eventType := range []string{"AnnonceValidee", "AnnonceRefusee", "AnnonceChallengee"} {
		require.NoError(t, bus.Subscribe(eventType, collector.EventHandler()))
	}
	return bus, collector
}

func verrouille(t *testing.T, repo domain.HistoriqueAnnonceRepository, now *time.Time, moderateur string) {
	t.Helper()
	require.NoError(t, BuildVerrouillerAnnonceCommand(repo, horloge(now))(VerrouillerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: moderateur}))
}

func TestDecisions_StoreTheEventThenPublishItWithItsAccessors(t *testing.T) {
	for _, d := range decisions() {
		t.Run(d.nom, func(t *testing.T) {
			repo, now := setup(t)
			verrouille(t, repo, now, "alice")
			bus, collector := collecte(t)

			require.NoError(t, d.decider(repo, bus, horloge(now), "alice"))

			h, _ := repo.Charger("a1")
			assert.False(t, h.EstAModerer(), "the annonce left the queue")
			stream := h.Evenements()
			assert.Equal(t, d.eventType, stream[len(stream)-1].EventType(), "the decision stays in the stream")
			require.Len(t, collector.GetEvents(), 1)
			published := collector.GetEvents()[0]
			assert.Equal(t, d.eventType, published.EventType())
			assert.Equal(t, d.contract(*now), published, "the contract carries the annonce, the date and, for a challenge, the motif")
		})
	}
}

func TestDecisions_AreRefusedWithoutAValidLockAndPublishNothing(t *testing.T) {
	for _, d := range decisions() {
		t.Run(d.nom, func(t *testing.T) {
			repo, now := setup(t)
			bus, collector := collecte(t)

			assert.ErrorIs(t, d.decider(repo, bus, horloge(now), "alice"), domain.ErrAnnoncePasVerrouilleeParModerateur, "not locked")
			verrouille(t, repo, now, "alice")
			assert.ErrorIs(t, d.decider(repo, bus, horloge(now), "bob"), domain.ErrAnnoncePasVerrouilleeParModerateur, "locked by another")
			assert.ErrorIs(t, d.decider(repo, bus, horloge(now), "x"), domain.ErrModerateurInvalide)
			*now = now.Add(2*time.Hour + time.Minute)
			assert.ErrorIs(t, d.decider(repo, bus, horloge(now), "alice"), domain.ErrVerrouExpire)

			h, _ := repo.Charger("a1")
			assert.Len(t, h.Evenements(), 2, "mise en file and lock only")
			assert.Empty(t, collector.GetEvents())
		})
	}
}

func TestDecisions_OfAnUnknownAnnonceAreIntrouvable(t *testing.T) {
	repo := inmemory.NewInMemoryHistoriqueAnnonceRepository()
	now := maintenant
	for _, d := range decisions() {
		bus, _ := collecte(t)
		assert.ErrorIs(t, d.decider(repo, bus, horloge(&now), "alice"), domain.ErrAnnonceIntrouvable, d.nom)
	}
}

func TestChallengerAnnonce_RejectsAnInvalidMotifBeforeTouchingTheStream(t *testing.T) {
	repo, now := setup(t)
	verrouille(t, repo, now, "alice")
	bus, collector := collecte(t)
	challenger := BuildChallengerAnnonceCommand(repo, bus, horloge(now))

	for _, motif := range []string{"", "court", "   neuf    ", strings.Repeat("x", 1001)} {
		err := challenger(ChallengerAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "alice", Motif: motif})
		assert.ErrorIs(t, err, domain.ErrMotifInvalide, motif)
	}
	h, _ := repo.Charger("a1")
	assert.True(t, h.EstAModerer())
	assert.Empty(t, collector.GetEvents())
}

func TestDecision_DoesNotPublishWhenTheStoreRefusesTheAppend(t *testing.T) {
	repo, now := setup(t)
	verrouille(t, repo, now, "alice")
	bus, collector := collecte(t)
	stale := &staleRepository{HistoriqueAnnonceRepository: repo}

	err := BuildValiderAnnonceCommand(stale, bus, horloge(now))(ValiderAnnonceCommandArgs{AnnonceID: "a1", Moderateur: "alice"})

	assert.ErrorIs(t, err, domain.ErrConflitVersion)
	assert.Empty(t, collector.GetEvents())
}

// staleRepository loses the race: the stream always grew since the load.
type staleRepository struct {
	domain.HistoriqueAnnonceRepository
}

func (staleRepository) Enregistrer(*domain.HistoriqueAnnonce) error { return domain.ErrConflitVersion }
