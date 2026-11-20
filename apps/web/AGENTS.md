# apps/web

The frontend of the marketplace: a TanStack Start app whose pages reach the API only through gateways. The home page shows a search bar and the latest annonces; `/recherche` lists results, `/annonces/$id` shows one with an unfolding contact form; `/annonces/deposer` deposits one; `/annonces/$id/supprimer` withdraws one (file `$id_.supprimer.tsx`: the `_` keeps it from nesting under the `$id` page, which has no outlet).

## Language

Shared terms: [`domain.md`](../../documentation/glossary/domain.md), [`technical.md`](../../documentation/glossary/technical.md).

## Map

- `src/routes/` — pages, one file per URL, each with its `*.test.tsx`. `__root.tsx` is the document shell.
- `src/features/<feature>/` — a gateway (port), its HTTP and in-memory adapters, its queries. `health/` is the smallest example (no page uses it any more), `annonce/` the full one.
- `src/components/` — `AnnonceCard` (result grid) and `RechercheForm` (search bar: mot-clé, catégorie, lieu and ville ou code postal, plus the prix bounds with `avecPrix` on `/recherche`; it takes the `regions` its page loaded), shared by `/` and `/recherche`.
- `src/test/annonces.ts` — `uneAnnonce` and `desAnnonces` fixtures for page tests.
- `src/gateways.ts` — the composition root: the only file that imports adapters.
- `src/lib/http.ts` — the only caller of `fetch`.
- `src/hooks/form.ts`, `src/components/FormComponents.tsx` — the form hook and its fields.
- `src/test/renderRoute.tsx` — renders a route on in-memory gateways.

Patterns, naming and the step-by-step workflow are in the `frontend-development` skill: **invoke it before adding or changing a feature**.

## Commands

From the repository root: `make test-web`, `make lint-web`, `make dev-web`. Inside `apps/web`:

```bash
pnpm dev              # http://localhost:3000
pnpm test             # Vitest, once
pnpm vitest run src/routes/index.test.tsx   # one file
pnpm typecheck && pnpm lint && pnpm check   # types, ESLint, Prettier
pnpm format           # fix formatting and lint
```

## Rules and pitfalls

- **A route never imports an adapter**: it reads `gateways` from `Route.useRouteContext()`.
- **Tests never mock `fetch` or a module**: they pass in-memory gateways to `renderRoute`.
- **`src/routeTree.gen.ts` is generated and committed**: never edit it by hand. `pnpm dev` and `pnpm build` rewrite it; after adding or renaming a route file without the dev server running, run `pnpm generate-routes` or the tests and the typecheck will not see the route.
- **Two base URLs**: the browser calls `/api/...`, which the Vite proxy forwards to the API; server-side rendering calls the API directly at `API_URL` (default `http://localhost:8080`). Both live in `src/lib/http.ts` and `vite.config.ts`: a gateway only gives the API path.
- **The API's error body is `{"erreur": "…", "champ"?: "…"}`**: `lib/http.ts` puts them in `ApiError.message` and `ApiError.champ`. A form shows an error with a `champ` under that field (`errorMap.onServer`, cleared on change), any other above its button. A gateway turns a 404 it expects into `null` (the page throws `notFound()` from its loader) rather than an error.
- **Format prices and dates with `features/annonce/format.ts`**: prices are centimes (zero is « Don »), dates use the Paris time zone so server and browser render the same text.
- **Imports inside `src/`** use the `#/` alias and no file extension.
- **`renderRoute` mounts on `document`**, as the app hydrates it: mounted in a div, the `<body>` of `__root.tsx` fights Spark's portals and the first re-render loops forever.
- **Search criteria live in the URL**: `/recherche` validates them with `rechercheSearchSchema` (unusable values dropped, never an error page) and loads through `loaderDeps`. The URL carries prix bounds in euros, `criteresDeRecherche` turns them into the centimes of `CriteresRecherche`. A new criterion is added to the schema, `CriteresRecherche`, `RechercheForm` and the HTTP gateway's query string; a changed criterion drops `page`. The lieu selector holds `region:84` or `departement:2A` and writes `region` or `departement` to the URL; `criteresDeRecherche` drops a code missing from the loaded régions, so the loader awaits `regionsQueryOptions` first.
- **The in-memory gateway mirrors the API's search** (every word, no case or accent, catégorie, prix bounds included, région or département, exact code postal or whole normalised ville, newest first, 20 a page); its default régions are a small subset of the API's table: change both together.
- **The in-memory gateway's `contacter` accepts `{ envoiImpossible: true }` options** to simulate a 502; use it in tests that assert the contact form shows the send-failure message.
- **Devtools are skipped in tests** (`import.meta.env.MODE !== 'test'` in `__root.tsx`): keep that guard.

