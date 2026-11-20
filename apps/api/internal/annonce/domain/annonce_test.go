package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validArgs() NewAnnonceArgs {
	return NewAnnonceArgs{
		Titre:       "Vélo de course",
		Description: "Cadre carbone, très bon état.",
		Prix:        45000,
		Categorie:   "loisirs",
		Ville:       "Lyon",
		CodePostal:  "69003",
		Pseudo:      "cycliste69",
		MotDePasse:  "demo1234",
		Email:       "cycliste@example.com",
		Telephone:   "0612345678",
		DeposeeLe:   time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC),
	}
}

func TestNewAnnonce_CreatesAnAnnonceEnAttenteModeration(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())

	require.NoError(t, err)
	_, parseErr := uuid.Parse(annonce.ID())
	assert.NoError(t, parseErr)
	assert.Equal(t, EnAttenteModeration, annonce.Statut())
	assert.False(t, annonce.EstActivee())
	assert.Equal(t, "Vélo de course", annonce.Titre())
	assert.Equal(t, "Cadre carbone, très bon état.", annonce.Description())
	assert.Equal(t, int64(45000), annonce.Prix().Centimes())
	assert.Equal(t, CategorieLoisirs, annonce.Categorie())
	assert.Equal(t, "69", annonce.Localisation().Departement().Code())
	assert.Equal(t, "cycliste69", annonce.Vendeur().Pseudo())
	assert.Equal(t, "cycliste@example.com", annonce.Vendeur().Email())
	assert.Equal(t, "0612345678", annonce.Vendeur().Telephone())
	assert.Empty(t, annonce.Images())
	assert.Equal(t, validArgs().DeposeeLe, annonce.DeposeeLe())
}

func TestNewAnnonce_GivesEachAnnonceItsOwnID(t *testing.T) {
	first, _ := NewAnnonce(validArgs())
	second, _ := NewAnnonce(validArgs())

	assert.NotEqual(t, first.ID(), second.ID())
}

func TestNewAnnonce_AcceptsAVendeurWithoutTelephone(t *testing.T) {
	args := validArgs()
	args.Telephone = ""

	annonce, err := NewAnnonce(args)

	require.NoError(t, err)
	assert.Empty(t, annonce.Vendeur().Telephone())
}

func TestNewAnnonce_RejectsInvalidValueObjects(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*NewAnnonceArgs)
		wantErr error
	}{
		{"titre too short", func(a *NewAnnonceArgs) { a.Titre = "ab" }, ErrTitreInvalide},
		{"titre blank", func(a *NewAnnonceArgs) { a.Titre = "     " }, ErrTitreInvalide},
		{"titre too long", func(a *NewAnnonceArgs) { a.Titre = strings.Repeat("a", 101) }, ErrTitreInvalide},
		{"description too short", func(a *NewAnnonceArgs) { a.Description = "123456789" }, ErrDescriptionInvalide},
		{"description too long", func(a *NewAnnonceArgs) { a.Description = strings.Repeat("a", 4001) }, ErrDescriptionInvalide},
		{"negative prix", func(a *NewAnnonceArgs) { a.Prix = -100 }, ErrPrixNegatif},
		{"unknown categorie", func(a *NewAnnonceArgs) { a.Categorie = "bijoux" }, ErrCategorieInconnue},
		{"malformed code postal", func(a *NewAnnonceArgs) { a.CodePostal = "690" }, ErrCodePostalInvalide},
		{"unknown departement", func(a *NewAnnonceArgs) { a.CodePostal = "96000" }, ErrDepartementInconnu},
		{"invalid pseudo", func(a *NewAnnonceArgs) { a.Pseudo = "a" }, ErrPseudoInvalide},
		{"invalid email", func(a *NewAnnonceArgs) { a.Email = "pas-un-email" }, ErrEmailInvalide},
		{"invalid telephone", func(a *NewAnnonceArgs) { a.Telephone = "12345" }, ErrTelephoneInvalide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := validArgs()
			tt.change(&args)

			_, err := NewAnnonce(args)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewAnnonce_AcceptsTitreAndDescriptionLengthBounds(t *testing.T) {
	tests := []struct {
		name        string
		titre       string
		description string
	}{
		{"shortest", "abc", "1234567890"},
		{"longest, counted in characters", strings.Repeat("é", 100), strings.Repeat("é", 4000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := validArgs()
			args.Titre = tt.titre
			args.Description = tt.description

			_, err := NewAnnonce(args)

			require.NoError(t, err)
		})
	}
}

func TestNewAnnonce_TrimsTitreAndDescription(t *testing.T) {
	args := validArgs()
	args.Titre = "  Vélo de course  "
	args.Description = "\n Cadre carbone, très bon état. \n"

	annonce, err := NewAnnonce(args)

	require.NoError(t, err)
	assert.Equal(t, "Vélo de course", annonce.Titre())
	assert.Equal(t, "Cadre carbone, très bon état.", annonce.Description())
}

func TestNewAnnonce_AcceptsADon(t *testing.T) {
	args := validArgs()
	args.Prix = 0

	annonce, err := NewAnnonce(args)

	require.NoError(t, err)
	assert.True(t, annonce.Prix().EstUnDon())
}

func TestRestoreAnnonce_KeepsTheStoredIDAndStatut(t *testing.T) {
	id := uuid.NewString()

	annonce, err := RestoreAnnonce(id, Statut(""), validArgs())

	require.NoError(t, err)
	assert.Equal(t, id, annonce.ID())
	assert.False(t, annonce.EstActivee())
}

func TestParseAnnonceID_RejectsAMalformedID(t *testing.T) {
	_, err := ParseAnnonceID("not-a-uuid")

	require.ErrorIs(t, err, ErrIDAnnonceInvalide)
}

func TestParseAnnonceID_AcceptsAUUID(t *testing.T) {
	id := uuid.NewString()

	parsed, err := ParseAnnonceID(id)

	require.NoError(t, err)
	assert.Equal(t, id, parsed)
}

func TestNewAnnonce_RefusesAMissingOrTooShortMotDePasse(t *testing.T) {
	for _, motDePasse := range []string{"", "1234567"} {
		args := validArgs()
		args.MotDePasse = motDePasse

		_, err := NewAnnonce(args)

		assert.ErrorIs(t, err, ErrMotDePasseTropCourt)
	}
}

func TestNewAnnonce_KeepsAMotDePasseThatCorrespondsOnlyToTheSaisie(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())

	require.NoError(t, err)
	assert.True(t, annonce.MotDePasseCorrespond("demo1234"))
	assert.False(t, annonce.MotDePasseCorrespond("demo12345"))
}

func TestSupprimer_WithdrawsTheAnnonceWithItsMotDePasse(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())

	require.NoError(t, annonce.Supprimer("demo1234"))

	assert.Equal(t, Supprimee, annonce.Statut())
	assert.False(t, annonce.EstActivee())
}

