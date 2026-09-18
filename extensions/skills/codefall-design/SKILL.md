---
name: codefall-design
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

Decide how a feature gets built, and put the work into the dependency graph so `codefall-implement`
can pick it up.

Two outputs. A **design document** at `docs/designs/DESIGN-NNN-slug.md` — the approach, the
architecture, and a staged task plan — when the work earns one. **The tasks in Beads**, with their
dependency edges, always.

Designing is not implementing. Once the document is written and the graph exists, stop.

Paths that start with `../` are relative to this skill's directory, not the user's project. A path
through `../../../.codefall/` is the one that leaves the skills directory: it names a file `codefall
init` installed in the project's own `.codefall/`.

## Files beside this one

Read each when its step says to; none is loaded up front.

- `reference/document.md` — the design document's shape: header, sections and their triggers, the
  Technical Context and Hard Constraints blocks, both forms of the Task Plan. Read before step 5.
- `reference/beads.md` — what gets created in Beads, the plan file, the edge direction, and how to
  verify the graph. Read before step 9.
- `reference/revising.md` — reconciling the graph when a design changes after its beads exist. Read
  for the Revise mode.
- `templates/designs/AGENTS.md` — the operative rules this skill installs at `docs/designs/AGENTS.md`.

## Scope — how, not what and not whether

| In scope | Out of scope | Whose |
| --- | --- | --- |
| The approach, and the decisions inside it | Whether this is worth building | `codefall-conceptualize` |
| Components, their relationships, and data flow | What a consumer observes when it works | `codefall-specify` |
| Contracts, types, schemas, storage, failure modes | What the screen looks like | `codefall-mock-up` |
| Which existing code changes and which is new | The project's architecture stance | `codefall-scaffold` |
| The task breakdown and its dependency edges | Writing the code | `codefall-implement` |
| Hard-to-reverse choices, recorded as ADRs | Estimates and assignment | the team |

**A design does not restate the specification.** What a consumer observes is `codefall-specify`'s;
cite it and move on.

**The project's stance is already decided** — layering, component boundaries, and how they are
enforced — in `docs/adrs/` and the scoped `AGENTS.md` files. Design within it. A design that needs
the stance changed says so once, then either follows the ADR or writes a superseding one. Never a
quiet exception.

## Scale the artifact to the work

A small required core, everything else conditional, and **no empty or placeholder sections — omit
them.**

Three tiers. The work picks the tier; the user can overrule it.

| Tier | When | Output |
| --- | --- | --- |
| **0 — no document** | All three hold: the change is contained to one component, no public interface changes, and the task graph is one or two beads | Beads only |
| **1 — minimal document** | Anything that crosses a component boundary, or fans out past roughly three dependent tasks | The three required sections |
| **2 — full document** | The same, plus any conditional section whose trigger fires | Required three plus what triggered |

- **Simple bug fixes land at tier 0.** Create the bead directly with the reproduction and the cause.
- **Tier 1 and tier 2 are not decided separately.** Write the required three, then walk the
  conditional triggers; whichever fire, fire.
- **An ADR is not gated on the tier.** A one-component fix can produce an ADR and no document.
- **Say the tier out loud, with the reason, before writing anything.** The judgement is yours to
  make and the user's to overrule:

> This touches one component, changes nothing public, and comes to two tasks. I would skip the
> design document and create the beads directly. Want the document anyway?

Push back once if you disagree with their answer, then do what they ask.

## The document

`docs/designs/DESIGN-NNN-slug.md`, git-tracked. Three digits, zero-padded, the highest existing
number plus one. **A design is never renumbered and its identifier is never reused**, including
after it is archived — the file moves, the identifier does not.

Three required sections — **Overview**, **Architecture**, **Task Plan** — and seven conditional ones,
each with a trigger. Research findings go inline, in Overview or Architecture next to the decision
they informed; there is no sibling `research.md`, `data-model.md`, `quickstart.md`, or `contracts/`.
The header, the section tables, the Technical Context and Hard Constraints blocks, and both forms of
the Task Plan are in `reference/document.md`.

