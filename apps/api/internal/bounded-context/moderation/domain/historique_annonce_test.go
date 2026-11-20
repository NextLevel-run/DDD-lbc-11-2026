package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maintenant = time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)

func unInstantane() InstantaneAnnonce {
	return InstantaneAnnonce{
		AnnonceID: "a1", Titre: "Vélo de ville", Description: "Très bon état.", Prix: 15000,
		Categorie: "vehicules", Ville: "Lyon", CodePostal: "69003", Pseudo: "cyclo",
		DeposeeLe: maintenant.Add(-time.Hour),
	}
}

func unModerateur(t *testing.T, pseudo string) Moderateur {
	t.Helper()
	m, err := NewModerateur(pseudo)
	require.NoError(t, err)
	return m
}

func historiqueEnFile() *HistoriqueAnnonce {
	h := NewHistoriqueAnnonce("a1")
	h.MettreEnFile(unInstantane(), maintenant)
	return RestoreHistoriqueAnnonce("a1", h.NouveauxEvenements())
}

func TestNewModerateur_AcceptsTwoToThirtyCharactersOnceTrimmed(t *testing.T) {
	m, err := NewModerateur("  alice ")
	require.NoError(t, err)
	assert.Equal(t, "alice", m.Pseudo())

	_, err = NewModerateur("a")
	assert.ErrorIs(t, err, ErrModerateurInvalide)
	_, err = NewModerateur("   ")
	assert.ErrorIs(t, err, ErrModerateurInvalide)
	_, err = NewModerateur("abcdefghijklmnopqrstuvwxyzabcde")
	assert.ErrorIs(t, err, ErrModerateurInvalide)
	_, err = NewModerateur("abcdefghijklmnopqrstuvwxyzabcd")
	assert.NoError(t, err)
}

func TestMettreEnFile_AppendsTheSnapshotToAnEmptyStream(t *testing.T) {
	h := NewHistoriqueAnnonce("a1")

	h.MettreEnFile(unInstantane(), maintenant)

	assert.Equal(t, EtatEnFile, h.Etat())
	assert.Equal(t, unInstantane(), h.Annonce())
	assert.Equal(t, maintenant, h.MisEnFileLe())
	require.Len(t, h.NouveauxEvenements(), 1)
	assert.Equal(t, "AnnonceMiseEnFile", h.NouveauxEvenements()[0].EventType())
}

func TestMettreEnFile_IgnoresAnAnnonceAlreadyInTheStore(t *testing.T) {
	h := historiqueEnFile()

	h.MettreEnFile(unInstantane(), maintenant.Add(time.Minute))

	assert.Empty(t, h.NouveauxEvenements())
	assert.Equal(t, maintenant, h.MisEnFileLe())
}

func TestMettreEnFile_PutsAChallengedAnnonceBackInTheQueueWithItsNewSnapshot(t *testing.T) {
	h := RestoreHistoriqueAnnonce("a1", []Evenement{
		AnnonceMiseEnFileEvent{Annonce: unInstantane(), Date: maintenant},
		AnnonceVerrouilleeEvent{Moderateur: "alice", Date: maintenant.Add(time.Minute)},
		AnnonceChallengeeEvent{annonceID: "a1", Moderateur: "alice", motif: "Photos manquantes", Date: maintenant.Add(2 * time.Minute)},
	})
	corrige := unInstantane()
	corrige.Titre = "Vélo de ville révisé"

	h.MettreEnFile(corrige, maintenant.Add(time.Hour))

	assert.Equal(t, EtatEnFile, h.Etat())
	assert.Equal(t, "Vélo de ville révisé", h.Annonce().Titre)
	assert.Equal(t, maintenant.Add(time.Hour), h.MisEnFileLe())
	require.Len(t, h.NouveauxEvenements(), 1)
	assert.Equal(t, "AnnonceMiseEnFile", h.NouveauxEvenements()[0].EventType())
	assert.Len(t, h.Evenements(), 4)
}