## Design system: Spark UI

leboncoin's Spark: `@spark-ui/components` (import one component per path, `@spark-ui/components/button`) and `@spark-ui/theme-utils`, both imported in `src/styles.css` with an `@source` on the components' `dist`, plus Nunito Sans from `@fontsource-variable/nunito-sans`.

- **Spark resets Tailwind's palette and scales.** `text-sm`, `p-4`, `gap-2`, `gray-500`, `max-w-3xl`, `font-semibold` produce nothing. Use Spark tokens: colors `bg-main`, `text-on-surface`, `border-outline`, opacity `/dim-1`…`/dim-5`; spacing `p-lg`, `gap-md`, `mt-3xl`, sizes `sz-*`; type `text-display-1`, `text-headline-1`, `text-body-2-highlight`; borders `border-sm`; `rounded-md`, `z-sticky`. Check the built CSS when unsure.
- **Keep the `@theme static` block for `--spacing-factor` in `styles.css`.** Every Spark size is `calc(… * var(--spacing-factor))`, and Tailwind emits only the theme variables a class uses directly: without it, paddings, gaps, radii and text sizes all collapse while colors still work.
- **Form fields are Spark** (`FormField` + `Input`/`Textarea`/`Select` in `components/FormComponents.tsx`); a schema rule that mirrors the API must change with it (`features/annonce/schemas.ts`).
- **Toasts**: `ToastProvider` wraps the app in `__root.tsx`; a page calls `useToastManager().add(...)`, which survives a navigation.
- **Prefer a Spark component** (`Button`, `TextLink`, `FormField`, `Toast`…) over a hand-styled element; `Button asChild` wraps a router `Link`.
- **Theme**: Spark switches on `<html data-theme="light|dark">`. `ThemeToggle` and the init script in `__root.tsx` always write the resolved theme there; the `auto` choice lives only in `localStorage`. Keep both in sync.
- **pnpm peer warnings are expected**: Spark pins exact peers (`react` 19.2.4, `tailwindcss` 4.1.18) and we run newer patch or minor versions, which work; `@eslint/js` wants ESLint 10 while `@tanstack/eslint-config` runs on 9. `pnpm peers check` lists them.

<!-- intent-skills:start -->
# TanStack Intent - before using a router or Start API for the first time, run the matching command.
tanstackIntent:
  - id: "@tanstack/react-start#react-start"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/react-start#react-start"
    for: "React bindings for TanStack Start: project setup, StartClient, StartServer, useServerFn."
  - id: "@tanstack/start-client-core#start-core"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/start-client-core#start-core"
    for: "Start overview: Vite plugin, getRouter() factory, root route document shell, routeTree.gen.ts."
  - id: "@tanstack/start-client-core#start-core/execution-model"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/start-client-core#start-core/execution-model"
    for: "What runs on the server, on the client, or on both."
  - id: "@tanstack/start-client-core#start-core/server-functions"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/start-client-core#start-core/server-functions"
    for: "createServerFn, input validation, calling server code from the client."
  - id: "@tanstack/router-core#router-core"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/router-core#router-core"
    for: "Route trees, createRouter, root route with context, file naming conventions."
  - id: "@tanstack/router-core#router-core/data-loading"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/router-core#router-core/data-loading"
    for: "Route loader, loaderDeps, pending and error components, router context as dependency injection."
  - id: "@tanstack/router-core#router-core/navigation"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/router-core#router-core/navigation"
    for: "Link, useNavigate, active props, preloading, navigation blocking."
  - id: "@tanstack/router-core#router-core/path-params"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/router-core#router-core/path-params"
    for: "Dynamic segments ($id), splat routes, useParams."
  - id: "@tanstack/router-core#router-core/search-params"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/router-core#router-core/search-params"
    for: "validateSearch with Zod, reading and writing search params, loaderDeps."
  - id: "@tanstack/router-core#router-core/not-found-and-errors"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/router-core#router-core/not-found-and-errors"
    for: "notFound(), notFoundComponent, errorComponent."
  - id: "@tanstack/router-core#router-core/type-safety"
    run: "pnpm dlx @tanstack/intent@latest load @tanstack/router-core#router-core/type-safety"
    for: "Type inference rules: never cast, Register declaration, from narrowing on hooks and Link."
<!-- intent-skills:end -->
