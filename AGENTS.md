# codefall-cli

A Go command-line tool. One surface, one app, one module.

**State: one component.** `internal/doctor/` (`codefall doctor`) is the first component and the
reference for the rules below; `cmd/codefall/main.go` is the composition root and builds the
injector.

## Applicable ADRs

The why lives in the ADRs. This file is the operative rules only — never restate the reasoning here.

- [ADR-BASE-01 Clean Architecture](docs/adrs/ADR-BASE-01-clean-architecture.md) ·
  [ADR-BASE-02 Package-by-component](docs/adrs/ADR-BASE-02-package-by-component.md) ·
  [ADR-GO-01 Dependency Injection](docs/adrs/ADR-GO-01-dependency-injection.md) ·
  [ADR-GO-02 Boundary enforcement](docs/adrs/ADR-GO-02-boundary-enforcement.md) ·
  [ADR-GO-03 Optional values](docs/adrs/ADR-GO-03-optional-values.md)
- This project's own decisions: [ADR-001 Facade contracts](docs/adrs/ADR-001-facade-contracts.md) ·
  [ADR-002 CLI libraries](docs/adrs/ADR-002-cli-libraries.md). New ones use
  [`docs/adrs/_TEMPLATE.md`](docs/adrs/_TEMPLATE.md); decisions still moving live in
  [`docs/decision-log.md`](docs/decision-log.md).

## Structure

- Package-by-component: capabilities as directories under `internal/<component>/`. The component's
  root package **is** the facade; exported identifiers there are its whole public API.
- What crosses a facade is a **contract** — a struct of primitives, ids, `mo.Option[T]`, and
  collections of those, declared in the component's root package. A facade never accepts or returns a
  `domain` type; reference another component's entity by id (ADR-001).
- Clean layers nest inside the component's own `internal/`: `domain/` `application/`
  `infrastructure/` `presentation/`.
- Shared technical modules under `internal/shared/<module>/`, each its own facade.
- One composition root per app at `cmd/<app>/main.go`.

## Layer rules

- Dependencies point inward only. Interfaces live in `application/` with the use cases that need
  them; `domain/` = entities + value objects + errors.
- `domain/` imports the standard library and `samber/mo` (ADR-GO-03) and nothing else.
  `application/` adds only its own component's `domain/`. No `net/http`, `database/sql`, drivers, or
  CLI framework in either.
- Handlers and commands are thin; use cases never see delivery types.

## DI

- `samber/do`, one injector per app, built in the composition root. Each component exports a
  registration function from its facade package; the root calls it. A component's concrete types are
  named only in that function — `main` cannot import `internal/<component>/internal/*`, so the
  compiler enforces this.
- **The provider returns the interface, never the concrete type** (this code lives in the
  component's registration function, not in `main`):

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

## CLI

- Cobra (`github.com/spf13/cobra`) defines the command tree; `fang.Execute` runs it from
  `cmd/codefall/main.go`, the only importer of `github.com/charmbracelet/fang`. Fang owns help,
  usage, error, `--version`, `completion`, and `man` output — change it with `fang.With*` options,
  never with Cobra templates.
- Commands are built in a component's `presentation/`, exported through its facade, and mounted on
  the root in the composition root. Flags are `pflag` — never the standard library `flag`.
- A component's facade exports `Register(do.Injector)` and `Command(do.Injector) *cobra.Command`;
  `main` calls the first and mounts the second. A component with several top-level commands adds
  `Commands(do.Injector) []*cobra.Command` when that happens.
- Styling is Lip Gloss v2 (`charm.land/lipgloss/v2`), used where it helps. Plain text is the
  default; parsed output (anything piped or `--json`-style) is never styled; human output goes
  through a `colorprofile`-aware writer (`lipgloss.Fprint*` or `colorprofile.NewWriter`), never a
  rendered style written straight to `os.Stdout`.
- Prompts are Huh v2 (`charm.land/huh/v2`), in `presentation/` only, collecting values into a
  contract for a use case. A command that prompts by default also runs without prompting: every
  prompted value is also a flag, and if `stdin` is not a terminal and a value is missing, fail
  naming the flag — never block on input. Run with `RunWithContext`; `WithAccessible` follows the
  `ACCESSIBLE` environment variable.
- Charm's v2 generation only. Nothing that imports the v1 paths
  (`github.com/charmbracelet/{lipgloss,bubbletea,bubbles,huh}`) is added. A TUI, if ever needed, is
  Bubble Tea (`charm.land/bubbletea/v2`), already in the graph through Huh — no new ADR for the
  library.

## Enforcement

- `internal/` facades are enforced by the compiler — a reach-around fails `go build`.
- Layer direction is enforced by `depguard` strict allow-lists in `.golangci.yml`, which is generated
  for this project rather than copied from a template.
- Run `go build ./...`, `go vet ./...`, `golangci-lint run`, `go test ./...` — or `make check`.
- **Verifying the rules takes two checks, because there are two mechanisms.** A cross-component
  reach-around must fail `go build`. An outward layer import must fail `golangci-lint run` *while
  still compiling*. Checking only the first tests the compiler, not the configuration.

## Workflow

- Conventional commits, with a body that says *why*.
- Work happens on a branch or worktree, never on `main`. Stacked branches are fine for landing
  large work as smaller reviewable pieces.
- Deliberately minimal: no branching model beyond this is decided here.

## Gotchas

- A `depguard` rule whose `files` pattern matches nothing reports `0 issues` and exits 0, looking
  exactly like a rule that works. Only a deliberate violation proves it.
- A `do.Provide` whose provider returns the concrete type compiles and then fails at runtime with
  `could not find service`. There is no compile-time guard for this — Go rejects
  `func As[Alias any, Initial Alias]()` with `cannot use a type parameter as constraint` — so the
  provider's return type is the discipline.
- **`.golangci.yml` needs a new `application-layer` allow-list entry for every component you add** —
  `depguard`'s `allow` is literal prefix matching with no globs, so a new component's `application/`
  package silently loses access to its own `domain/` until its import path is named there. The
  `shared-modules` rule needs a deny entry per component for the same reason.
- **The `shared-modules` rule matches no files** until `internal/shared/` exists, and so reports `0
  issues` — indistinguishable from a broken config. Prove it against a deliberate violation when
  that directory arrives. The `domain-layer` and `application-layer` rules were re-proven on
  2026-08-27 against a Cobra import from `internal/doctor/internal/application/`.
- **`depguard` matches `_test.go` too.** Inner-layer tests are internal test packages (`package
  domain`, `package application`), and a `domain` test cannot import `os` — which is why the schema
  test that holds `schemas/settings.schema.json` equal to the domain constants lives in the facade
  package.
- **`go.mod` requires `samber/do`, `samber/mo`, Cobra, Fang, Lip Gloss v2, and `colorprofile`.** Huh
  (ADR-002) arrives with the first prompt. Add each library with the code that needs it, not up
  front.
- **A facade that returns a `domain` type compiles for its callers even though they cannot import
  that package.** Go's `internal/` rule restricts naming a package, not holding a value: `o :=
  orders.Find(id)` infers the type and `o.Total()` works, while `var o *domain.Order` does not
  compile. Neither the compiler nor `depguard` catches this — ADR-001 is a review rule.
- **A v1-generation Charm import compiles and quietly doubles the dependency graph.** On any change
  to `go.mod`, check `go mod graph` for a `github.com/charmbracelet/lipgloss` or `bubbletea` line
  without `/v2`. Nothing else catches it (ADR-002).