func TestMettreEnFile_IgnoresAValideeOrRefuseeAnnonce(t *testing.T) {
	for name, decision := range map[string]Evenement{
		"validee": AnnonceValideeEvent{annonceID: "a1", Moderateur: "alice", Date: maintenant},
		"refusee": AnnonceRefuseeEvent{annonceID: "a1", Moderateur: "alice", Date: maintenant},
	} {
		t.Run(name, func(t *testing.T) {
			h := RestoreHistoriqueAnnonce("a1", []Evenement{
				AnnonceMiseEnFileEvent{Annonce: unInstantane(), Date: maintenant}, decision,
			})
			etat := h.Etat()

			h.MettreEnFile(unInstantane(), maintenant.Add(time.Hour))

			assert.Empty(t, h.NouveauxEvenements())
			assert.Equal(t, etat, h.Etat())
		})
	}
}

func TestRestoreHistoriqueAnnonce_ReplaysTheStreamIntoItsState(t *testing.T) {
	h := RestoreHistoriqueAnnonce("a1", []Evenement{
		AnnonceMiseEnFileEvent{Annonce: unInstantane(), Date: maintenant},
		AnnonceVerrouilleeEvent{Moderateur: "alice", Date: maintenant.Add(time.Minute)},
	})

	assert.Equal(t, EtatVerrouillee, h.Etat())
	assert.Equal(t, 2, h.Version())
	assert.Empty(t, h.NouveauxEvenements())
	moderateur, depuis, ok := h.Verrou(maintenant.Add(2 * time.Minute))
	assert.True(t, ok)
	assert.Equal(t, "alice", moderateur)
	assert.Equal(t, maintenant.Add(time.Minute), depuis)
}

func TestRestoreHistoriqueAnnonce_BringsAReleasedAnnonceBackToTheQueue(t *testing.T) {
	h := RestoreHistoriqueAnnonce("a1", []Evenement{
		AnnonceMiseEnFileEvent{Annonce: unInstantane(), Date: maintenant},
		AnnonceVerrouilleeEvent{Moderateur: "alice", Date: maintenant},
		VerrouAnnonceRelacheEvent{Moderateur: "alice", Cause: CauseManuel, Date: maintenant},
	})

	assert.Equal(t, EtatEnFile, h.Etat())
	_, _, ok := h.Verrou(maintenant)
	assert.False(t, ok)
}

func TestVerrouiller_LocksAnAnnonceOfTheQueue(t *testing.T) {
	h := historiqueEnFile()

	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant.Add(time.Minute)))

	assert.Equal(t, EtatVerrouillee, h.Etat())
	require.Len(t, h.NouveauxEvenements(), 1)
	assert.Equal(t, AnnonceVerrouilleeEvent{Moderateur: "alice", Date: maintenant.Add(time.Minute)}, h.NouveauxEvenements()[0])
}

func TestVerrouiller_IsRefusedWhenAnotherModerateurHoldsAValidLock(t *testing.T) {
	h := historiqueEnFile()
	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant))
	h = RestoreHistoriqueAnnonce("a1", h.Evenements())

	err := h.Verrouiller(unModerateur(t, "bob"), maintenant.Add(DureeVerrou))

	assert.ErrorIs(t, err, ErrAnnonceDejaVerrouillee)
	assert.Empty(t, h.NouveauxEvenements())
}

func TestVerrouiller_IsIdempotentForTheHolder(t *testing.T) {
	h := historiqueEnFile()
	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant))
	h = RestoreHistoriqueAnnonce("a1", h.Evenements())

	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant.Add(time.Hour)))

	assert.Empty(t, h.NouveauxEvenements())
}

func TestVerrouiller_ReleasesAnExpiredLockThenLocksAgain(t *testing.T) {
	h := historiqueEnFile()
	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant))
	h = RestoreHistoriqueAnnonce("a1", h.Evenements())
	plusTard := maintenant.Add(DureeVerrou + time.Second)

	require.NoError(t, h.Verrouiller(unModerateur(t, "bob"), plusTard))

	nouveaux := h.NouveauxEvenements()
	require.Len(t, nouveaux, 2)
	assert.Equal(t, VerrouAnnonceRelacheEvent{Cause: CauseExpiration, Date: plusTard}, nouveaux[0])
	assert.Equal(t, AnnonceVerrouilleeEvent{Moderateur: "bob", Date: plusTard}, nouveaux[1])
	moderateur, _, ok := h.Verrou(plusTard)
	assert.True(t, ok)
	assert.Equal(t, "bob", moderateur)
}

