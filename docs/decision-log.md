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

## Open

- **Name the capabilities.** The top-level components are still unknown. Start with one coarse
  component under `internal/`, not three speculative ones.
- **The composition root does not exist yet.** ADR-GO-01 stands, but `main.go` builds no injector
  because there is no object graph. Build it with the first component; the provider must return the
  interface, never the concrete type, and the root calls each component's exported registration
  function and resolves facades only (ADR-GO-01 as amended 2026-08-19).
- **The layer rules currently match no files.** `.golangci.yml` reports `0 issues` because there is
  no `internal/*/internal/domain` or `application` to check — per ADR-GO-02 that looks identical to
  a broken config. Re-prove both halves with the first component.

## Parking lot

Product detail heard during scaffolding but deliberately not acted on — it belongs to `specify` and
`architect`, not here.

- (empty — the project was scaffolded before its capabilities were described)
