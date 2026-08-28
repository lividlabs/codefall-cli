# Decision log

The scratchpad for decisions in flight. Settled decisions graduate to an ADR in `docs/adrs/` and are
struck from `Open`; this file never holds the reasoning that belongs in an ADR.

## Locked

Decided at scaffold, 2026-08-16.

- **One surface.** `codefall-cli` is a single Go CLI — no backing service, no UI, no seam. Profile:
  `go`. Inherited: ADR-BASE-01, ADR-BASE-02, ADR-GO-01, ADR-GO-02, ADR-GO-03.
- **Module path** — `github.com/lividlabs/codefall-cli`. No git remote set yet.
- **One app, one package.** Single Go module at the repo root; `cmd/codefall/` is the only binary.
  Codefall has no house opinion on repo layout, and one app doesn't need one.
- **Package-by-component**, chosen over ports-and-adapters even though the capabilities are unnamed —
  see the "For this project" section added to ADR-BASE-02.
- **No components yet, and no example one.** `internal/` is empty by choice: a placeholder component
  was scaffolded, reviewed, and deleted as ceremony that would be replaced wholesale in the next
  release. `main.go` prints the banner directly and wires nothing.
- **No dependencies.** `go.mod` requires nothing. `samber/do` (ADR-GO-01) and `samber/mo`
  (ADR-GO-03) are governing decisions, not yet imports — they arrive with the code that needs them.
- **Boundary enforcement is wired**, ahead of the code it will govern: `internal/` facades
  (compiler) plus `depguard` strict allow-lists in `.golangci.yml` (ADR-GO-02), both proven against
  deliberate violations at scaffold time.
- **Contracts cross facades, not entities.** A facade returns a data shape declared in the
  component's root package, never a `domain` type — recorded as `ADR-001`. Go's `internal/` rule
  does not cover this: it restricts naming a package, not holding a value, so a leaked entity
  compiles for a caller that cannot import it. Neither the compiler nor `depguard` can catch it.
- **`ADR-BASE-03` (extraction readiness) is deliberately not part of this project's set.** Its gate
  is whether a surface could ever be split into a service, and a CLI cannot. Four of its five rules
  have no referent here; the fifth is `ADR-001`. If a backing service ever appears, revisit.
- follows codefall `0.4.0` as of `2026-08-19`
- grafted codefall `0.4.0` on `2026-08-19` — took the `ADR-BASE-02` revision (re-applied in place,
  carrying the `For this project` amendment forward; no supersession, as no code had been written
  against it) and the `AGENTS.md` workflow defaults. Skipped `ADR-BASE-03` as inapplicable.
  `ADR-BASE-01` and `ADR-GO-01/02/03` were already current.
- **ADR-GO-01 amended, 2026-08-19.** Its claim that the composition root "is the only place that
  names concrete implementations" contradicted the layout: `cmd/` cannot import
  `internal/<component>/internal/*`. Concretes are named in each component's registration function;
  `main` calls those functions and resolves facades only. The same fix is owed to the plugin
  templates (go and typescript-react profiles).
- **CLI libraries, 2026-08-27.** Cobra for the command tree, run through Fang; Lip Gloss v2 for
  styled output where it helps; Huh v2 for prompts. Charm's v2 generation only, on the
  `charm.land/…/v2` paths — the two majors do not mix. Recorded as `ADR-002`. Bubble Tea arrives
  transitively with Huh and is the answer if a TUI is ever needed; that is a consequence of ADR-002,
  not a new decision. None of it is imported yet — each library arrives with the code that first
  calls it.