func TestVerrouiller_RelocksForTheSameModerateurAfterExpiration(t *testing.T) {
	h := historiqueEnFile()
	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant))
	h = RestoreHistoriqueAnnonce("a1", h.Evenements())

	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant.Add(3*time.Hour)))

	assert.Len(t, h.NouveauxEvenements(), 2)
}

func TestVerrouiller_AnswersIntrouvableForAnEmptyStream(t *testing.T) {
	err := NewHistoriqueAnnonce("a1").Verrouiller(unModerateur(t, "alice"), maintenant)

	assert.ErrorIs(t, err, ErrAnnonceIntrouvable)
}

func TestVerrou_ExpiresOnlyAfterTwoHours(t *testing.T) {
	h := historiqueEnFile()
	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant))

	_, _, ok := h.Verrou(maintenant.Add(DureeVerrou))
	assert.True(t, ok)
	_, _, ok = h.Verrou(maintenant.Add(DureeVerrou + time.Nanosecond))
	assert.False(t, ok)
}

func TestRelacher_PutsTheAnnonceBackInTheQueue(t *testing.T) {
	h := historiqueEnFile()
	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant))
	h = RestoreHistoriqueAnnonce("a1", h.Evenements())

	require.NoError(t, h.Relacher(unModerateur(t, "alice"), maintenant.Add(time.Minute)))

	assert.Equal(t, EtatEnFile, h.Etat())
	assert.Equal(t, VerrouAnnonceRelacheEvent{Moderateur: "alice", Cause: CauseManuel, Date: maintenant.Add(time.Minute)}, h.NouveauxEvenements()[0])
}

func TestRelacher_IsReservedToTheHolderOfAValidLock(t *testing.T) {
	enFile := historiqueEnFile()
	verrouillee := historiqueEnFile()
	require.NoError(t, verrouillee.Verrouiller(unModerateur(t, "alice"), maintenant))
	verrouillee = RestoreHistoriqueAnnonce("a1", verrouillee.Evenements())

	assert.ErrorIs(t, enFile.Relacher(unModerateur(t, "alice"), maintenant), ErrAnnoncePasVerrouilleeParModerateur)
	assert.ErrorIs(t, verrouillee.Relacher(unModerateur(t, "bob"), maintenant), ErrAnnoncePasVerrouilleeParModerateur)
	assert.ErrorIs(t, verrouillee.Relacher(unModerateur(t, "alice"), maintenant.Add(3*time.Hour)), ErrVerrouExpire)
	assert.ErrorIs(t, NewHistoriqueAnnonce("a1").Relacher(unModerateur(t, "alice"), maintenant), ErrAnnonceIntrouvable)
	assert.Empty(t, verrouillee.NouveauxEvenements())
}

func historiqueVerrouille(t *testing.T) *HistoriqueAnnonce {
	t.Helper()
	h := historiqueEnFile()
	require.NoError(t, h.Verrouiller(unModerateur(t, "alice"), maintenant))
	return RestoreHistoriqueAnnonce("a1", h.Evenements())
}

func unMotif(t *testing.T) Motif {
	t.Helper()
	m, err := NewMotif("Photos manquantes")
	require.NoError(t, err)
	return m
}

func TestNewMotif_AcceptsTenToThousandCharactersOnceTrimmed(t *testing.T) {
	m, err := NewMotif("  Photos manquantes ")
	require.NoError(t, err)
	assert.Equal(t, "Photos manquantes", m.Texte())

	_, err = NewMotif("trop court")
	assert.NoError(t, err) // exactly 10
	_, err = NewMotif("court")
	assert.ErrorIs(t, err, ErrMotifInvalide)
	_, err = NewMotif("   court    ")
	assert.ErrorIs(t, err, ErrMotifInvalide)
	_, err = NewMotif(strings.Repeat("é", 1000))
	assert.NoError(t, err)
	_, err = NewMotif(strings.Repeat("é", 1001))
	assert.ErrorIs(t, err, ErrMotifInvalide)
}

