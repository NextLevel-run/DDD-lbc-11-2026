package consumer

import (
	"log"

	"ddd-second-hand-marketplace/internal/annonce/application/command"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// The decisions of moderation are read through local interfaces: this context
// never imports the moderation one.
type annonceValideeEvent interface{ AnnonceID() string }

type annonceRefuseeEvent interface{ AnnonceID() string }

type annonceChallengeeEvent interface {
	AnnonceID() string
	Motif() string
}

// Stable types of the moderation events.
const (
	annonceValideeEventType    = "AnnonceValidee"
	annonceRefuseeEventType    = "AnnonceRefusee"
	annonceChallengeeEventType = "AnnonceChallengee"
)

// SubscribeAnnonceValidee puts the annonce online on each AnnonceValidee, then
// tells its vendeur. A failure, of the activation or of the email, is logged,
// never retried, and never fails the decision.
func SubscribeAnnonceValidee(bus eventbus.Bus, activer command.ActiverAnnonceCommand, envoyer command.EnvoyerAnnonceValideeCommand) error {
	return bus.Subscribe(annonceValideeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(annonceValideeEvent)
		if !ok {
			return nil
		}
		annonce, err := activer(e.AnnonceID())
		if err != nil {
			log.Printf("annonce validee: annonce %s: %v", e.AnnonceID(), err)
			return nil
		}
		if err := envoyer(annonce); err != nil {
			log.Printf("annonce validee: email, annonce %s: %v", e.AnnonceID(), err)
		}
		return nil
	})
}

// SubscribeAnnonceChallengee sends the annonce back to its vendeur on each
// AnnonceChallengee, with the motif, then tells them. A failure is logged,
// never retried, and never fails the decision.
func SubscribeAnnonceChallengee(bus eventbus.Bus, marquer command.MarquerAnnonceChallengeeCommand, envoyer command.EnvoyerAnnonceChallengeeCommand) error {
	return bus.Subscribe(annonceChallengeeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(annonceChallengeeEvent)
		if !ok {
			return nil
		}
		annonce, err := marquer(e.AnnonceID(), e.Motif())
		if err != nil {
			log.Printf("annonce challengee: annonce %s: %v", e.AnnonceID(), err)
			return nil
		}
		if err := envoyer(annonce); err != nil {
			log.Printf("annonce challengee: email, annonce %s: %v", e.AnnonceID(), err)
		}
		return nil
	})
}

// SubscribeAnnonceRefusee marks the annonce refusee on each AnnonceRefusee. No
// email: a refusal means fraud. A failure is logged and never retried.
func SubscribeAnnonceRefusee(bus eventbus.Bus, marquer command.MarquerAnnonceRefuseeCommand) error {
	return bus.Subscribe(annonceRefuseeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(annonceRefuseeEvent)
		if !ok {
			return nil
		}
		if err := marquer(e.AnnonceID()); err != nil {
			log.Printf("annonce refusee: annonce %s: %v", e.AnnonceID(), err)
		}
		return nil
	})
}
