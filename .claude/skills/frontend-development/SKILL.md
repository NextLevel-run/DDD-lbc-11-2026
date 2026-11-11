---
name: frontend-development
description: Guide for building features in the TanStack Start frontend (apps/web) with dependency inversion. Use when adding or changing a page, a route, a form, a call to the API, or a frontend test.
---

# Frontend Development

How a feature is built in `apps/web`. Every path below is relative to `apps/web/`. The vocabulary (feature, gateway, adapter, route context) is defined in `documentation/glossary/technical.md`; business names come from `documentation/glossary/domain.md`.

The frontend follows the same dependency rule as the API: the UI depends on a **port**, never on `fetch`.

```
routes/ (pages)  ──▶  features/<feature>/queries.ts  ──▶  <Feature>Gateway (port)
                                                              ▲            ▲
                                                   http-…-gateway   in-memory-…-gateway
                                                     (the app)          (the tests)
```

`features/health/` and `routes/index.tsx` are the worked example: read them before writing a feature.

## Anatomy of a feature

```
src/features/<feature>/
├── <feature>-gateway.ts            # Port: the interface and the types it speaks
├── http-<feature>-gateway.ts       # Adapter: calls the API through lib/http.ts
├── in-memory-<feature>-gateway.ts  # Adapter: holds state in memory, for tests
├── queries.ts                      # queryOptions / mutationOptions taking the gateway
└── schemas.ts                      # Zod schemas of the forms (when there is a form)
src/routes/<path>.tsx               # Page: reads the gateway from the route context
src/routes/<path>.test.tsx          # Test: renders the page on in-memory gateways
```

## Rules

1. **A route never imports an adapter.** It reads its port from the route context: `const { gateways } = Route.useRouteContext()`. Only `src/gateways.ts` imports `http-*` and `in-memory-*` files.
2. **Only `lib/http.ts` calls `fetch`**, and only `http-*-gateway.ts` files call `lib/http.ts`.
3. **The port speaks the UI's language.** Its types are what the pages need, named with the domain glossary. When the API's JSON differs, the HTTP adapter translates; the DTO never leaves the adapter.
4. **Server state lives in TanStack Query.** No `useEffect` to fetch, no copy of fetched data in `useState`.
5. **A test never mocks `fetch` or a module.** It passes an in-memory gateway to `renderRoute`.
6. **`routeTree.gen.ts` is generated**: never edit it; `pnpm generate-routes` (or the dev server) rewrites it.

## Workflow: adding a feature

### 1. Port

Declare what the pages need in `features/<feature>/<feature>-gateway.ts`: an interface whose methods return promises, and the types they exchange.

```ts
export interface Health {
  status: 'ok'
}

export interface HealthGateway {
  check: () => Promise<Health>
}
```

### 2. In-memory adapter

Write `in-memory-<feature>-gateway.ts` first: it is what the tests run on. A factory returning the port, holding its state in a closure, with options for the cases a test needs to set up (seed data, a failure).

### 3. Register the port

In `src/gateways.ts`, add the port to the `Gateways` interface and the in-memory adapter to `createInMemoryGateways`. TypeScript now refuses to build until `createHttpGateways` provides it too: that is step 6.

### 4. Queries

In `queries.ts`, one factory per read or write, taking the gateway:

```ts
export const healthQueryOptions = (gateway: HealthGateway) =>
  queryOptions({ queryKey: ['health'], queryFn: () => gateway.check() })
```

A mutation invalidates the query keys it makes stale in its `onSuccess`.

### 5. Page and test

Write the test first, in `routes/<path>.test.tsx`:

```tsx
await renderRoute('/', {
  gateways: { health: createInMemoryHealthGateway({ reachable: false }) },
})
expect(await screen.findByText('API unreachable')).toBeInTheDocument()
```

Query by role and visible text, as a user would. Then the page, in `routes/<path>.tsx`: `createFileRoute`, the gateway from `Route.useRouteContext()`, `useQuery` / `useMutation` on the factories of step 4. Render the pending, error and empty states, not only the success.

Data a page cannot render without goes in the route `loader` (`context.queryClient.ensureQueryData(...)`) and is read with `useSuspenseQuery`.

### 6. HTTP adapter

Write `http-<feature>-gateway.ts` with `request<T>()` from `lib/http.ts` and register it in `createHttpGateways`. `request` throws an `ApiError` carrying the HTTP status and the API's `error` message. Paths are the API's own (`/health`): the `/api` prefix and the server-side base URL are `lib/http.ts`'s business.

### 7. Verify

`make check` from the repository root, then look at the page in the browser with the API running (`make dev`).

## Forms

Use `useAppForm` from `hooks/form.ts` with the field components of `components/FormComponents.tsx` (`TextField`, `TextArea`, `Select`, `SubscribeButton`). Validate with a Zod schema from the feature's `schemas.ts`, passed to the form's `validators`. Submit through a mutation; show an `ApiError`'s message next to the form rather than swallowing it.

## Naming

| Thing | Convention | Example |
| --- | --- | --- |
| Feature folder | kebab-case, a domain glossary name | `features/classified-ad/` |
| Port | `<Feature>Gateway` | `ClassifiedAdGateway` |
| Adapter factory | `create<Kind><Feature>Gateway` | `createHttpClassifiedAdGateway` |
| Query factory | `<what>QueryOptions` | `classifiedAdQueryOptions` |
| Test name | `it('does X when Y')`, no _should_ | `it('says the API is unreachable when the health check fails')` |

Imports inside `src/` use the `#/` alias (`#/features/health/queries`), without file extension.

## TanStack specifics

`AGENTS.md` lists the TanStack Intent commands that load the up-to-date guidance for routing, data loading, search params and server functions. Run the matching one before using an API of the router or of Start you have not used in this codebase yet.
