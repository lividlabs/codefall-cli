---
name: design
description: Decide how a feature gets built and put the work into the graph — read the docs and the code, judge whether the change warrants a design document at all, write one scaled to the work at docs/designs/, record hard-to-reverse choices as ADRs, and create the task graph in Beads from the document's staged task plan.
argument-hint: "[the spec, the concept, or what you want built]"
disable-model-invocation: true
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Write
  - Edit
  - Bash
  - WebSearch
  - WebFetch
---

# Design

Decide how a feature gets built, and put the work into the dependency graph so `implement` can pick
it up.

There are two outputs and the second is the one that always exists. The first is a **design document**
at `docs/designs/DESIGN-NNN-slug.md`, holding the approach, the architecture, and a staged task plan.
The second is **the tasks in Beads**, with their dependency edges — the graph `implement` walks.

Not every change earns a document. A one-component bug fix with two tasks gets beads and nothing
else. The document is scaled to the work, and the work decides the tier — see
[Scale the artifact to the work](#scale-the-artifact-to-the-work).

Designing is not implementing. Once the document is written and the graph exists, stop.

Plugin paths in this document — the ones that start with `../` — are relative to this skill's
directory, the one holding this `SKILL.md`. Resolve them from where the file lives; they are not
relative to the user's project.

## Scope — how, not what and not whether

| In scope | Out of scope | Whose |
| --- | --- | --- |
| The approach, and the decisions inside it | Whether this is worth building | `conceptualize` |
| Components, their relationships, and data flow | What a consumer observes when it works | `specify` |
| Contracts, types, schemas, storage, failure modes | What the screen looks like | `mock-up` |
| Which existing code changes and which is new | The project's architecture stance | `scaffold` |
| The task breakdown and its dependency edges | Writing the code | `implement` |
| Hard-to-reverse choices, recorded as ADRs | Estimates and assignment | the team |

The line that matters most is the one above `specify`. **A design does not restate the
specification.** If you find yourself writing what a user will observe, you are copying a document
that already exists — cite it and move on.

The line below `scaffold` matters nearly as much. **The project's layering, its component
boundaries, and how they are enforced are already decided**, in `docs/adrs/` and the scoped
`AGENTS.md` files. Read them and design within them. A design that needs the stance changed is
saying so out loud, once, and then either following the ADR or writing a project ADR that supersedes
it — never quietly ignoring it.

## Scale the artifact to the work

The same philosophy `conceptualize` applies to concepts: a small required core, everything else
conditional, and **no empty or placeholder sections — omit them.**

Three tiers. The work picks the tier; the user can overrule it.

| Tier | When | Output |
| --- | --- | --- |
| **0 — no document** | All three hold: the change is contained to one component, no public interface changes, and the task graph is one or two beads | Beads only |
| **1 — minimal document** | Anything that crosses a component boundary, or fans out past roughly three dependent tasks | The three required sections |
| **2 — full document** | The same, plus any conditional section whose trigger fires | Required three plus what triggered |

**Simple bug fixes land at tier 0.** Create the bead directly with the reproduction and the cause,
and let `implement` pick it up. Writing a design document for a null check is the ceremony this tier
exists to avoid.

**Tier 1 and tier 2 are not a decision you make separately.** Write the required three, then walk
the conditional triggers; whichever fire, fire. A design with none of them is a tier 1 design, and
nothing announces that.

**An ADR is not gated on the tier.** A hard-to-reverse choice can turn up in a one-component fix,
and then that fix produces an ADR and no design document. See [ADRs](#adrs).

**Say the tier out loud, with the reason, before writing anything.** The judgement is yours to make
and the user's to overrule:

> This touches one component, changes nothing public, and comes to two tasks. I would skip the
> design document and create the beads directly. Want the document anyway?

Push back once if you disagree with their answer, then do what they ask. Per the repo's rules,
disagreeing about size is not a refusal.

## The document

`docs/designs/DESIGN-NNN-slug.md`. Git-tracked, so it diffs in a pull request and a later `design`
run can grep it.

Three digits, zero-padded, taken as the highest existing number plus one. **A design is never
renumbered and its identifier is never reused**, including after it is archived — the file moves,
the identifier does not change. That is the same discipline this repo applies to ADR history,
concept identifiers, and spec identifiers.

### The header

A title heading and two rows, the same shape `CONCEPT` and `SPEC` documents use. The body starts at
`## Overview`.

```markdown
# DESIGN-007: Stage context handoff

**Status:** Ready — 2026-08-31
**Related:** spec: SPEC-004 · adr: ADR-003
```

`Status` is always present. `Related` appears only when there is something to relate.

**`Related` is a labeled list, not a set of rows.** That is `conceptualize`'s convention, and it is
taken here for one reason: a new kind of upstream artifact costs a label rather than an edit to this
template and this skill. When a bug-report verb lands, its designs carry `bug: BUG-012` and nothing
else moves.

| Label | Holds |
| --- | --- |
| `spec` | The specification this design implements |
| `concept` | The concept that framed it — **only when there is no spec** |
| `adr` | The ADR or ADRs this design produced |

**Never record a hop you can derive.** A spec already names its concept in its own `**Concept:**`
row, so a design that carries both is keeping a second copy of a link nothing keeps in sync. Record
the spec, and a reader wanting the concept opens the spec. A design written from a concept with no
spec records the concept, because then it is the only upstream link there is.

A design with neither has no `Related` row at all until it produces an ADR.

### Required sections

| Section | Holds |
| --- | --- |
| **Overview** | The approach in a paragraph, plus the key decisions and why each was made |
| **Architecture** | Components, their relationships, and data flow. Mermaid where a diagram earns its place. May be a paragraph for small work |
| **Task Plan** | The staging table, which collapses to an ID mapping once the beads exist |

**Research findings go inline**, in Overview or Architecture, wherever they bear on a decision.
There is **no sibling `research.md`**, and no `data-model.md`, `quickstart.md`, or `contracts/`
either. A finding that is not attached to a decision is a note nobody reads; a finding split into
its own file is a decision's reasoning filed away from the decision.

### Conditional sections

Include a section only when its trigger fires. Nothing here is written to fill a heading.

| Section | Include when |
| --- | --- |
| **Components and Interfaces** | New or changed public types, APIs, or contracts |
| **Data Models** | Anything persists, or the shape of state changes |
| **Error Handling** | New failure modes, or the fix *is* a failure mode |
| **Testing Strategy** | The test approach is non-obvious. Skip it when "add a regression test" covers it |
| **Technical Context** | A new dependency, or infrastructure is touched |
| **Alternatives Considered** | A real choice was made. This is also the ADR trigger |
| **Hard Constraints** | There are invariants worth asserting directly |

**Technical Context is a fill-in-the-blank block**, taken from spec-kit, and it is filled in because
prose hand-waves where a list cannot:

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

Drop a line whose answer is genuinely "unchanged" rather than writing the word.

**Hard Constraints is EARS-lite.** `THE SYSTEM SHALL …`, for invariants only — concurrency, data
integrity, security boundaries. Not the whole document, just the ones worth testing directly:

```markdown
## Hard Constraints

- THE SYSTEM SHALL hold no more than one open write transaction per booking at a time.
- THE SYSTEM SHALL reject a context payload whose signature does not verify, before parsing it.
```

`specify` writes acceptance criteria in full EARS, and this is deliberately not that. A criterion is
what a consumer observes; a hard constraint is an invariant of the implementation, which the
specification cannot state because it does not know how the thing is built.

## Task Plan

The task plan exists in two forms, and it is **rewritten in place** when it moves from the first to
the second. It is never appended to.

### Staged

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

**The local IDs exist so the dependency edges are reviewable before Beads IDs exist.** That is their
whole purpose. A flat list of tasks with no edges does not catch the thing review is for — a wrong
ordering, or a missing prerequisite — and by the time the beads exist the graph is already wired.

**The Design ref column is the traceability link.** It names the section of this design that
motivated the task, so `implement` gets a pointer to the relevant fifteen lines rather than the
whole document. Kiro's `tasks.md` cites requirement identifiers here; this cites the design section,
because the requirement is one link further out and the section is what the implementer actually
needs open.

### After creation

```markdown
## Task Plan
> Created in Beads 2026-08-29. Beads is authoritative.

Epic: bd-a2g · T1→bd-unz · T2→bd-s58 · T3→bd-cj4 · T4→bd-p71
```

**Why it collapses.** Beads is the source of truth the moment the issues exist. A duplicate task
list left behind in a git-tracked document will diverge from the graph, and nobody will update it.

**Why the mapping survives.** A later `design` run on a changed document needs to know which beads
this design already produced, so it can update them rather than duplicating the graph. The mapping
line is how it knows.

Bead identifiers are `<prefix>-<hash>`, where the prefix is the project's — `bd-unz`, `booking-a2g`.
They are not sequential; take them from the creation output rather than predicting them.

### Local IDs are append-only

`T1`, `T2`, `T3` are permanent within a design, exactly as requirement and criterion numbers are
permanent within a spec. **A retired local ID is never reused.** Removing T2 from the table does not
free `T2` for the next task added — that would silently repoint the mapping line at a different
bead.

## ADRs

An ADR is a separate artifact, `docs/adrs/ADR-NNN-title.md`, in the classic Nygard shape the project
already uses — Status, Context, Decision, Consequences, Related. Write it from
`../scaffold/templates/adrs/_TEMPLATE.md`.

**The number continues the project's own sequence.** Read `docs/adrs/`, take the highest bare
`ADR-NNN` plus one, starting at `ADR-001`. The prefixed sequences — `ADR-BASE-NN` from the
stack-agnostic core, `ADR-<PREFIX>-NN` from a surface profile — are inherited stance and are never
continued here. `scaffold` writes ADR-001 when a project has a client-server seam, so a scaffolded
project may already have one.

**The trigger** is the design's **Alternatives Considered** section containing a choice that is
hard to reverse or that other components will build on:

- a new dependency;
- a schema or protocol decision other components will be written against;
- a rejected alternative that cost real analysis.

Most designs need none. A choice between two obvious implementations of a private function is not an
ADR, however carefully it was made.

**Do not duplicate the rationale in both documents.** The design's Alternatives Considered names the
choice and points at the ADR; the ADR carries the reasoning. Link it from the `adr` label on the
design's `Related` row.

**A ratified ADR is never rewritten.** That rule is the repo's, it binds this verb, and it holds in
the user's project as well as here: a revision lands as a **new, superseding ADR**, and the only
in-place edit ever made to an existing one is flipping its Status line to
`Superseded by <id> — <date>`. If this design changes a decision an existing ADR made, that is a
superseding ADR, not an edit.

Draft the ADR and show it before writing, the same as the design document. It ships `Accepted` with
a real date once the user confirms it — Codefall's stance is that what gets proposed gets built, so
a `Proposed` ADR nobody will revisit is a state with no exit.

## Status and lifecycle

Three states, one word plus a date.

| Status | Meaning | Lives in |
| --- | --- | --- |
| `Draft — <date>` | Being written. The user stopped and is coming back | `docs/designs/` |
| `Ready — <date>` | Written and agreed, beads created. The normal end of a session | `docs/designs/` |
| `Archived — <date>` | Superseded or dropped | `docs/designs/archive/` |

**The design's status describes the document, never the work.** Whether the work is queued,
underway, or done is Beads' to say, and it says it better than a status line can. This is `specify`'s
rule, taken for `specify`'s reason: it diverges from `conceptualize`, which carries an `Active` state
precisely because a concept has no tracker representation to carry it. A design has one — the beads
it created.

So there is no `Active` here and nothing for `implement` to transition. `bd list` and `bd ready`
answer the question a status line would only approximate.

**`Ready` is the normal end of a session, not `Draft`.** A design the user worked through and agreed
with is finished, whatever they called it when they started. Set `Draft` only when they said they
are stopping and will come back.

`Archived` gains a `**Replaced by:** DESIGN-NNN — <date>` line when something took its place, and
omits it when the design was simply dropped. Archiving moves the file to `docs/designs/archive/`
under the same name; the identifier stays valid and citations still resolve, to the new path.

**A design that no longer describes the code is revised or archived, not labelled.** The revision
path reconciles the graph as well, which a status word cannot do — see
[When a design changes](#when-a-design-changes-after-its-beads-exist).

Status transitions are this skill's to make. Report the state and offer; never transition a design
on your own initiative.

## The designs directory

`design` maintains `docs/designs/AGENTS.md`. It is written when the directory is created and added
on a later run if it is missing. A repo that has never run this skill has no `docs/designs/` to
scope, so that is the whole retrofit story.

**Never overwrite a file that has drifted.** When one exists and its content differs from what this
skill ships, show the user the difference and ask. The difference is a hand edit until they say
otherwise, and a repair that silently discards someone's rule is worse than a stale file. Replace it
only on a yes; on a no, leave it and say nothing further about it.

The file's content:

```markdown
# docs/designs — operative rules

- `archive/` is history. Do not read it unless the user asks about a superseded design by name.
- A design's identifier is permanent. `DESIGN-007` means the same document after it is archived.
- Local task identifiers (`T1`, `T2`) are permanent within a design and append-only. A retired one
  is never reused.
- Once a Task Plan has collapsed to an ID mapping, Beads is authoritative for the tasks. Never
  restore a task table over it.
- ADRs live in `docs/adrs/` and are never rewritten. A revision is a new, superseding ADR.
- Status describes the document, never the work. Beads holds work state.
- Status transitions are `/design`'s to make, never a hand edit.
```

## Beads

Beads holds the tasks and the graph. The design document holds the approach.

### What gets created

| Bead | One per | Type | Body |
| --- | --- | --- | --- |
| Epic | design document | `epic` | The Overview, and the path to the document |
| Task | Task Plan row | `task`, or `bug` where it is one | What to do, the design ref, and acceptance criteria |

**Every task bead carries acceptance criteria** — two to five checkable statements, drawn from the
spec requirements the task serves and the Hard Constraints that bind it. A criterion that traces to
a spec requirement cites it (`R3: context survives a restart`). They are the definition of done
`implement` verifies before closing the bead, so write them as checks, not as restated task text.

Tier 0 has no epic. One or two beads are created directly, with the reproduction, the cause, and
acceptance criteria in the body — there is no document for them to point at, so the bead must be
self-sufficient.

Every bead gets `--spec-id` set to the design document's path. It is a backstop: the mapping line in
the document is the primary record, and `spec_id` is what finds the beads again if that line is
lost.

### Creating the graph

`bd create --graph` takes a plan file and creates every node and edge in one call, then prints the
mapping the document needs. The plan's node keys **are** the Task Plan's local IDs, which is what
makes the staging table worth writing.

```json
{
  "nodes": [
    { "key": "EPIC", "type": "epic",
      "title": "DESIGN-007: Stage context handoff",
      "description": "<the Overview, and docs/designs/DESIGN-007-stage-context.md>" },
    { "key": "T1", "type": "task", "parent_key": "EPIC",
      "title": "Add StageContext type + serde",
      "description": "<what to do> · Design ref: DESIGN-007 § Components and Interfaces" },
    { "key": "T2", "type": "task", "parent_key": "EPIC",
      "title": "Wire context load into /scaffold",
      "description": "<what to do> · Design ref: DESIGN-007 § Architecture" }
  ],
  "edges": [
    { "from_key": "T2", "to_key": "T1", "type": "blocks" }
  ]
}
```

```bash
bd create --graph <plan.json> --dry-run    # validates the graph, creates nothing
bd create --graph <plan.json>
```

```
Created 3 issues
  EPIC -> bd-a2g
  T1 -> bd-unz
  T2 -> bd-s58
```

**The edge direction is the trap.** Despite the type name, `from_key` is the **dependent** and
`to_key` is the **blocker**: `{"from_key": "T2", "to_key": "T1", "type": "blocks"}` means T2 is
blocked by T1. It maps straight off the staging table — the row's own ID is `from_key`, and each
entry in its **Depends on** column is a `to_key`. Getting it backwards produces a graph that runs in
reverse, and nothing but the ready set will tell you.

**The plan file carries only these fields.** `key`, `title`, `type`, `description`, `labels`,
`priority`, `parent_key` on a node; `from_key`, `to_key`, `type` on an edge. Anything else is
**silently dropped** with a warning. Neither `--spec-id` nor `--acceptance` is among them, so both
are set afterwards, in one update pass over the creation output:

```bash
bd update <id> --spec-id docs/designs/DESIGN-007-stage-context.md \
  --acceptance $'R3: context survives a restart\nHard constraint: one open write txn per booking'
```

### Verify the graph

Creation is not verification, and a reversed edge is invisible in the creation output.

```bash
bd ready          # the unblocked set
bd dep cycles     # must find none
```

**The ready set must be exactly the rows whose Depends on column is `—`**, plus the epic. If a task
with prerequisites is ready, or a root task is not, the edges went in backwards — fix them before
collapsing the table, while the local IDs still line up with what you sent.

## When a design changes after its beads exist

The mapping line makes this tractable: it says which bead each local ID became, so a later run
updates the graph rather than duplicating it.

**Read the current beads first.** `bd show <id> --json` for each bead in the mapping. Beads holds
the content that was created, which is why the collapsed Task Plan does not need a per-row hash —
the comparison is against the graph itself, not against a record of what the document used to say.

Then, row by row:

| Case | What happens |
| --- | --- |
| The row changed | [Edited or replaced](#when-a-task-row-changes), decided by whether anyone is holding the bead |
| The row is new | A new local ID, a new bead, appended to the mapping line |
| An edge changed | `bd dep add` or `bd dep remove`, then re-verify with `bd ready` |
| The row is gone | Its local ID retires. **Report the bead and let the user choose** — close it, or leave it open because work already happened against it |

**A removed row is never closed silently.** Somebody may be holding that ticket, and a design edit
is not evidence that its work stopped mattering. This is the same rule `specify` applies when a
requirement leaves a spec, for the same reason.

### When a task row changes

Read the bead first — `bd show <id> --json` reports its status, assignee, and comment count.

**Untouched** — `open`, unassigned, no comments — is edited, whatever changed. Nobody has acted on
the old wording, so there is nobody to mislead and the size of the change does not matter.

**Touched** — claimed, in progress, commented on, or closed — asks one question: would the work done
against the old wording still be correct and sufficient under the new wording? Yes, edit it. No,
replace it: create the new bead, `bd supersede <old> --with <new>`, and repoint the mapping line.

A ticket must not change under someone holding it. Where nobody is holding it, editing costs
nothing.

Worked, so two sessions apply it the same way:

| Old row | New row | Bead state | Result |
| --- | --- | --- | --- |
| Wire context load into `/scaffold` | same, with a Design ref added | any | **edit** — nothing about the work changed |
| Add `StageContext` type | Add `StageContext` type + serde | open, unclaimed | **edit** — nobody started; the bead becomes the bigger task |
| Add `StageContext` type | Add `StageContext` type + serde | closed | **replace** — the type exists, the serde does not, and editing a closed bead leaves it unbuilt |
| Emit context on `/conceptualize` exit | Emit context on every verb's exit | claimed | **replace** — they are building one thing and would silently owe five |

`bd supersede <old> --with <new>` closes the old bead with a reference to its replacement, so the
record of what the task used to be survives. **The local ID does not change** — `T2` still means
`T2`; only the bead behind it moves, and the mapping line records the new one.

**A row that has become two separate tasks is a removal plus two additions**, not a replacement, and
the removal goes through the report-don't-close rule above.

Say per row which you did and why, in the report.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

### 1. Check preconditions

Run the shared check against the user's project — the script's path resolves from this skill's
directory, and the argument is the project. It reports what is set up and repairs nothing.

```bash
"../../shared/preflight.sh" .
```

`beads=ok` advances to step 2. Otherwise read `beads_reason`, tell the user what is missing, hand
over the command that fixes it, and **stop**:

| `beads_reason` | What is wrong | Give them |
| --- | --- | --- |
| `not_installed` | `bd` is not on PATH | `brew install beads` |
| `not_initialized` | this repository has no beads database | `bd init` |
| `unreadable` | bd found a database and could not read it | quote `beads_detail` — bd's own words are more use than a paraphrase |

**Never run the remedy.** `bd init` writes `.beads/`, git hooks, `.claude/settings.json`, and a
block in `AGENTS.md` and `CLAUDE.md`, then commits all of it. That is the user's decision, not a
repair a skill makes on its way to somewhere else.

Beads is a hard gate here, unlike the design document: the graph is this skill's output, and a
design document with no tasks behind it hands `implement` nothing.

### 2. Take the input

One open question, unless the invocation already answered it:

> "What are we designing? A spec identifier, a concept, or just tell me what needs building."

**Then look for a spec.** Read `docs/specs/` — not `archive/` — and offer the relevant one:

> SPEC-004 covers booking history and looks like what you are describing. Design against it?

A spec is not required. It carries the *what*, so taking one means this skill never has to
re-derive it, and a design written without one is designing against a target nobody wrote down. Say
that once if there is no spec and the work is more than a fix, and offer `/specify`. On no,
continue — this is a requirement the user can break.

**A spec that is not ready is a stop.** Two gates, both declared by `specify` and its tracker
profile:

- The spec document says `Status: Draft`. It is still being written; designing against it wastes the
  design.
- Its tracker issues carry `requires-mockup`. A requirement with an unspecified visual surface is
  not designable, which is why the label exists.

Read the status from the document. Read the labels from the tracker per
`../specify/trackers/<name>/PROFILE.md`. If the tracker is unreachable, say so and ask the user
whether the mockups exist rather than guessing.

Name the gate, say what clears it, and stop. Do not design half of a spec around a blocked
requirement.

**Then read what frames it.** The spec's concept, if it names one. `docs/concepts/` if no spec
framed the work. Both carry constraints — a concept's **Environment & constraints** section is
written for exactly this moment.

### 3. Read the docs and the code

This is the step that makes the design worth anything, and skipping it is how a design proposes
something the project already decided against.

**The project's stance** — `docs/adrs/`, every `AGENTS.md` in the tree, and `docs/decision-log.md`.
The layering, the component boundaries, the DI approach, and how boundaries are enforced are already
decided. Design within them.

**The existing designs** — `docs/designs/`, not `archive/`. If one already covers this, say so and
link it before going further; the user may want to revise that one.

**The code** — the components this touches, their facades, and what already exists that this can
use. Read the artifacts, not their names.

**The graph** — `bd list` and `bd search` for work that already exists against this area. A task
this design would create that is already a bead is a dependency edge, not a new task.

Report what you found before designing. If the work already exists, in code or in the graph, say so
and stop rather than designing around it.

### 4. Decide the tier, and whether there is an ADR

Both judgements, stated together, before any writing:

> This crosses the context store and the scaffold command, and comes to four tasks, so I would write
> a design document. The choice between a signed payload and a session lookup is hard to reverse and
> other components will build on it, so that is an ADR as well.

Walk the [tier table](#scale-the-artifact-to-the-work) and the [ADR trigger](#adrs) explicitly. They
are independent: tier 0 work can produce an ADR, and a tier 2 design usually produces none.

**At tier 0, skip to step 7.** There is no document to write.

### 5. Design it

The work itself. Settle, in this order, and only what applies:

1. **The approach** — how this gets built, in a paragraph, and the decisions inside it.
2. **The components** — what changes, what is new, what talks to what.
3. **The contracts** — the types, APIs, and messages that cross a boundary.
4. **What persists** — the data model, and any migration it implies.
5. **What fails** — the new failure modes, and what the system does about each.
6. **What is left out** — the approaches considered and set aside.

**Research inline as you go, and attach every finding to the decision it bears on.** A library's
behavior, a protocol's constraint, a benchmark — it goes in Overview or Architecture next to the
choice it informed, and not into a file of its own.

**"Like $LIBRARY does it."** When the user references another project or library, offer once to look
it up. On yes, research it and summarize only what changes a decision here. Confirm the summary with
the user before it reaches the document. On no, move on without searching.

**Raise a concern once, then defer.** When the approach the user wants has a problem they may not
have weighed, name it, say why, and let them decide. Cap at two rounds. If it stays unresolved it
goes into the document as a stated risk, and you move on. The user knows the system and you may be
wrong.

**Do not bikeshed.** Naming, and which of two equivalent shapes is better, do not change what gets
built. When a question is interesting but changes nothing, do not ask it.

### 6. Stage the tasks

Cut the design into tasks and wire the edges.

**Tasks decompose by what can be built and verified on its own**, against the design. That is a
different cut from `specify`'s: requirements decompose by what a consumer can observe, and one
requirement routinely becomes several tasks.

The sizing question is whether one person could pick the task up, finish it, and have something
that either works or does not. A task nobody can tell is done is too big or too vague.

**Then wire the edges, and check the ordering by reading it backwards**: for each task, what must
exist before it can start? A task with no answer is a root. A cycle is an error in the cut — fix the
cut, not the edges.

Write the staging table. This is the artifact review is for, and it is why the local IDs exist.

### 7. Draft and confirm

Compose the full document and **show it before anything is written**. Nothing lands unapproved.

Say which conditional sections you left out and why — "no Data Models section, because nothing here
persists" — so the user can catch an omission that was actually a gap.

Show the ADR too, if there is one, and say plainly that it ships `Accepted`.

Then set the status: `Ready`, unless they said they are stopping and coming back, which is `Draft`.

**At tier 0, this is the confirmation instead**: the beads you would create, their titles, their
bodies, and their edges. The user approves the graph, not a document.

### 8. Write the document

Write `docs/designs/DESIGN-NNN-slug.md` with the Task Plan in its **staged** form, the ADR if there
is one, and `docs/designs/AGENTS.md` if it was missing.

Write the document before creating the beads. If bead creation fails halfway, the design still
exists on disk and the run is resumable; the other way round leaves a graph nothing explains.

Do not commit.

### 9. Create the graph

Build the plan file from the staging table, per [Beads](#beads). Dry-run it, create it, set
`--spec-id` and `--acceptance` on every task bead, then verify with `bd ready` and `bd dep cycles`.

If the ready set does not match the roots of the staging table, fix the edges now, before the table
collapses.

### 10. Collapse the Task Plan

Rewrite the Task Plan section in place with the mapping line and today's date. **Rewrite, not
append** — the staging table comes out. Beads is authoritative from here.

At tier 0 there is nothing to collapse.

### 11. Link back and report

Fill in the `plan:` field on the framing concept's `Related` line with this design's identifier.
That line is written expecting this. Where a spec framed the work, the concept is the one named in
the spec's `**Concept:**` row — one hop, because the design does not record it twice. Add the
identifier and change nothing else in the file.

There is no back-link to write into the spec: the design's `spec` label carries the connection, and
`grep SPEC-004` finds the design, its tasks' references, and the spec itself in one pass.

Report:

- the design's path, identifier, and status, or that this was tier 0 and why;
- any ADR written, and what it decided;
- every bead created, with its local ID and its title;
- the ready set — which tasks `implement` can start on today;
- anything left unresolved, and any concern the user overruled.

Do not commit. Do not start implementing.

## Other modes

Invoking this skill on an existing design does one of four things. Ask which if it is not obvious.

- **Promote** `Draft` to `Ready`, or **reopen** `Ready` to `Draft` when no beads exist yet.
- **Revise** a design and reconcile its graph, per
  [When a design changes](#when-a-design-changes-after-its-beads-exist). A design the code has moved
  past is revised, not labelled — there is no status that says so.
- **Archive** a design: set `Status: Archived`, add `**Replaced by:**` if something took its place,
  move the file to `docs/designs/archive/`, and report its open beads to the user rather than
  closing them.
- **Add tasks** to an existing design — new local IDs appended to the table, new beads appended to
  the mapping. Retired local IDs stay retired.

Every one of these is the user's decision. Report the state and offer; never transition a design on
your own initiative.

## Lineage

What this took from elsewhere, and what it deliberately did not, so nobody re-adds it.

**From Kiro**, the primary model: the design document's section set, and the principle that the
design document is architecture only, with tasks living somewhere else. Kiro requires six sections —
Overview, Architecture, Components and Interfaces, Data Models, Error Handling, Testing Strategy —
and three of those are conditional here, because a required section with nothing behind it teaches
readers to skim. Also from Kiro: research inline rather than in sibling files, and bug fixes routed
away from the full artifact, which is tier 0.

**From spec-kit**: the Technical Context block, which forces concreteness where prose hand-waves;
the shape of the Alternatives table; and numbered-identifier traceability, which here is the local
task IDs and the Design ref column.

**Dropped from spec-kit**: the constitution check, the phase gate and its re-check ceremony, the
Project Structure section, and the `research.md` / `data-model.md` / `quickstart.md` / `contracts/`
file sprawl. Codefall's stance is that what gets proposed gets built, so the gating and negotiation
loops have nothing to gate — what is kept is the structure that keeps documents scannable and
machine-parseable.

## Rules

- **Nothing is written without the user confirming it first** — the document, the ADR, and at tier 0
  the beads.
- **Scale the artifact to the work.** Tier 0 is a real outcome, not a failure to write a document.
- **Never fill a heading.** A conditional section with nothing behind it is deleted, heading and all.
- **The design says how, never what.** What a consumer observes belongs to `specify`, and a design
  that restates it is duplicating a document that will change without it.
- **Design within the project's ADRs.** Changing the stance is a superseding ADR, said out loud,
  never a quiet exception.
- **A ratified ADR is never rewritten.** A revision is a new, superseding ADR; the only in-place
  edit is the Status line.
- **Identifiers are append-only** — design numbers, and local task IDs within a design. A retired
  one is never reused.
- **Beads is authoritative once the tasks exist.** The Task Plan collapses to a mapping and never
  grows a duplicate table.
- **Every task bead carries acceptance criteria** — checkable, citing spec requirement IDs where
  they trace. They are what `implement` verifies before closing the bead.
- **Verify the graph before collapsing the table.** `bd ready` and `bd dep cycles`, against the
  staging table's roots.
- **A removed task's bead is reported, never closed silently.** Work may already have happened
  against it.
- **A ticket must not change under someone holding it.** An untouched bead is edited whatever
  changed; a bead someone is holding is replaced when the work already done would no longer count.
- **Status describes the document, never the work.** Work state belongs to Beads.
- **Never record a hop you can derive.** A design carries its spec, or its concept when there is no
  spec — not both.
- **Research goes inline**, attached to the decision it informed. No sibling research files.
- **Push back once, then defer** — on the tier, on the approach, on the cut. The user knows the
  system and you may be wrong.
- **Never overwrite a file that has drifted.** Show the difference and ask.