func TestValider_AppendsTheDecisionAndLeavesTheQueue(t *testing.T) {
	h := historiqueVerrouille(t)

	e, err := h.Valider(unModerateur(t, "alice"), maintenant.Add(time.Minute))

	require.NoError(t, err)
	assert.Equal(t, EtatValidee, h.Etat())
	assert.False(t, h.EstAModerer())
	assert.Equal(t, "AnnonceValidee", e.EventType())
	assert.Equal(t, "a1", e.AnnonceID())
	require.Len(t, h.NouveauxEvenements(), 1)
	assert.Equal(t, e, h.NouveauxEvenements()[0])
	_, _, locked := h.Verrou(maintenant)
	assert.False(t, locked)
}

func TestRefuser_AppendsTheDecisionAndLeavesTheQueue(t *testing.T) {
	h := historiqueVerrouille(t)

	e, err := h.Refuser(unModerateur(t, "alice"), maintenant)

	require.NoError(t, err)
	assert.Equal(t, EtatRefusee, h.Etat())
	assert.False(t, h.EstAModerer())
	assert.Equal(t, "AnnonceRefusee", e.EventType())
	assert.Equal(t, "a1", e.AnnonceID())
}

func TestChallenger_AppendsTheDecisionWithItsMotif(t *testing.T) {
	h := historiqueVerrouille(t)

	e, err := h.Challenger(unModerateur(t, "alice"), unMotif(t), maintenant)

	require.NoError(t, err)
	assert.Equal(t, EtatChallengee, h.Etat())
	assert.False(t, h.EstAModerer())
	assert.Equal(t, "AnnonceChallengee", e.EventType())
	assert.Equal(t, "a1", e.AnnonceID())
	assert.Equal(t, "Photos manquantes", e.Motif())
	assert.Equal(t, "alice", e.Moderateur)
}

func TestDecisions_RequireAValidLockHeldByTheModerateur(t *testing.T) {
	decisions := map[string]func(*HistoriqueAnnonce, Moderateur, time.Time) error{
		"valider": func(h *HistoriqueAnnonce, m Moderateur, now time.Time) error {
			_, err := h.Valider(m, now)
			return err
		},
		"refuser": func(h *HistoriqueAnnonce, m Moderateur, now time.Time) error {
			_, err := h.Refuser(m, now)
			return err
		},
		"challenger": func(h *HistoriqueAnnonce, m Moderateur, now time.Time) error {
			_, err := h.Challenger(m, unMotif(t), now)
			return err
		},
	}
	for nom, decider := range decisions {
		t.Run(nom, func(t *testing.T) {
			alice, bob := unModerateur(t, "alice"), unModerateur(t, "bob")

			h := historiqueEnFile()
			assert.ErrorIs(t, decider(h, alice, maintenant), ErrAnnoncePasVerrouilleeParModerateur, "not locked")
			h = historiqueVerrouille(t)
			assert.ErrorIs(t, decider(h, bob, maintenant), ErrAnnoncePasVerrouilleeParModerateur, "locked by another")
			assert.ErrorIs(t, decider(h, alice, maintenant.Add(DureeVerrou+time.Second)), ErrVerrouExpire, "expired")
			assert.NoError(t, decider(historiqueVerrouille(t), alice, maintenant.Add(DureeVerrou)), "valid at exactly 2 hours")
			assert.ErrorIs(t, decider(NewHistoriqueAnnonce("a1"), alice, maintenant), ErrAnnonceIntrouvable, "empty stream")
			assert.Empty(t, h.NouveauxEvenements(), "a refused decision appends nothing")
		})
	}
}

func TestDecidedAnnonce_CannotBeLockedNorDecidedAgain(t *testing.T) {
	h := historiqueVerrouille(t)
	alice := unModerateur(t, "alice")
	_, err := h.Valider(alice, maintenant)
	require.NoError(t, err)
	h = RestoreHistoriqueAnnonce("a1", h.Evenements())

	assert.ErrorIs(t, h.Verrouiller(alice, maintenant), ErrAnnoncePasEnFile)
	_, err = h.Refuser(alice, maintenant)
	assert.ErrorIs(t, err, ErrAnnoncePasVerrouilleeParModerateur)
	assert.Equal(t, EtatValidee, h.Etat())
}
