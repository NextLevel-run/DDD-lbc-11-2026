package contracts_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/stretchr/testify/assert"
)

// The event types travel on the bus and are read by other contexts: they are a
// contract, so a rename must break this test.
func TestEventTypes_AreStable(t *testing.T) {
	date := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		event eventbus.DomainEvent
		want  string
	}{
		{contracts.AnnonceDeposeeEvent{Date: date}, "AnnonceDeposee"},
		{contracts.AnnonceCorrigeeEvent{Date: date}, "AnnonceCorrigee"},
		{contracts.AnnonceValideeEvent{Date: date}, "AnnonceValidee"},
		{contracts.AnnonceRefuseeEvent{Date: date}, "AnnonceRefusee"},
		{contracts.AnnonceChallengeeEvent{Date: date}, "AnnonceChallengee"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			assert.Equal(t, c.want, c.event.EventType())
			assert.Equal(t, date, c.event.(interface{ OccurredAt() time.Time }).OccurredAt())
		})
	}
}

func TestEventTypeConstants_MatchTheEvents(t *testing.T) {
	assert.Equal(t, contracts.AnnonceDeposeeEvent{}.EventType(), contracts.AnnonceDeposeeEventType)
	assert.Equal(t, contracts.AnnonceCorrigeeEvent{}.EventType(), contracts.AnnonceCorrigeeEventType)
	assert.Equal(t, contracts.AnnonceValideeEvent{}.EventType(), contracts.AnnonceValideeEventType)
	assert.Equal(t, contracts.AnnonceRefuseeEvent{}.EventType(), contracts.AnnonceRefuseeEventType)
	assert.Equal(t, contracts.AnnonceChallengeeEvent{}.EventType(), contracts.AnnonceChallengeeEventType)
}
