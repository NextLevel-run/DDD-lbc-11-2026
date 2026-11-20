package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	annonceinmemory "ddd-second-hand-marketplace/internal/annonce/adapter/driven/inmemory"
	annonceconsumer "ddd-second-hand-marketplace/internal/annonce/adapter/driving/consumer"
	annoncehttp "ddd-second-hand-marketplace/internal/annonce/adapter/driving/http"
	annoncecommand "ddd-second-hand-marketplace/internal/annonce/application/command"
	annoncequery "ddd-second-hand-marketplace/internal/annonce/application/query"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// serverConfig is what the composition root takes from the environment.
type serverConfig struct {
	Mailer      mailer.Mailer
	DemoData    bool
	FrontendURL string // base of the links in emails
}

// fromEmail is the fixed sender of every email of the platform.
const fromEmail = "noreply@seconde-main.local"

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
		cfg.Mailer, fromEmail, cfg.FrontendURL,
	)); err != nil {
		return nil, nil, fmt.Errorf("consumer: %w", err)
	}
	annoncehttp.NewHandler(
		annoncequery.BuildConsulterAnnonceQuery(annonceRepository),
		annoncequery.BuildRechercherAnnoncesQuery(annonceRepository),
		annoncequery.BuildListerRegionsQuery(),
		annoncecommand.BuildDeposerAnnonceCommand(annonceRepository, eventBus, time.Now),
		annoncecommand.BuildSupprimerAnnonceCommand(annonceRepository, eventBus, time.Now),
		annoncecommand.BuildContacterCommand(annonceRepository, cfg.Mailer, eventBus, time.Now, fromEmail, cfg.FrontendURL),
	).RegisterRoutes(mux)

	return mux, demoIDs, nil
}
