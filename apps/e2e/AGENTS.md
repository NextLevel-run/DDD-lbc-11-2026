# apps/e2e

The Playwright suite that drives the real stack: browser, frontend, API and Mailcatcher. It owns its pnpm package and lockfile; it tests no code of its own and imports nothing from the other apps.

## Language

Shared terms: [`domain.md`](../../documentation/glossary/domain.md), [`technical.md`](../../documentation/glossary/technical.md).

## Map

- `playwright.config.ts` — starts its own API (`:18080`, `DEMO_DATA=off`) and frontend (`:13000`), so it never touches the state of `make dev`. Ports live in `ports.ts`.
- `tests/*.spec.ts` — one file per journey or concern; `tests/support.ts` holds the helpers (`ouvrir`, `deposer`, `deposerEnLigne`, `unJeton`, Mailcatcher lookups). `parcours-vendeur` also runs the correction journey: challenge on the hidden moderation pages, the email link, the correction, a second validation.
- Mailcatcher (`:1025` SMTP, `:1080` API) is shared with `make mail` and must be running.

## Commands

From the repository root: `make e2e` (starts Mailcatcher if `:1080` does not answer, installs chromium, runs) and `make lint-e2e` (Prettier + `tsc`, no browser, part of `make check`). CI runs the suite in its own `e2e` job; `make check` never launches a browser.

## Scope

Only what needs the real stack: server rendering, the HTTP gateway against the API, real CSS, asynchronous email. Page behaviour belongs in Vitest on in-memory gateways, API cases in Go handler tests.

## Invariants and pitfalls

- **Load every page through `ouvrir(page, url)`**: it waits for React hydration. Acting earlier submits the server-rendered form natively or clicks a dead button, a past cause of flaky failures.
- **Tests share one in-memory API**: each builds its own annonces around `unJeton()` and searches by that token. Never assert on global counts or demo data (they are off).
- **Set data up through the API**: the deposit form is exercised once, in `parcours-vendeur`. A deposited annonce awaits moderation and is online nowhere: `deposer` leaves it so, `deposerEnLigne` has it validated through the moderation API, for a test that searches or opens it.
- **Find an email by its unique recipient** in Mailcatcher's `/messages`, never by position; a recipient who gets several emails (confirmation, challenge) is told apart by a `sujet` fragment.