func TestSupprimer_RefusesAWrongMotDePasseAndKeepsTheAnnonceEnLigne(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())

	err = annonce.Supprimer("mauvais-mot")

	assert.ErrorIs(t, err, ErrMotDePasseIncorrect)
	assert.True(t, annonce.EstActivee())
}

func TestSupprimer_AnswersIntrouvableWhenAlreadySupprimee(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())
	require.NoError(t, err)
	require.NoError(t, annonce.Activer())
	require.NoError(t, annonce.Supprimer("demo1234"))

	assert.ErrorIs(t, annonce.Supprimer("demo1234"), ErrAnnonceIntrouvable)
	assert.ErrorIs(t, annonce.Supprimer("mauvais-mot"), ErrAnnonceIntrouvable)
}

func TestActiver_MakesAnAnnonceEnAttenteModerationActivee(t *testing.T) {
	annonce, _ := NewAnnonce(validArgs())

	require.NoError(t, annonce.Activer())

	assert.Equal(t, Activee, annonce.Statut())
	assert.True(t, annonce.EstActivee())
}

func TestChallenger_KeepsTheMotifAndIsNotVisible(t *testing.T) {
	annonce, _ := NewAnnonce(validArgs())

	require.NoError(t, annonce.Challenger("Photos manquantes"))

	assert.Equal(t, Challengee, annonce.Statut())
	assert.Equal(t, "Photos manquantes", annonce.Motif())
	assert.False(t, annonce.EstActivee())
}

func TestRefuser_MakesTheAnnonceRefusee(t *testing.T) {
	annonce, _ := NewAnnonce(validArgs())

	require.NoError(t, annonce.Refuser())

	assert.Equal(t, Refusee, annonce.Statut())
	assert.False(t, annonce.EstActivee())
}

func TestModerationTransitions_AreRefusedUnlessEnAttenteModeration(t *testing.T) {
	for _, statut := range []Statut{Activee, Challengee, Refusee, Supprimee} {
		annonce, err := RestoreAnnonce(uuid.NewString(), statut, validArgs())
		require.NoError(t, err)

		assert.ErrorIs(t, annonce.Activer(), ErrTransitionStatutInvalide, statut)
		assert.ErrorIs(t, annonce.Challenger("Motif suffisant"), ErrTransitionStatutInvalide, statut)
		assert.ErrorIs(t, annonce.Refuser(), ErrTransitionStatutInvalide, statut)
		assert.Equal(t, statut, annonce.Statut())
	}
}

func TestSupprimer_AnswersIntrouvableWhenNotActivee(t *testing.T) {
	for _, statut := range []Statut{EnAttenteModeration, Challengee, Refusee} {
		annonce, err := RestoreAnnonce(uuid.NewString(), statut, validArgs())
		require.NoError(t, err)

		assert.ErrorIs(t, annonce.Supprimer("demo1234"), ErrAnnonceIntrouvable, statut)
		assert.Equal(t, statut, annonce.Statut())
	}
}

func uneAnnonceChallengee(t *testing.T) *Annonce {
	t.Helper()
	annonce, err := NewAnnonce(validArgs())
	require.NoError(t, err)
	require.NoError(t, annonce.Challenger("Photos manquantes"))
	return annonce
}

