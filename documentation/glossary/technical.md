# Technical glossary

Architecture and tooling vocabulary, in English. One line per term: what it **is** in this
repository, not its textbook definition nor what the code shows. Format: `- **Term** — definition.`

## Design

- **Bounded context** — one model and its language; in the API, one folder under `internal/bounded-context/`, one per subdomain.
- **Module** — a folder with its own `AGENTS.md`: an app or a bounded context.

## API: hexagonal architecture

- **Hexagonal architecture** — dependencies point inward: adapter → application → domain.
- **Domain** — the pure business code of a bounded context; it imports only the standard library.
  - **Aggregate** — a unit of change guarded by its root, built valid by a `New…` constructor.
  - **Value object** — an immutable value with no identity, valid from construction.
  - **Domain event** — a past-tense fact a command publishes once its change is persisted.
  - **Domain error** — an `Err…` value for a business failure; the HTTP adapter maps it to a status.
- **Port** — an interface the domain declares for what it needs from outside.
  - **Repository** — the port that stores and loads one aggregate.
- **Application** — the use cases of a bounded context, one per file.
  - **Command** — a use case that changes state: load, call the aggregate, persist, publish.
  - **Query** — a use case without side effect that returns a view.
  - **View** — a read-only structure shaped for its reader; never an aggregate.
- **Adapter** — code connecting a bounded context to the outside.
  - **Driving adapter** — calls the application: an HTTP handler, a consumer.
  - **Driven adapter** — implements a port: a repository, a mailer.
  - **In-memory adapter** — a driven adapter holding its data in memory; the API and the tests run on them.
  - **Consumer** — a driving adapter that runs a use case when an event type is published.
- **Composition root** — where adapters are built and injected: `newServer` in `cmd/api/server.go` (API), `gateways.ts` (frontend).

## API: event sourcing

- **Event sourcing** — an aggregate's state is whatever replaying its stream gives; only events are stored.
  - **Stream** — the ordered events of one aggregate; its length is the version.
  - **Event store** — the repository of a stream: loads it, appends events under a version check.
  - **Optimistic version check** — an append fails with `ErrConflitVersion` if the stream grew since it was loaded.

## API: shared infrastructure

- **Event bus** — delivers domain events to their subscribed consumers, synchronously or asynchronously.
- **Event collector** — a test helper that records the events a bus received.
- **Contracts** — `internal/contracts`: the events that cross a context boundary, as plain data; the only code two contexts share.
- **Mailer** — the port for sending an email: SMTP (Mailcatcher in dev), fake (logs) or spy (tests).

## Frontend

- **Feature** — a folder under `features/` with everything one business capability needs.
- **Gateway** — a feature's port to the API; pages depend on it, never on `fetch`.
  - **HTTP gateway** — the gateway adapter that calls the API.
  - **In-memory gateway** — the gateway adapter every frontend test runs on.
- **Route** — a file under `routes/` mapping a URL to a page; the route tree is generated from it.
- **Route context** — what the router hands every route: the query client and the gateways.
- **Query options** — a factory taking a gateway and returning one read's key and fetch function.

## Tooling

- **Docs budget** — the size limits on `AGENTS.md` files and glossaries, checked on edit, by `make check` and in CI.
- **Spec** — one feature described from the user's side, under `documentation/specs/`.
- **Ticket** — one vertical slice of a spec and its blockers, under `documentation/tickets/`.
