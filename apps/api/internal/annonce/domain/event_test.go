package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnnonceDeposeeEvent_CarriesTheAnnonce(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())
	require.NoError(t, err)
	occurredAt := time.Date(2026, 11, 2, 10, 0, 1, 0, time.UTC)

	event := NewAnnonceDeposeeEvent(occurredAt, annonce)

	assert.Equal(t, "AnnonceDeposee", event.EventType())
	assert.NotEmpty(t, event.ID())
	assert.Equal(t, occurredAt, event.OccurredAt())
	assert.Same(t, annonce, event.Annonce)
}

func TestNewAnnonceSupprimeeEvent_CarriesIDRaisonAndDate(t *testing.T) {
	raison, err := NewRaisonSuppression("vendue-ailleurs")
	require.NoError(t, err)
	occurredAt := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)

	event := NewAnnonceSupprimeeEvent(occurredAt, "abc", raison)

	assert.Equal(t, "AnnonceSupprimee", event.EventType())
	assert.NotEmpty(t, event.ID())
	assert.Equal(t, "abc", event.AnnonceID)
	assert.Equal(t, raison, event.Raison)
	assert.Equal(t, occurredAt, event.SupprimeeLe)
	assert.Equal(t, occurredAt, event.OccurredAt())
}

func TestAnnonceDeposeeEvent_ExposesPrimitiveAccessors(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())
	require.NoError(t, err)

	event := NewAnnonceDeposeeEvent(time.Now(), annonce)

	assert.Equal(t, annonce.ID(), event.AnnonceID())
	assert.Equal(t, "Vélo de course", event.Titre())
	assert.Equal(t, "Cadre carbone, très bon état.", event.Description())
	assert.Equal(t, int64(45000), event.Prix())
	assert.Equal(t, "loisirs", event.Categorie())
	assert.Equal(t, "Lyon", event.Ville())
	assert.Equal(t, "69003", event.CodePostal())
	assert.Equal(t, "cycliste69", event.Pseudo())
	assert.Equal(t, validArgs().DeposeeLe, event.DeposeeLe())
}

func TestNewAnnonceCorrigeeEvent_ExposesTheSameAccessorsAsTheDepot(t *testing.T) {
	annonce, err := NewAnnonce(validArgs())
	require.NoError(t, err)
	require.NoError(t, annonce.Challenger("Photos manquantes"))
	require.NoError(t, annonce.Corriger("demo1234", uneCorrection()))
	occurredAt := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)

	event := NewAnnonceCorrigeeEvent(occurredAt, annonce)

	assert.Equal(t, "AnnonceCorrigee", event.EventType())
	assert.NotEmpty(t, event.ID())
	assert.Equal(t, occurredAt, event.OccurredAt())
	assert.Equal(t, annonce.ID(), event.AnnonceID())
	assert.Equal(t, "Vélo de course révisé", event.Titre())
	assert.Equal(t, "Cadre carbone, photos ajoutées.", event.Description())
	assert.Equal(t, int64(42000), event.Prix())
	assert.Equal(t, "vehicules", event.Categorie())
	assert.Equal(t, "Paris", event.Ville())
	assert.Equal(t, "75011", event.CodePostal())
	assert.Equal(t, "cycliste69", event.Pseudo())
	assert.Equal(t, validArgs().DeposeeLe, event.DeposeeLe())
}