## ADRs

A separate artifact, `docs/adrs/ADR-NNN-title.md`, in the Nygard shape — Status, Context, Decision,
Consequences, Related — written from `../codefall-scaffold/templates/adrs/_TEMPLATE.md`.

- **The number continues the project's own sequence:** the highest bare `ADR-NNN` in `docs/adrs/`
  plus one, starting at `ADR-001`. The prefixed sequences — `ADR-BASE-NN`, `ADR-<PREFIX>-NN` — are
  inherited stance and are never continued here.
- **The trigger** is the design's Alternatives Considered holding a choice that is hard to reverse
  or that other components will build on: a new dependency; a schema or protocol decision other
  components will be written against; a rejected alternative that cost real analysis. Most designs
  need none.
- **One home for the rationale.** The design names the choice and points at the ADR from the `adr`
  label on `Related`; the ADR carries the reasoning.
- **A ratified ADR is never rewritten.** A revision is a new, superseding ADR; the only in-place
  edit is the Status line, to `Superseded by <id> — <date>`. A design that changes a decision an
  existing ADR made writes a superseding ADR.
- Draft the ADR and show it before writing. It ships `Accepted` with a real date once the user
  confirms it.

## Status and lifecycle

Three states, one word plus a date.

| Status | Meaning | Lives in |
| --- | --- | --- |
| `Draft — <date>` | Being written. The user stopped and is coming back | `docs/designs/` |
| `Ready — <date>` | Written and agreed, beads created. The normal end of a session | `docs/designs/` |
| `Archived — <date>` | Superseded or dropped | `docs/designs/archive/` |

- **Status describes the document, never the work.** Work state is Beads' — `bd list` and
  `bd ready`. There is no `Active` and nothing for `codefall-implement` to transition.
- **`Ready` is the normal end of a session.** Set `Draft` only when the user said they are stopping
  and will come back.
- `Archived` gains a `**Replaced by:** DESIGN-NNN — <date>` line when something took its place.
  Archiving moves the file to `docs/designs/archive/` under the same name; citations still resolve.
- A design that no longer describes the code is revised or archived, not labelled. Revising
  reconciles the graph — `reference/revising.md`.
- Transitions are this skill's to make. Report the state and offer; never transition a design on
  your own initiative.

## The designs directory

`codefall-design` maintains `docs/designs/AGENTS.md` from `templates/designs/AGENTS.md`: written when
the directory is created, added on a later run if it is missing.

**Never overwrite a file that has drifted.** When one exists and differs from the template, show the
difference and ask. Replace it only on a yes; on a no, leave it and say nothing further about it.

## Beads

Beads holds the tasks and the graph; the document holds the approach. One epic per design document
and one task bead per Task Plan row, every task bead carrying two to five checkable acceptance
criteria that cite the spec requirements they trace to. Tier 0 has no epic: one or two
self-sufficient beads. Every bead gets `--spec-id` set to the document's path. The plan file, the
edge direction, the field list, and the verification are in `reference/beads.md`.

## Project customizations

Follow `../../../.codefall/shared/customizations.md` for this verb.

## Process

### 1. Check preconditions

Run the shared check against the user's project. It reports what is set up and repairs nothing.

```bash
"../../../.codefall/shared/preflight.sh" .
```

`beads=ok` advances to step 2. Otherwise read `beads_reason`, tell the user what is missing, hand
over the command that fixes it, and **stop**:

| `beads_reason` | What is wrong | Give them |
| --- | --- | --- |
| `not_installed` | `bd` is not on PATH | `brew install beads` |
| `not_initialized` | this repository has no beads database | `bd init` |
| `unreadable` | bd found a database and could not read it | quote `beads_detail` |