- **First component, 2026-08-27.** `internal/doctor/` is the first capability under `internal/`:
  `codefall doctor` reports whether a directory has `.codefall/settings.json`, Beads, and an
  authenticated `gh` with the scopes codefall needs. It reports and never repairs — it prints the
  remedy for every unmet check and `codefall init` will be the thing that runs them. The composition
  root now exists: `cmd/codefall/main.go` builds the injector, calls `doctor.Register`, mounts
  `doctor.Command`, and hands the tree to `fang.Execute`. The layer rules were re-proven the way
  ADR-GO-02 requires — a facade reach-around from `main` failed `go build`, and a Cobra import from
  `internal/doctor/internal/application/` compiled and failed `golangci-lint run` on the
  `application-layer` rule. Facade shape, settling ADR-002's open question:
  `Register(do.Injector)` plus `Command(do.Injector) *cobra.Command`, with
  `Commands(do.Injector) []*cobra.Command` when a component has several top-level commands. No
  report contract crosses the facade yet, because nothing outside doctor consumes one; add
  `doctor.Report`/`doctor.CheckResult` with the first external consumer, likely `init` (ADR-001).
  Now imported: Cobra, Fang, `samber/do`, `samber/mo`, `charm.land/lipgloss/v2`, and
  `github.com/charmbracelet/colorprofile`. Huh still waits for the first prompt.
  Bubble Tea and Bubbles v2 joined with doctor's spinner, in `presentation/` only — an ADR-002
  consequence rather than a new decision, and not the TUI question ADR-003 would settle.
  `schemas/settings.schema.json` is the published definition of `.codefall/settings.json`; the Go
  validator is hand-written in `domain/` so no schema library ships in the binary, and a test holds
  the two equal. One correction to ADR-002: Fang v1.0.0 does import `charm.land/lipgloss/v2`, so its
  note that Fang imports neither Lip Gloss major is out of date. Harmless — it is the v2 major, and
  the v1 generation stays out of the graph.
- **Second component, 2026-08-27.** `internal/setup/` is `codefall init`, the thing that creates
  what doctor checks. The package is named `setup` rather than `init` because `package init` is
  legal but importing it is not: Go rejects `import ".../internal/init"` outright, since `init` must
  be a func, so the import would need an alias. One aliased import in the whole codebase is a worse
  trade than naming the directory after what it does. The command is `init`, unaffected.
  A run is an ordered list of steps, each reporting Done or Skipped through an `Observer` the
  presentation layer implements twice: a spinner with the running step's title when stdout is a
  terminal, plain lines everywhere else. This change carries one step, `settings`; installing the
  Claude Code plugin and running `bd init` are the next two, appended to the same list. Settings
  that already exist are left alone and the step reports Skipped — `--force` rewrites them — so init
  is safe to run again once it has more steps than this. `--harness` is accepted and validated now,
  with `claude-code` the only value, because the steps after this one are the ones that need it.
  Huh v2 is now imported, the first prompt in the project (ADR-002). The tracker survey lists Jira
  and Linear with "(not yet available)" in the label and refuses them in the field's own validation:
  Huh v2 has no disabled option, and a tracker that is not offered at all reads as a tracker nobody
  thought of. The layer rules were re-proven the way ADR-GO-02 requires — a Cobra import from
  `internal/setup/internal/application/` compiled and failed `golangci-lint run` on the
  `application-layer` rule. Still no report contract crosses a facade: init does not consume
  doctor's.

## Open

- **UI composition.** Shared UI widgets — theme, styles, the colour-profile writer, key maps,
  reusable Bubble Tea models (list, table, status bar) — under `internal/shared/ui/`, generic over
  the data they show; rule 4 means that module never imports a business component. Each component's
  `presentation/` binds its own data to a shared widget, and a shell owns arrangement and navigation
  (`main` for the command tree; `main` or an `internal/tui/` component for a TUI — unsettled). A
  facade re-exports a view type by alias when the shell must name it. Alternatives seen and not
  taken: one shared module holding all presentation (breaks rule 4); a shell rendering generic
  widgets from contracts alone (widens every facade). Graduates to `ADR-003` with the first
  component that has a view, or a committed TUI, whichever comes first.

## Parking lot

Product detail heard during scaffolding but deliberately not acted on — it belongs to `specify` and
`architect`, not here.

- (empty — the project was scaffolded before its capabilities were described)
