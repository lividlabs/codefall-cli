# ADR-001: Facade Contracts, Not Entities

## Status

Accepted — 2026-08-19

## Context

ADR-GO-02 enforces the import half of component boundaries: `internal/` facades decide *who* may
call a component, and `depguard` decides which layers may import which. Neither says anything about
*what* crosses the facade. A caller with a legal import can still receive another component's domain
entity — methods, invariants, and all — and from that moment every change to the entity is a change
to the caller. The import graph stays clean while the coupling grows. Go's `internal/` rule does not
close this: it restricts *importing* a package, not *holding* its values. A facade that returns a
`domain` entity compiles for callers that cannot import `domain` at all — type inference hands them
the value, methods and exported fields included — so neither the compiler nor `depguard` catches the
leak, because no import line exists to flag.

Codefall's ADR-BASE-03 (extraction readiness) carries this rule as one of five, but that ADR is
gated on surfaces that can be split into services, and a CLI cannot be — it is correctly absent from
this project. Four of its five rules have no referent here. This one does: entity coupling across
facades hurts a CLI exactly as much as a platform, because its payoff is encapsulation, not
deployment. So the one rule is adopted on its own, as a project decision.

## Decision

- What crosses a component facade is a **contract** — a data shape of primitives and ids — never a
  domain entity.
- The entity stays inside the component that owns it. A component that needs another's data receives
  a shape defined for that exchange (an `OrderSummary`), not the owner's model (`orders.Order`).
- Reference another component's records **by id**, and resolve them through that component's facade.

## Consequences

- Facades define and map data shapes, so a cross-component call costs a mapping that passing the
  existing object would not. Accepted: the shape is a contract chosen on purpose; the entity is
  coupling by default.
- This is a review concern, not a lint rule — the types crossing a facade are visible in review, but
  no tool here checks them. The operative rule lives in AGENTS.md.
- If this project ever grows a surface that can be split, ADR-BASE-03 arrives with its other four
  rules and repeats this one; it may supersede this ADR then.

## Related

- ADR-BASE-02 — Package-by-component (the boundary this rule keeps meaningful)
- ADR-GO-02 — Boundary enforcement (the import half of the same concern)
- ADR-BASE-01 — Clean Architecture (entities live in `domain/`, inside their component)
