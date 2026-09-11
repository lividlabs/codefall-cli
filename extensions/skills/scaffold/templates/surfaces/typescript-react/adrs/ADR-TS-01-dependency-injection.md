# ADR-TS-01: Dependency Injection

## Status

Accepted — <date>

## Context

Clean Architecture (ADR-BASE-01) has inner layers depend on interfaces that outer layers implement;
something must bind adapters to those interfaces at a composition root. Options are manual wiring or
a container. For TypeScript, **Inversify** is the mature IoC container and handles the common
"one interface, multiple implementations, rebind in tests" case cleanly.

## Decision

### Container

- **Inversify** across all TypeScript packages. **One composition root per app**; library packages
  export `ContainerModule`s the app's root loads. Concrete classes are named in each package's
  `ContainerModule`; the app's root only loads modules.
- On the **Next.js topology** the root is a module, not an entry point, because every route handler
  is its own entry. Build the container once at module load and cache it on `globalThis`, so dev
  hot-reload does not rebuild it per request:

  ```ts
  declare global { var __container: Container | undefined }
  export const container: Container = (globalThis.__container ??= buildContainer())
  ```

  **Two file kinds may touch the container, and no others**: a route handler, and a server-component
  page that prefetches for a publicly indexed route (ADR-TS-02). Both are entry points — the outermost
  edge of a request — and both do the same thing: resolve what they need and pass those use cases as
  arguments inward, into a `presentation/` handler or a client component's props. Neither the handler
  nor any component imports the container. That keeps the service-locator rule below intact by
  confining the locator to the boundary rather than by counting files.
  This container is deliberately module-global; see the profile's per-request trap for why the query
  client and the stores must not be.

### Injection discipline

- **Inject only interfaces** — every injected dependency is typed as an interface, never a concrete
  class. (Practical exception: config/value objects injected by token.)
- **Always name the token explicitly** — `@inject(TYPES.Thing)` on every constructor parameter, with
  the parameter's own type imported as `import type`. Never rely on Inversify inferring the type from
  emitted metadata.

  Implicit resolution reads `design:paramtypes`, which records runtime types. An interface has no
  runtime type — it erases to `Object` — so under **every** toolchain, `tsc` included, inference
  cannot resolve an interface-typed parameter. The rule above is what makes injection work at all,
  not a workaround for a particular bundler. It also removes the portability hazard: with tokens
  everywhere, the same code resolves identically under `tsc`, SWC, and Babel.

### Naming

- **Interface** = the bare Thing, no `I`-prefix — `UserRepository`, not `IUserRepository`.
- **Implementation** = `<Qualifier>Thing` — `PostgresUserRepository`, `InMemoryUserRepository`.
- Use **`Default`** as the qualifier for the single canonical implementation (`DefaultClock`); an
  edge-case sibling (`SystemClock`) does not force the original to be renamed.

### Tokens

- DI tokens live in a **`types.ts`** exporting `TYPES = { X: Symbol.for("X") }` (Inversify's
  convention). Tests **rebind** a token to a mock — the standard test seam.

### Frontend boundary (important)

- The container wires **framework-agnostic logic** (use cases, gateways, services). **React-native
  tools stay idiomatic** — the query client via its own provider, stores as hooks — and are **not**
  registered in the container. React reaches the container via context + a `useInjection`-style
  hook. Never use the container as a service locator sprinkled inside components.

### Non-TypeScript layers

- A non-TS layer (e.g. a Rust/Tauri shell) uses idiomatic **constructor injection / trait objects**,
  not Inversify.

## Consequences

- Decorators require `reflect-metadata` and tsconfig `experimentalDecorators`. **Leave
  `emitDecoratorMetadata` off.** It cannot help — an interface-typed parameter erases to `Object`, so
  there is nothing useful to emit — and turning it on lets a class-typed parameter with a forgotten
  token resolve under some toolchains and fail under others.
- Rebind-to-mock is a uniform, clean test seam across packages.
- Some bundle-size cost on the frontend; acceptable for app surfaces.
- The "magic" of a container is centralized in one composition root, not scattered.

## Related

- ADR-BASE-01 — Clean Architecture (the interfaces being bound)
- ADR-BASE-02 — Package-by-component (per-component modules the root loads)
- ADR-TS-02 — Frontend state (why the query client / stores stay outside the container)