func uneCorrection() CorrectionArgs {
	return CorrectionArgs{
		Titre: "Vélo de course révisé", Description: "Cadre carbone, photos ajoutées.", Prix: 42000,
		Categorie: "vehicules", Ville: "Paris", CodePostal: "75011",
	}
}

func TestCorriger_ReplacesTheFieldsClearsTheMotifAndSendsItBackToModeration(t *testing.T) {
	annonce := uneAnnonceChallengee(t)
	id := annonce.ID()

	require.NoError(t, annonce.Corriger("demo1234", uneCorrection()))

	assert.Equal(t, EnAttenteModeration, annonce.Statut())
	assert.Empty(t, annonce.Motif())
	assert.Equal(t, "Vélo de course révisé", annonce.Titre())
	assert.Equal(t, "Cadre carbone, photos ajoutées.", annonce.Description())
	assert.Equal(t, int64(42000), annonce.Prix().Centimes())
	assert.Equal(t, "vehicules", annonce.Categorie().Code())
	assert.Equal(t, "Paris", annonce.Localisation().Ville())
	assert.Equal(t, "75", annonce.Localisation().Departement().Code())
	assert.Equal(t, id, annonce.ID())
}

func TestCorriger_KeepsTheVendeurTheMotDePasseAndTheDateDeDepot(t *testing.T) {
	annonce := uneAnnonceChallengee(t)

	require.NoError(t, annonce.Corriger("demo1234", uneCorrection()))

	assert.Equal(t, "cycliste69", annonce.Vendeur().Pseudo())
	assert.Equal(t, "cycliste@example.com", annonce.Vendeur().Email())
	assert.Equal(t, "0612345678", annonce.Vendeur().Telephone())
	assert.True(t, annonce.MotDePasseCorrespond("demo1234"))
	assert.Equal(t, validArgs().DeposeeLe, annonce.DeposeeLe())
}

func TestCorriger_RefusesAWrongMotDePasseAndChangesNothing(t *testing.T) {
	annonce := uneAnnonceChallengee(t)

	err := annonce.Corriger("mauvais-mot", uneCorrection())

	assert.ErrorIs(t, err, ErrMotDePasseIncorrect)
	assert.Equal(t, Challengee, annonce.Statut())
	assert.Equal(t, "Photos manquantes", annonce.Motif())
	assert.Equal(t, "Vélo de course", annonce.Titre())
}

func TestCorriger_AnswersIntrouvableUnlessTheAnnonceIsChallengee(t *testing.T) {
	for name, prepare := range map[string]func(*Annonce) error{
		"en attente": func(*Annonce) error { return nil },
		"activee":    (*Annonce).Activer,
		"refusee":    (*Annonce).Refuser,
	} {
		t.Run(name, func(t *testing.T) {
			annonce, _ := NewAnnonce(validArgs())
			require.NoError(t, prepare(annonce))
			statut := annonce.Statut()

			assert.ErrorIs(t, annonce.Corriger("demo1234", uneCorrection()), ErrAnnonceIntrouvable)
			assert.ErrorIs(t, annonce.Corriger("mauvais-mot", uneCorrection()), ErrAnnonceIntrouvable)
			assert.Equal(t, statut, annonce.Statut())
			assert.Equal(t, "Vélo de course", annonce.Titre())
		})
	}
}

func TestCorriger_AppliesTheValidationsOfTheDepotAndChangesNothingOnFailure(t *testing.T) {
	cases := map[string]struct {
		mutate func(*CorrectionArgs)
		want   error
	}{
		"titre":       {func(c *CorrectionArgs) { c.Titre = "ab" }, ErrTitreInvalide},
		"description": {func(c *CorrectionArgs) { c.Description = "court" }, ErrDescriptionInvalide},
		"prix":        {func(c *CorrectionArgs) { c.Prix = -1 }, ErrPrixNegatif},
		"categorie":   {func(c *CorrectionArgs) { c.Categorie = "inconnue" }, ErrCategorieInconnue},
		"ville":       {func(c *CorrectionArgs) { c.Ville = " " }, ErrVilleVide},
		"code postal": {func(c *CorrectionArgs) { c.CodePostal = "75" }, ErrCodePostalInvalide},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			annonce := uneAnnonceChallengee(t)
			correction := uneCorrection()
			tc.mutate(&correction)

			assert.ErrorIs(t, annonce.Corriger("demo1234", correction), tc.want)

			assert.Equal(t, Challengee, annonce.Statut())
			assert.Equal(t, "Photos manquantes", annonce.Motif())
			assert.Equal(t, "Vélo de course", annonce.Titre())
			assert.Equal(t, int64(45000), annonce.Prix().Centimes())
		})
	}
}

func TestAutoriserCorrection_ChecksTheStatutThenTheMotDePasse(t *testing.T) {
	annonce := uneAnnonceChallengee(t)

	assert.NoError(t, annonce.AutoriserCorrection("demo1234"))
	assert.ErrorIs(t, annonce.AutoriserCorrection("mauvais-mot"), ErrMotDePasseIncorrect)

	autre, _ := NewAnnonce(validArgs())
	assert.ErrorIs(t, autre.AutoriserCorrection("demo1234"), ErrAnnonceIntrouvable)
}