Then read the checkout lines. `behind` above `0` or `refresh=stale` means the environment may not
match `main`: say so and offer `/codefall-refresh` before continuing. `refresh=undeclared` names
`/codefall-equip` instead. Never pull the checkout or run the local commands from here.

**Never run the remedy.** `bd init` writes `.beads/`, git hooks, `.claude/settings.json`, and a
block in `AGENTS.md` and `CLAUDE.md`, then commits all of it. That is the user's decision.

Beads is a hard gate: the graph is this skill's output.

### 2. Take the input

One open question, unless the invocation already answered it:

> "What are we designing? A spec identifier, a concept, or just tell me what needs building."

**Then look for a spec.** Read `docs/specs/` — not `archive/` — and offer the relevant one:

> SPEC-004 covers booking history and looks like what you are describing. Design against it?

A spec is not required. If there is none and the work is more than a fix, say once that the design
has no written target and offer `/specify`. On no, continue.

**A spec that is not ready is a stop.** Two gates:

- The spec document says `Status: Draft`.
- Its tracker issues carry `requires-mockup`. Read the labels from the tracker per
  `../codefall-specify/trackers/<name>/PROFILE.md`; if the tracker is unreachable, say so and ask
  whether the mockups exist rather than guessing.

Name the gate, say what clears it, and stop. Do not design half of a spec around a blocked
requirement.

**Then read what frames it.** The spec's concept, if it names one. `docs/concepts/` if no spec
framed the work. A concept's **Environment & constraints** section is written for this moment.

### 3. Read the docs and the code

- **The project's stance** — `docs/adrs/`, every `AGENTS.md` in the tree, and
  `docs/decision-log.md`. Design within them.
- **The existing designs** — `docs/designs/`, not `archive/`. If one already covers this, say so
  and link it; the user may want to revise that one.
- **The code** — the components this touches, their facades, and what already exists that this can
  use. Read the artifacts, not their names.
- **The graph** — `bd list` and `bd search` for work that already exists against this area. A task
  this design would create that is already a bead is a dependency edge, not a new task.

Report what you found before designing. If the work already exists, in code or in the graph, say so
and stop.

### 4. Decide the tier, and whether there is an ADR

Both judgements, stated together, before any writing:

> This crosses the context store and the scaffold command, and comes to four tasks, so I would write
> a design document. The choice between a signed payload and a session lookup is hard to reverse and
> other components will build on it, so that is an ADR as well.

