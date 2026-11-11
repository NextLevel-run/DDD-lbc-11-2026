# AGENTS.md — second-hand marketplace

Shared grounding for coding agents working on this repository: a pedagogical second-hand marketplace built with Domain-Driven Design and hexagonal architecture. Follow the pointers for details.

## Stack

- **Monorepo** — `apps/api` and `apps/web`, driven by the root `Makefile`.
- **API** (`apps/api`) — Go, standard library HTTP server, one bounded context per folder under `internal/`, in-memory adapters only (no database).
- **Frontend** (`apps/web`) — TanStack Start (React 19, file-based routes), TanStack Query and Form, Tailwind CSS, Vitest + Testing Library, pnpm.

| Service  | Port | Note                                    |
| -------- | ---- | --------------------------------------- |
| Frontend | 3000 | Vite dev server, proxies `/api` → 8080  |
| API      | 8080 | `GET /health` answers `{"status":"ok"}` |

## Essential commands

Run them from the repository root.

```bash
make install   # API modules + frontend dependencies
make dev       # API and frontend together
make check     # everything CI runs: docs budget, format, lint, typecheck, tests, builds
make test      # tests of both apps
make up        # the same two services in Docker Compose
```

`make help` lists the rest. Run `make check` before saying a change is done.

## Vocabulary

The two glossaries are loaded with this file and are the only ones: [`technical.md`](documentation/glossary/technical.md) (architecture) and [`domain.md`](documentation/glossary/domain.md) (business). There is no `CONTEXT.md`: a skill that asks for one means `domain.md`.

- Use the glossaries' names, spelled as in the code.
- Add a missing term to its glossary in the same change as the code that uses it; never invent one in code. Each glossary's header says how an entry is written.

## Contribution rules

- **Ask before coding.** Clarify the requirement first; the `grill-me` skill does it thoroughly.
- **Read the local `AGENTS.md` before editing a module**: `apps/api/`, `apps/web/`, and each bounded context under `apps/api/internal/`.
- **Use the matching skill**: `tactical-ddd` for the API, `frontend-development` for the frontend, `scaffold-bc` for a new bounded context.
- **Dependencies point inward**, in both apps: adapters depend on the application and the domain, never the reverse. The domain imports no framework and no adapter.
- **One bounded context never imports another.** They talk through events.
- **Test on in-memory adapters**, in both apps. No mock of `fetch`, no mocking library for ports.
- **Write everything in English**: code, tests, docs, commit messages.
- **Secrets**: never commit a `.env`.

## Working on a feature

1. `/grill-with-docs` — challenge the idea and settle its vocabulary in `domain.md`.
2. `/to-spec` — write the spec to `documentation/specs/`.
3. `/to-tickets` — split it into vertical slices under `documentation/tickets/`.
4. Implement one ticket at a time with `tactical-ddd` and `frontend-development`, tests first.
5. `make check`, then the `docs-scribe` agent before committing.

`codebase-design` helps when a module's shape is in question; `to-questionnaire` when someone else holds the answer; `handoff` to pass the work to a fresh session.

## Agent docs

- An `AGENTS.md` is a map plus the rules an agent would break, never an inventory of the code. A bounded context's has five sections: role, Language, Boundaries, Flows, Invariants and pitfalls.
- Budget: 16 KiB per `AGENTS.md`, 20 KiB per glossary, 300 characters per glossary entry (`tools/docs-budget/`). A hook checks it right after an edit; `make check` and CI check it too. Over budget means compact, not split.
- The `docs-scribe` agent keeps these files in sync with the code.

## Pointers

- [`apps/api/AGENTS.md`](apps/api/AGENTS.md) — API layout, commands and rules
- [`apps/web/AGENTS.md`](apps/web/AGENTS.md) — frontend layout, commands and rules
- [`documentation/`](documentation/) — glossaries, specs, tickets
- [`prompts/`](prompts/) — the prompts worth keeping, one file each
