package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	annonceinmemory "ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driven/inmemory"
	annonceconsumer "ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driving/consumer"
	annoncehttp "ddd-second-hand-marketplace/internal/bounded-context/annonce/adapter/driving/http"
	annoncecommand "ddd-second-hand-marketplace/internal/bounded-context/annonce/application/command"
	annoncequery "ddd-second-hand-marketplace/internal/bounded-context/annonce/application/query"
	moderationinmemory "ddd-second-hand-marketplace/internal/bounded-context/moderation/adapter/driven/inmemory"
	moderationconsumer "ddd-second-hand-marketplace/internal/bounded-context/moderation/adapter/driving/consumer"
	moderationhttp "ddd-second-hand-marketplace/internal/bounded-context/moderation/adapter/driving/http"
	moderationcommand "ddd-second-hand-marketplace/internal/bounded-context/moderation/application/command"
	moderationquery "ddd-second-hand-marketplace/internal/bounded-context/moderation/application/query"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// serverConfig is what the composition root takes from the environment.
type serverConfig struct {
	Mailer      mailer.Mailer
	DemoData    bool
	FrontendURL string // base of the links in emails
}

// newServer builds every adapter and use case, wires the consumers and returns
// the HTTP handler, with the IDs of the demo annonces it loaded. The
// integration tests run it as is, with a mailer spy.
func newServer(cfg serverConfig) (http.Handler, []string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Shared infrastructure.
	eventBus := eventbus.NewAsyncInMemoryEventBus()

	// Bounded context: annonce.
	annonceRepository := annonceinmemory.NewInMemoryAnnonceRepository()
	var demoIDs []string
	if cfg.DemoData {
		ids, err := loadDemoAnnonces(annonceRepository)
		if err != nil {
			return nil, nil, fmt.Errorf("demo data: %w", err)
		}
		demoIDs = ids
	}
	if err := annonceconsumer.SubscribeConfirmationDepot(eventBus, annoncecommand.BuildEnvoyerConfirmationDepotCommand(
		annonceRepository, cfg.Mailer, "noreply@seconde-main.local", cfg.FrontendURL,
	)); err != nil {
		return nil, nil, fmt.Errorf("consumer: %w", err)
	}
	if err := annonceconsumer.SubscribeAnnonceValidee(eventBus,
		annoncecommand.BuildActiverAnnonceCommand(annonceRepository),
		annoncecommand.BuildEnvoyerAnnonceValideeCommand(cfg.Mailer, "noreply@seconde-main.local", cfg.FrontendURL),
	); err != nil {
		return nil, nil, fmt.Errorf("consumer: %w", err)
	}
	if err := annonceconsumer.SubscribeAnnonceChallengee(eventBus,
		annoncecommand.BuildMarquerAnnonceChallengeeCommand(annonceRepository),
		annoncecommand.BuildEnvoyerAnnonceChallengeeCommand(cfg.Mailer, "noreply@seconde-main.local", cfg.FrontendURL),
	); err != nil {
		return nil, nil, fmt.Errorf("consumer: %w", err)
	}
	if err := annonceconsumer.SubscribeAnnonceRefusee(eventBus,
		annoncecommand.BuildMarquerAnnonceRefuseeCommand(annonceRepository),
	); err != nil {
		return nil, nil, fmt.Errorf("consumer: %w", err)
	}
	annoncehttp.NewHandler(
		annoncequery.BuildConsulterAnnonceQuery(annonceRepository),
		annoncequery.BuildRechercherAnnoncesQuery(annonceRepository),
		annoncequery.BuildListerRegionsQuery(),
		annoncecommand.BuildDeposerAnnonceCommand(annonceRepository, eventBus, time.Now),
		annoncecommand.BuildSupprimerAnnonceCommand(annonceRepository, eventBus, time.Now),
		annoncequery.BuildConsulterAnnonceAEditerQuery(annonceRepository),
		annoncecommand.BuildCorrigerAnnonceCommand(annonceRepository, eventBus, time.Now),
	).RegisterRoutes(mux)

	// Bounded context: moderation. It reads AnnonceDeposee and AnnonceCorrigee
	// from internal/contracts and never imports annonce; annonce reads its
	// decisions the same way.
	historiqueRepository := moderationinmemory.NewInMemoryHistoriqueAnnonceRepository()
	if err := moderationconsumer.SubscribeMiseEnFile(eventBus, moderationcommand.BuildMettreEnFileCommand(
		historiqueRepository, time.Now,
	)); err != nil {
		return nil, nil, fmt.Errorf("consumer: %w", err)
	}
	moderationhttp.NewHandler(
		moderationquery.BuildListerAnnoncesAModererQuery(historiqueRepository, time.Now),
		moderationquery.BuildConsulterAnnonceAModererQuery(historiqueRepository, time.Now),
		moderationcommand.BuildVerrouillerAnnonceCommand(historiqueRepository, time.Now),
		moderationcommand.BuildRelacherVerrouCommand(historiqueRepository, time.Now),
		moderationcommand.BuildValiderAnnonceCommand(historiqueRepository, eventBus, time.Now),
		moderationcommand.BuildRefuserAnnonceCommand(historiqueRepository, eventBus, time.Now),
		moderationcommand.BuildChallengerAnnonceCommand(historiqueRepository, eventBus, time.Now),
	).RegisterRoutes(mux)

	return mux, demoIDs, nil
}
