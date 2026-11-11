# apps/web

The frontend of the marketplace: a TanStack Start app whose pages reach the API only through gateways. It starts as a shell: a layout, a home page that checks the API's health, and the helpers every feature reuses.

## Language

Shared terms: [`domain.md`](../../documentation/glossary/domain.md), [`technical.md`](../../documentation/glossary/technical.md).

## Map

- `src/routes/` — pages, one file per URL, each with its `*.test.tsx`. `__root.tsx` is the document shell.
- `src/features/<feature>/` — a gateway (port), its HTTP and in-memory adapters, its queries. `health/` is the example to copy.
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
- **Imports inside `src/`** use the `#/` alias and no file extension.
- **Devtools are skipped in tests** (`import.meta.env.MODE !== 'test'` in `__root.tsx`): keep that guard.

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
