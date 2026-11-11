# Technical glossary

Architecture and tooling vocabulary of the marketplace. One unambiguous English name per concept,
spelled as in the code. Business terms live in `domain.md`.

An entry says what a term is, in one or two sentences, 300 characters at most. A term earns an
entry when it names an architecture or tooling concept that several modules share. A single-module
internal, a setting, a variant of another entry or a rule detail gets none: it belongs to its
module's `AGENTS.md` or to the code. An entry stands alone: no reference to a spec or a file.

## Design

- **Ubiquitous language** — the vocabulary the business and the code share inside one bounded
  context, kept in the domain glossary.
- **Subdomain** — a part of the business problem, classed core, supporting or generic.
- **Bounded context** — the boundary inside which one model and its language hold: in the API, one
  folder under `internal/`. Here each subdomain maps to one bounded context.
- **Module** — a folder of code with strong cohesion that has its own `AGENTS.md`: an app, a
  bounded context.

## API: hexagonal architecture

- **Hexagonal architecture** — the domain sits at the center and reaches the outside only through
  ports; adapters plug into them. Dependencies point inward: adapter → application → domain.
- **Domain** — the pure business code of a bounded context: aggregates, value objects, domain
  events, domain errors and ports. It imports nothing but the standard library.
  - **Aggregate** — a cluster of objects changed as one unit through its root, which guards its
    invariants. Built by a `New…` constructor that validates.
  - **Entity** — an object with an identity that lives inside an aggregate and is reached through
    its root only.
  - **Value object** — an immutable object defined by its value, with no identity, valid from
    construction (a price, an email).
  - **Domain event** — a fact that happened in the domain, named in the past tense, emitted by a
    command once the change is persisted.
  - **Domain error** — a business failure the domain declares as an `Err…` value; the HTTP adapter
    turns it into a status code.
- **Port** — an interface the domain declares for what it needs from the outside.
  - **Repository** — the port that stores and loads one aggregate.
- **Application** — the use cases of a bounded context, one per file.
  - **Use case** — one entry point into a bounded context: a command or a query.
  - **Command** — a use case that changes state: load, call the aggregate, persist, publish events.
  - **Query** — a use case that reads and has no side effect; it returns a view.
  - **View** — a read-only structure shaped for its reader, returned by a query. Never an aggregate.
- **Adapter** — code that connects a bounded context to the outside.
  - **Driving adapter** — calls the application: an HTTP handler, an event consumer.
  - **Driven adapter** — is called by the application through a port: a repository implementation,
    a mailer.
  - **In-memory adapter** — a driven adapter that keeps its data in memory. The API runs on them
    and so do the tests; there is no database.
  - **Consumer** — a driving adapter subscribed to an event type, which runs a use case in reaction.
- **Composition root** — the one place where adapters are built and injected into use cases:
  `main.go` in the API, `gateways.ts` in the frontend.

## API: shared infrastructure

- **Event bus** — publishes domain events to the consumers subscribed to their type. Its in-memory
  implementation delivers either synchronously or asynchronously.
- **Event collector** — a test helper subscribed to an event bus that records the events it
  receives, so a test can assert on what a command emitted.
- **Mailer** — the port for sending an email. The fake mailer logs instead of sending; the mailer
  spy records the emails for a test to assert on.

## Frontend

- **Feature** — a folder under `features/` holding everything the UI needs for one business
  capability: its port, its adapters, its queries and its form schemas.
- **Gateway** — the frontend's port: the interface a feature declares for what it needs from the
  API. Pages depend on it, never on `fetch`.
  - **HTTP gateway** — the adapter that implements a gateway by calling the API.
  - **In-memory gateway** — the adapter that implements a gateway in memory; every frontend test
    runs on these.
- **Route** — a file under `routes/` that maps a URL to a page; the route tree is generated from
  the file names.
- **Route context** — the object the router hands to every route, carrying the query client and
  the gateways. It is how a page receives its dependencies.
- **Query options** — a factory returning the key and the fetch function of one read, taking its
  gateway as argument.

## Tooling

- **Docs budget** — the size limit on agent docs: 16 KiB per `AGENTS.md`, 20 KiB per glossary, 300
  characters per glossary entry. Checked after an agent's edit, by `make check` and in CI.
- **Spec** — a Markdown file under `documentation/specs/` describing one feature from the user's
  side, written by the `to-spec` skill.
- **Ticket** — a Markdown file under `documentation/tickets/` describing one vertical slice of a
  spec and the tickets that block it, written by the `to-tickets` skill.
