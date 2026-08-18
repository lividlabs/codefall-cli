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

## Open

- **Name the capabilities.** The top-level components are still unknown. Start with one coarse
  component under `internal/`, not three speculative ones.
- **CLI framework.** A library is expected to run the whole thing (`cobra`, `urfave/cli`, or
  similar) but none is chosen. It is a `presentation/` concern and must stay out of `domain/` and
  `application/` — ADR-GO-02 rule 5. Record the choice as `ADR-001`.
- **The composition root does not exist yet.** ADR-GO-01 stands, but `main.go` builds no injector
  because there is no object graph. Build it with the first component; the provider must return the
  interface, never the concrete type.
- **The layer rules currently match no files.** `.golangci.yml` reports `0 issues` because there is
  no `internal/*/internal/domain` or `application` to check — per ADR-GO-02 that looks identical to
  a broken config. Re-prove both halves with the first component.

## Parking lot

Product detail heard during scaffolding but deliberately not acted on — it belongs to `specify` and
`architect`, not here.

- (empty — the project was scaffolded before its capabilities were described)