Walk the [tier table](#scale-the-artifact-to-the-work) and the [ADR trigger](#adrs) explicitly. They
are independent.

**At tier 0, skip to step 7.**

### 5. Design it

Read `reference/document.md`. Settle, in this order, and only what applies:

1. **The approach** — how this gets built, in a paragraph, and the decisions inside it.
2. **The components** — what changes, what is new, what talks to what.
3. **The contracts** — the types, APIs, and messages that cross a boundary.
4. **What persists** — the data model, and any migration it implies.
5. **What fails** — the new failure modes, and what the system does about each.
6. **What is left out** — the approaches considered and set aside.

**Research inline as you go**, and attach every finding to the decision it bears on, in Overview or
Architecture.

**"Like $LIBRARY does it."** When the user references another project or library, offer once to look
it up. On yes, research it and summarize only what changes a decision here; confirm the summary
before it reaches the document. On no, move on without searching.

**Raise a concern once, then defer.** Name it, say why, and let them decide. Cap at two rounds; if
it stays unresolved it goes into the document as a stated risk.

**Do not bikeshed.** Naming, and which of two equivalent shapes is better, do not change what gets
built. When a question changes nothing, do not ask it.

### 6. Stage the tasks

**Tasks decompose by what can be built and verified on its own**, against the design — a different
cut from `codefall-specify`'s, where requirements decompose by what a consumer observes, and one
requirement routinely becomes several tasks.

The sizing test: could one person pick the task up, finish it, and have something that either works
or does not? A task nobody can tell is done is too big or too vague.

**Wire the edges, and check the ordering by reading it backwards**: for each task, what must exist
before it can start? A task with no answer is a root. A cycle is an error in the cut — fix the cut,
not the edges.

**A task that introduces a tool carries the script change as a criterion.** When a task's
predicted files include a compose file, a migrations directory, a lockfile, a codegen config, or an
`.env.example`, one of its acceptance criteria says the project's declared `start` and `update`
scripts were changed for it, following `codefall-equip`. The environment a teammate refreshes into
is part of what the task delivers. A project with no `local` block gets the criterion "equip the
project" on the first such task instead.

Write the staging table.

### 7. Draft and confirm

Compose the full document and **show it before anything is written**. Nothing lands unapproved.

Say which conditional sections you left out and why — "no Data Models section, because nothing here
persists" — so the user can catch an omission that was a gap.

Show the ADR too, if there is one, and say plainly that it ships `Accepted`.

Then set the status: `Ready`, unless they said they are stopping and coming back, which is `Draft`.

**At tier 0, this is the confirmation instead**: the beads you would create, their titles, their
bodies, and their edges. The user approves the graph, not a document.

### 8. Write the document

Write `docs/designs/DESIGN-NNN-slug.md` with the Task Plan in its **staged** form, the ADR if there
is one, and `docs/designs/AGENTS.md` if it was missing.

Write the document before creating the beads, so a failed creation leaves a resumable run rather
than a graph nothing explains.

Do not commit.

### 9. Create the graph

Read `reference/beads.md`. Build the plan file from the staging table, dry-run it, create it, set
`--spec-id` and `--acceptance` on every task bead, then verify with `bd ready` and `bd dep cycles`.

If the ready set does not match the roots of the staging table, fix the edges now, before the table
collapses.

### 10. Collapse the Task Plan

Rewrite the Task Plan section in place with the mapping line and today's date. **Rewrite, not
append** — the staging table comes out. Beads is authoritative from here.

At tier 0 there is nothing to collapse.

### 11. Link back and report

Fill in the `plan:` field on the framing concept's `Related` line with this design's identifier.
Where a spec framed the work, the concept is the one named in the spec's `**Concept:**` row. Add the
identifier and change nothing else in the file.

There is no back-link to write into the spec: the design's `spec` label carries the connection.

Report:

- the design's path, identifier, and status, or that this was tier 0 and why;
- any ADR written, and what it decided;
- every bead created, with its local ID and its title;
- the ready set — which tasks `codefall-implement` can start on today;
- anything left unresolved, and any concern the user overruled.

Do not commit. Do not start implementing.

## Other modes

Invoking this skill on an existing design does one of four things. Ask which if it is not obvious.

- **Promote** `Draft` to `Ready`, or **reopen** `Ready` to `Draft` when no beads exist yet.
- **Revise** a design and reconcile its graph, per `reference/revising.md`. A design the code has
  moved past is revised, not labelled.
- **Archive** a design: set `Status: Archived`, add `**Replaced by:**` if something took its place,
  move the file to `docs/designs/archive/`, and report its open beads to the user rather than
  closing them.
- **Add tasks** to an existing design — new local IDs appended to the table, new beads appended to
  the mapping. Retired local IDs stay retired.

Every one of these is the user's decision. Report the state and offer; never transition a design on
your own initiative.

## Rules

- **Nothing is written without the user confirming it first** — the document, the ADR, and at tier 0
  the beads.
- **Scale the artifact to the work.** Tier 0 is a real outcome, not a failure to write a document.
- **Never fill a heading.** A conditional section with nothing behind it is deleted, heading and all.
- **The design says how, never what.** What a consumer observes belongs to `codefall-specify`, and a design
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
  they trace. They are what `codefall-implement` verifies before closing the bead.
- **A task that introduces infrastructure, a dependency, a migration, or generated code names the
  local-script change in its criteria.** The scripts stay current at the point of introduction,
  never as a follow-up.
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
