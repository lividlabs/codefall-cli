# The design document

The shape of `docs/designs/DESIGN-NNN-slug.md`. Read before designing (step 5) and again when
drafting (step 7).

## Contents

- The header
- Required sections
- Conditional sections
- Technical Context
- Hard Constraints
- Task Plan, staged
- Task Plan, after creation
- Local IDs are append-only

## The header

A title heading and two rows, the same shape `CONCEPT` and `SPEC` documents use. The body starts at
`## Overview`.

```markdown
# DESIGN-007: Stage context handoff

**Status:** Ready — 2026-08-31
**Related:** spec: SPEC-004 · adr: ADR-003
```

`Status` is always present. `Related` appears only when there is something to relate, and is a
labeled list, not a set of rows.

| Label | Holds |
| --- | --- |
| `spec` | The specification this design implements |
| `concept` | The concept that framed it — **only when there is no spec** |
| `adr` | The ADR or ADRs this design produced |

**Never record a hop you can derive.** A spec names its concept in its own `**Concept:**` row, so a
design records the spec alone. A design written from a concept with no spec records the concept. A
design with neither has no `Related` row until it produces an ADR.

## Required sections

| Section | Holds |
| --- | --- |
| **Overview** | The approach in a paragraph, plus the key decisions and why each was made |
| **Architecture** | Components, their relationships, and data flow. Mermaid where a diagram earns its place. May be a paragraph for small work |
| **Task Plan** | The staging table, which collapses to an ID mapping once the beads exist |

**Research findings go inline**, in Overview or Architecture, next to the decision they bear on.
There is **no sibling `research.md`**, and no `data-model.md`, `quickstart.md`, or `contracts/`.

## Conditional sections

Include a section only when its trigger fires. Nothing is written to fill a heading.

| Section | Include when |
| --- | --- |
| **Components and Interfaces** | New or changed public types, APIs, or contracts |
| **Data Models** | Anything persists, or the shape of state changes |
| **Error Handling** | New failure modes, or the fix *is* a failure mode |
| **Testing Strategy** | The test approach is non-obvious. Skip it when "add a regression test" covers it |
| **Technical Context** | A new dependency, or infrastructure is touched |
| **Alternatives Considered** | A real choice was made. This is also the ADR trigger |
| **Hard Constraints** | There are invariants worth asserting directly |

## Technical Context

A fill-in-the-blank block. Drop a line whose answer is "unchanged" rather than writing the word.

```markdown
## Technical Context

- **Language / version** — Go 1.23
- **Primary dependencies** — `samber/do`, `pgx/v5`
- **Storage** — Postgres 16, existing `bookings` schema
- **Testing** — `go test`, `testcontainers` for the integration pass
- **Target platform** — Linux container, deployed as the existing API service
- **Performance goals** — p99 under 200ms for the read path
- **Constraints** — no new outbound network dependency
- **Scale** — roughly 40 writes/second at peak, unchanged by this work
```

## Hard Constraints

EARS-lite: `THE SYSTEM SHALL …`, for invariants only — concurrency, data integrity, security
boundaries. Only the ones worth testing directly. This is not `codefall-specify`'s full EARS: a
criterion is what a consumer observes; a hard constraint is an invariant of the implementation.

```markdown
## Hard Constraints

- THE SYSTEM SHALL hold no more than one open write transaction per booking at a time.
- THE SYSTEM SHALL reject a context payload whose signature does not verify, before parsing it.
```

## Task Plan, staged

The Task Plan has two forms and is **rewritten in place** when it moves from the first to the
second. It is never appended to.

```markdown
## Task Plan
> Staged. Not yet in Beads.

| ID | Task | Depends on | Design ref |
|----|------|-----------|------------|
| T1 | Add `StageContext` type + serde | — | Components |
| T2 | Wire context load into `/scaffold` | T1 | Architecture |
| T3 | Emit context on `/conceptualize` exit | T1 | Data Models |
| T4 | Integration test: concept → scaffold | T2, T3 | Testing Strategy |
```

The local IDs make the dependency edges reviewable before Beads IDs exist. The **Design ref** column
names the section of this design that motivated the task, so `codefall-implement` opens the relevant
fifteen lines rather than the whole document.

## Task Plan, after creation

```markdown
## Task Plan
> Created in Beads 2026-08-29. Beads is authoritative.

Epic: booking-DESIGN-007 · T1→booking-DESIGN-007-T1 · T2→booking-DESIGN-007-T2 · T3→booking-DESIGN-007-T3 · T4→booking-DESIGN-007-T4
```

Beads is the source of truth once the issues exist; the staging table comes out. The mapping line
stays so a later run on a changed document can update the graph rather than duplicate it.

Bead identifiers are the document's own numbering behind the project's prefix —
`<prefix>-DESIGN-NNN` for the epic, `<prefix>-DESIGN-NNN-Tn` for each task — set by `bd rename`
after creation, per `beads.md`. The line lists them anyway: a replaced task's bead is
`<prefix>-DESIGN-NNN-Tn-2`, and the line is what says which bead is current.

## Local IDs are append-only

`T1`, `T2`, `T3` are permanent within a design. **A retired local ID is never reused.** Removing T2
from the table does not free `T2` for the next task added — that would silently repoint the mapping
line at a different bead.
