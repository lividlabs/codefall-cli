# ADR-TS-02: Frontend State Management

## Status

Accepted — <date>

## Context

A frontend juggles four genuinely different kinds of state. Conflating them — e.g. one global store
holding server responses, derived UI flags, and form text — forces you to hand-roll cache
invalidation for server data and needlessly globalizes ephemeral state.

## Decision

Four categories, one tool each:

| State kind | Tool | Examples |
|---|---|---|
| Server state | **TanStack Query** | anything owned by an API / backend |
| Navigation state | **the router** | current screen, back stack, route params |
| Shared client state | **Zustand** | selection, layout, theme — client-owned, cross-component |
| Local UI state | **React `useState` / `useReducer`** | form fields, toggles, hover |

### Decision rule

Ask in order: server-owned? → TanStack Query. Router-owned? → the router. Shared client state? →
Zustand. Component-local? → React.

### Rules

- **Do not mirror server state into Zustand** (banned). Zustand may hold a *key/reference* (e.g. a
  selected id) that Query then resolves; the server data itself stays in Query.
- **Do not mirror navigation state into Zustand** (banned). The router owns the current route, back
  stack, and route params. Read them from the router; a store may hold state *keyed by* a route,
  never the route itself.
- Real-time transports (WebSocket, SSE) **patch or invalidate the Query cache** rather than living in
  a parallel store.
- Per ADR-TS-01, the query client and the stores are frontend adapter-ring pieces and stay
  framework-idiomatic (not in the DI container).

### Lifetimes on the Next.js topology

The four categories above are unchanged. What changes is **instantiation**, and getting it wrong
leaks one user's data to another.

- **A `QueryClient` is created per request on the server** and reused as a singleton in the browser.
  TanStack Query's SSR guide prescribes exactly this, "to prevent data leakage between requests."
- **Zustand stores are created per request**, via a provider at the component level — never as module
  globals. Zustand's Next.js guide: *"No global stores — Because the store should not be shared
  across requests, it should not be defined as a global variable."*
- **Server Components neither read nor write stores.** They cannot use hooks or context, and on this
  topology all state lives in client components anyway.

Contrast this with the DI container, which *is* a module-scoped global (ADR-TS-01). The rule behind
both: module-global is correct for stateless wiring and a data-leak bug for anything holding request
state.

### Server rendering, per route

Decide route by route, on one question: **is this publicly reachable and worth indexing?**

- **Yes** — prefetch on the server, `dehydrate` the client, and wrap the tree in a
  `HydrationBoundary`. The client then renders from a warm cache instead of re-fetching.
- **No** — fetch client-side as usual. An authenticated view gains nothing from server rendering.

The question is not static-versus-dynamic and not first-party-versus-user-generated. A user-authored
public list is dynamic, user-generated, and squarely in the first case.

## Consequences

- Server-state concerns (caching, refetch, retry, loading/error) are the library's, not bespoke
  code.
- Four tools to learn, but each does one job; the decision rule removes "where does this go?".
- The no-mirroring rules are essential and enforced in review.
- **Redux is deliberately not chosen** — heavier than this split needs, and it still requires a
  query layer for server state.
- On Next.js, the per-request rules are a correctness constraint rather than a preference, and their
  failure mode is cross-user data exposure under concurrency. Treat a module-level store or query
  client as a defect, not a shortcut.

## Related

- ADR-BASE-01 — Clean Architecture (the frontend's inner rings)
- ADR-TS-01 — Dependency Injection (the frontend DI boundary)
