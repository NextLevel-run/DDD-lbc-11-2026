# apps/api

The Go API of the marketplace: one bounded context per folder under `internal/`, each a hexagon, wired together in `cmd/api/server.go`. No database: every driven adapter is in memory.

## Language

Shared terms: [`domain.md`](../../documentation/glossary/domain.md), [`technical.md`](../../documentation/glossary/technical.md).

## Map

- `cmd/api/main.go` — reads the environment (`PORT`, default 8080) and starts the server.
- `cmd/api/server.go` — the composition root, `newServer(serverConfig{Mailer, DemoData, FrontendURL})`: builds the adapters, the use cases, registers routes and consumers; returns the handler and the demo IDs. `server_test.go` tests it over `httptest` with the mailer spy. The two contexts, `annonce` and `moderation` (event sourced), share one async event bus.
- `internal/<bounded-context>/` — `domain/`, `application/{command,query}/`, `adapter/{driven,driving}/`, and its own `AGENTS.md`.
- `pkg/eventbus/`, `pkg/mailer/` — shared infrastructure, each with its test helper under `testing/`.
- Demo data: `cmd/api/demo_annonces.go` seeds a dozen annonces at startup and logs a link to one; `DEMO_DATA=off` skips them. Only `server_test.go` loads them (the loader, and the demo annonces the QA guides rely on); other tests build their own data.
- Mailer: `newMailer` in `main.go` picks the SMTP mailer when `SMTP_ADDR` is set (Mailcatcher, `make mail`, UI on :1080), the fake mailer otherwise; `FRONTEND_URL` (default `http://localhost:3000`) is the base of the links in emails. The SMTP mailer takes bare addresses only (no display name). Tests use the mailer spy, never SMTP.

Patterns, naming and the step-by-step workflow are in the `tactical-ddd` skill: **invoke it before implementing or refactoring**. A new bounded context starts with the `scaffold-bc` skill.

## Commands

From the repository root: `make test-api`, `make lint-api`, `make dev-api`. Inside `apps/api`:

```bash
go test ./...                              # all tests
go test -run TestName ./internal/<bc>/...  # one test
go build -o bin/api ./cmd/api              # the binary; ./bin/api listens on :8080
```

## Rules

- **Read the bounded context's `AGENTS.md` before changing it.**
- **One use case per file**: commands in `application/command/`, queries in `application/query/`.
- **Validate in constructors**: an aggregate or a value object that exists is valid.
- **A command publishes its events after it has persisted**, never before.
- **Test with the in-memory adapters** and `pkg/eventbus/testing.EventCollector`; a test asserts on the returned value, the stored state and the events.
- **HTTP errors are JSON with French keys**: `{"erreur": "…"}` with a French message; routes use the method patterns of `http.ServeMux` (`GET /annonces/{id}`).
- **A use case is wired in `server.go` or it does not run**: a new command, query or consumer is not done until it is built and registered there, and covered by `server_test.go` if it adds a route or a consumer.
