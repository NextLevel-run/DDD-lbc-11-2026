package consumer

import (
	"log"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/application/command"
	"ddd-second-hand-marketplace/internal/contracts"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// SubscribeAnnonceValidee puts the annonce online on each AnnonceValidee, then
// tells its vendeur. A failure, of the activation or of the email, is logged,
// never retried, and never fails the decision.
func SubscribeAnnonceValidee(bus eventbus.Bus, activer command.ActiverAnnonceCommand, envoyer command.EnvoyerAnnonceValideeCommand) error {
	return bus.Subscribe(contracts.AnnonceValideeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(contracts.AnnonceValideeEvent)
		if !ok {
			return nil
		}
		annonce, err := activer(e.AnnonceID)
		if err != nil {
			log.Printf("annonce validee: annonce %s: %v", e.AnnonceID, err)
			return nil
		}
		if err := envoyer(annonce); err != nil {
			log.Printf("annonce validee: email, annonce %s: %v", e.AnnonceID, err)
		}
		return nil
	})
}

// SubscribeAnnonceChallengee sends the annonce back to its vendeur on each
// AnnonceChallengee, with the motif, then tells them. A failure is logged,
// never retried, and never fails the decision.
func SubscribeAnnonceChallengee(bus eventbus.Bus, marquer command.MarquerAnnonceChallengeeCommand, envoyer command.EnvoyerAnnonceChallengeeCommand) error {
	return bus.Subscribe(contracts.AnnonceChallengeeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(contracts.AnnonceChallengeeEvent)
		if !ok {
			return nil
		}
		annonce, err := marquer(e.AnnonceID, e.Motif)
		if err != nil {
			log.Printf("annonce challengee: annonce %s: %v", e.AnnonceID, err)
			return nil
		}
		if err := envoyer(annonce); err != nil {
			log.Printf("annonce challengee: email, annonce %s: %v", e.AnnonceID, err)
		}
		return nil
	})
}

// SubscribeAnnonceRefusee marks the annonce refusee on each AnnonceRefusee. No
// email: a refusal means fraud. A failure is logged and never retried.
func SubscribeAnnonceRefusee(bus eventbus.Bus, marquer command.MarquerAnnonceRefuseeCommand) error {
	return bus.Subscribe(contracts.AnnonceRefuseeEventType, func(event eventbus.DomainEvent) error {
		e, ok := event.(contracts.AnnonceRefuseeEvent)
		if !ok {
			return nil
		}
		if err := marquer(e.AnnonceID); err != nil {
			log.Printf("annonce refusee: annonce %s: %v", e.AnnonceID, err)
		}
		return nil
	})
}
