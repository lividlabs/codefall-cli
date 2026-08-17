# codefall-cli

A Go command-line tool. One surface, one app, one module.

**State: scaffolded, empty.** `internal/` has no components and `cmd/codefall/main.go` prints a
banner without wiring anything. The rules below are the ones the first component must satisfy — they
are not describing code that exists yet.

## Applicable ADRs

The why lives in the ADRs. This file is the operative rules only — never restate the reasoning here.

- [ADR-BASE-01 Clean Architecture](docs/adrs/ADR-BASE-01-clean-architecture.md) ·
  [ADR-BASE-02 Package-by-component](docs/adrs/ADR-BASE-02-package-by-component.md) ·
  [ADR-GO-01 Dependency Injection](docs/adrs/ADR-GO-01-dependency-injection.md) ·
  [ADR-GO-02 Boundary enforcement](docs/adrs/ADR-GO-02-boundary-enforcement.md) ·
  [ADR-GO-03 Optional values](docs/adrs/ADR-GO-03-optional-values.md)
- This project's own decisions start at `ADR-001` — none yet. New ones use
  [`docs/adrs/_TEMPLATE.md`](docs/adrs/_TEMPLATE.md); decisions still moving live in
  [`docs/decision-log.md`](docs/decision-log.md).

## Structure

- Package-by-component: capabilities as directories under `internal/<component>/`. The component's
  root package **is** the facade; exported identifiers there are its whole public API.
- Clean layers nest inside the component's own `internal/`: `domain/` `application/`
  `infrastructure/` `presentation/`.
- Shared technical modules under `internal/shared/<module>/`, each its own facade.
- One composition root per app at `cmd/<app>/main.go`.

## Layer rules

- Dependencies point inward only. Interfaces live in `application/` with the use cases that need
  them; `domain/` = entities + value objects + errors.
- `domain/` imports the standard library and nothing else. `application/` adds only its own
  component's `domain/`. No `net/http`, `database/sql`, drivers, or CLI framework in either.
- Handlers and commands are thin; use cases never see delivery types.

## DI

- `samber/do`, one injector per app, built in the composition root. It is the only place naming
  concrete implementations.
- **The provider returns the interface, never the concrete type:**

  ```go
  do.Provide(injector, func(i do.Injector) (application.OrderRepository, error) {
      return infrastructure.NewPostgresOrderRepository(do.MustInvoke[*sql.DB](i)), nil
  })
  ```

- Resolve with `do.Invoke` / `do.MustInvoke`. Use `do.As[Concrete, Interface]` to register an alias
  when a concrete type must also be reachable by interface. Avoid `do.InvokeAs` — it is a runtime
  scan and goes ambiguous once two services satisfy the same interface.
- Inject interfaces only, declared by the consumer in `application/`.
- **One interface per gateway role per component**, not per use case. Split narrower only when a
  different component needs a subset. If two interfaces in the same `application/` package differ by
  fewer than two methods, they are one interface.
- `Thing` / `<Qualifier>Thing` naming — `OrderRepository`, `PostgresOrderRepository`, `DefaultClock`.
  No `I` prefix, no `-er` suffix on domain roles. One implementation may satisfy several interfaces;
  do not force a canonical pairing.
- Tests pass hand-written fakes to the same constructors. Compile-time assertions, if written, go in
  `_test.go` in the pointer form: `var _ application.OrderRepository = (*PostgresOrderRepository)(nil)`.

## Optional values

- Absence is `mo.Option[T]`, not `*T` and not a sentinel zero value. `mo.Some("")` is present with an
  empty value and is not the same as `mo.None[string]()`.
- Errors stay `(T, error)` with `%w` wrapping and `errors.Is` / `errors.As`. Do not use `mo.Result`.
- If a caller needs to know *why* something is missing, it is an error, not a `None`.

## Enforcement

- `internal/` facades are enforced by the compiler — a reach-around fails `go build`.
- Layer direction is enforced by `depguard` strict allow-lists in `.golangci.yml`, which is generated
  for this project rather than copied from a template.
- Run `go build ./...`, `go vet ./...`, `golangci-lint run`, `go test ./...` — or `make check`.
- **Verifying the rules takes two checks, because there are two mechanisms.** A cross-component
  reach-around must fail `go build`. An outward layer import must fail `golangci-lint run` *while
  still compiling*. Checking only the first tests the compiler, not the configuration.

## Gotchas

- A `depguard` rule whose `files` pattern matches nothing reports `0 issues` and exits 0, looking
  exactly like a rule that works. Only a deliberate violation proves it.
- A `do.Provide` whose provider returns the concrete type compiles and then fails at runtime with
  `could not find service`. There is no compile-time guard for this — Go rejects
  `func As[Alias any, Initial Alias]()` with `cannot use a type parameter as constraint` — so the
  provider's return type is the discipline.
- **`.golangci.yml` needs a new `application-layer` allow-list entry for every component you add** —
  `depguard`'s `allow` is literal prefix matching with no globs, so a new component's `application/`
  package silently loses access to its own `domain/` until its import path is named there. Rule 4
  (shared modules) is commented out in that file for the same reason: it has no component facade to
  deny yet.
- **The layer rules match no files today** and so report `0 issues` — indistinguishable from a broken
  config. They were proven against deliberate violations at scaffold time; prove them again with the
  first component.
- **`go.mod` requires nothing.** `samber/do` and `samber/mo` are decided (ADR-GO-01, ADR-GO-03) but
  not yet imported. Add each with the code that needs it, not up front.
- **A CLI framework is expected but not chosen** (docs/decision-log.md, Open). Whichever it is, it is
  a `presentation/` concern — it must not appear in `domain/` or `application/`.
