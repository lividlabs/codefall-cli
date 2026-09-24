---
name: codefall-design
description: Decide how a feature gets built and put the work into the graph — read the docs and the code, judge whether the change warrants a design document at all, write one scaled to the work at docs/designs/, record hard-to-reverse choices as ADRs, decide which tasks are verified through the wired product and draft their test case and its criteria into the bead, and create the task graph in Beads from the document's staged task plan.
argument-hint: "[the spec, the vision, or what you want built]"
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
  Technical Context and Hard Constraints blocks, the Task Plan and its callout. Read before step 5.
- `reference/beads.md` — what gets created in Beads, the test case a bead's criteria name, the plan
  file, the edge direction, and how to verify the graph. Read before step 9.
- `reference/revising.md` — reconciling the graph when a design changes after its beads exist. Read
  for the Revise mode.
- `templates/designs/AGENTS.md` — the operative rules this skill installs at `docs/designs/AGENTS.md`.

## Scope — how, not what and not whether

| In scope | Out of scope | Whose |
| --- | --- | --- |
| The approach, and the decisions inside it | Whether this is worth building | `codefall-envision` |
| Components, their relationships, and data flow | What a consumer observes when it works | `codefall-specify` |
| Contracts, types, schemas, storage, failure modes | What the screen looks like | `codefall-mock-up` |
| Which existing code changes and which is new | The project's architecture stance | `codefall-scaffold` |
| The task breakdown and its dependency edges | Writing the code | `codefall-implement` |
| Which tasks need a test case, and its criteria | Writing the case file, and running it | `codefall-implement`, `codefall-test` |
| Hard-to-reverse choices, recorded as ADRs | Estimates and assignment | the team |

**A design does not restate the specification.** What a consumer observes is `codefall-specify`'s;
cite it and move on.

**The project's stance is already decided** — layering, component boundaries, and how they are
enforced — in `docs/adrs/` and the scoped `AGENTS.md` files. Design within it. A design that needs
the stance changed says so once, then either follows the ADR or writes a superseding one.

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
- **Say the tier out loud, with the reason, before writing anything** — step 4 has the shape. The
  judgement is yours to make and the user's to overrule.

## The document

`docs/designs/DESIGN-NNN-slug.md`, git-tracked. Three digits, zero-padded, the highest existing
number plus one. **A design is never renumbered and its identifier is never reused**, including
after it is archived — the file moves, the identifier does not.

Three required sections — **Overview**, **Architecture**, **Task Plan** — and seven conditional ones,
each with a trigger. The header, the section tables, the Technical Context and Hard Constraints
blocks, and the Task Plan's callout are in `reference/document.md`.

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
  edit is the Status line, to `Superseded by <id> — <date>`.
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
- `Archived` moves the file to `docs/designs/archive/` under the same name; citations still resolve.
- A design that no longer describes the code is revised or archived, not labelled. Revising
  reconciles the graph — `reference/revising.md`.

## The designs directory

`codefall-design` maintains `docs/designs/AGENTS.md` from `templates/designs/AGENTS.md`: written when
the directory is created, added on a later run if it is missing.

**Never overwrite a file that has drifted.** One that exists and differs from the template: show the
difference and ask. Replace only on a yes; on a no, leave it and say nothing further about it.

## Beads

Beads holds the tasks and the graph; the document holds the approach. One epic, `<prefix>-DESIGN-NNN`,
and one task bead per row, `<prefix>-DESIGN-NNN-Tn`. Tier 0 has no epic: one or two self-sufficient
beads. More in `reference/beads.md`.

## Project customizations

Follow `../../../.codefall/shared/customizations.md` for this verb.

## Process

### 1. Check preconditions

Run the shared check against the user's project. It reports what is set up and repairs nothing.

```bash
"../../../.codefall/shared/preflight.sh" .
```

`beads=ok` advances to step 2. Otherwise read `beads_reason`, tell the user what is missing, hand
over the command that fixes it, say to rerun this verb after it, and **stop**:

| `beads_reason` | What is wrong | Give them |
| --- | --- | --- |
| `not_installed` | `bd` is not on PATH | `brew install beads` |
| `not_initialized` | this repository has no beads database | `bd init` |
| `unreadable` | bd found a database and could not read it | quote `beads_detail` |

Then read the checkout lines. `behind` above `0` or `refresh=stale` means the environment may not
match `main`: say so and run `/codefall-refresh` before continuing. `refresh=undeclared` names
`/codefall-equip` instead. Never pull the checkout or run the local commands from here.

**Never run the remedy.** `bd init` writes `.beads/`, git hooks, and blocks in `AGENTS.md` and
`CLAUDE.md`, then commits; that is the user's decision.

Beads is a hard gate: the graph is this skill's output.

### 2. Take the input

One open question, unless the invocation already answered it:

> "What are we designing? A spec identifier, a vision, or just tell me what needs building."

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

**Then read what frames it.** The spec's vision, if it names one. `docs/visions/` if no spec
framed the work. A vision's **Environment & constraints** section is written for this moment.

### 3. Read the docs and the code

- **The project's stance** — `docs/adrs/`, every `AGENTS.md` in the tree, and
  `docs/decision-log.md`. Design within them.
- **The existing designs** — `docs/designs/`, not `archive/`. If one already covers this, say so
  and link it; the user may want to revise that one.
- **The code** — the components this touches, their facades, and what already exists that this can
  use. Read the artifacts, not their names.
- **The graph** — `bd dolt pull`, then `bd list` and `bd search` for existing work. A task
  this design would create that is already a bead is a dependency edge, not a new task.

Report what you found before designing. If the work already exists, say so and stop.

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

**Raise a concern once, then defer.** Name it, say why, and let them decide. Cap at two rounds;
unresolved, it goes into the document as a stated risk.

**Do not bikeshed.** Naming, and which of two equivalent shapes is better, do not change what gets
built; do not ask.

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

**A task verified through the wired product carries a test case.** Its acceptance criteria name the
case (`<area>/<slug>`), its modalities, and every criterion the case will hold — each cited in full
(`SPEC-003-REQ-01-AC-01`) or marked `derived` with the requirement it elaborates
(`SPEC-003-REQ-01`) and one line saying what it adds. `agentic` only where verifying an outcome
needs judgement; `spec` otherwise; neither where unit tests verify the task, and a driver being
available is no reason to add one. A gap the criteria expose in the spec goes in this run's report
for `codefall-specify`, never patched into the case. The form is in `reference/beads.md` and the
case-file format `codefall-implement` writes to is `../codefall-test/reference/case-file.md`.

Write the staging table.

### 7. Draft and confirm

Compose the full document and **show it before anything is written**. Nothing lands unapproved.

Say which conditional sections you left out and why — "no Data Models section, because nothing here
persists" — so the user can catch an omission that was a gap.

**Show the acceptance criteria of every task that carries a test case**, derived criteria included.
Approving the plan is the sign-off those criteria need; nothing later asks for it.

Show the ADR too, if there is one, and say plainly that it ships `Accepted`.

Then set the status: `Ready`, unless they said they are stopping and coming back, which is `Draft`.

**At tier 0, this is the confirmation instead**: the beads you would create, their titles, their
bodies, and their edges. The user approves the graph, not a document.

### 8. Branch, then write the document

Branch first, per `../../../.codefall/shared/landing.md` — `design/DESIGN-NNN-slug` — unless this
is a tier 0 run with no ADR, which writes no file.

Write `docs/designs/DESIGN-NNN-slug.md` with the Task Plan **staged**, the ADR if there is one,
and `docs/designs/AGENTS.md` if it was missing.

Write the document before creating the beads, so a failed creation leaves a resumable run rather
than a graph nothing explains.

### 9. Create the graph

Read `reference/beads.md`. Build the plan file from the table, dry-run it, create it, rename to the
IDs, set `--spec-id` and `--acceptance`, verify with `bd ready` and `bd dep cycles`, and push.

If the ready set does not match the table's roots, fix the edges now, before the callout is
rewritten.

### 10. Mark the Task Plan created

Rewrite the callout above the table per `reference/document.md`: the date, the epic's full ID, the
per-row form, and that Beads is authoritative. **The table stays.**

At tier 0 there is no document to mark.

### 11. Link back, commit, and report

Fill in the `plan:` field on the framing vision's `Related` line with this design's identifier.
Where a spec framed the work, the vision is the one named in the spec's `**Vision:**` row. Add the
identifier and change nothing else in the file.

There is no back-link to write into the spec: the design's `spec` label carries the connection.

Then land it per `../../../.codefall/shared/landing.md`, with `.beads/interactions.jsonl` when it
changed.

Report:

- the design's path, identifier, and status, or that this was tier 0 and why;
- any ADR written, and what it decided;
- every bead created, with its local ID, its title, and any test case its criteria name;
- any gap the case criteria exposed in the spec, for `codefall-specify`;
- the ready set — which tasks `codefall-implement` can start on today;
- anything left unresolved, and any concern the user overruled;
- the branch and the pull request;
- **last, what the user does next**: merge the pull request, then `/codefall-implement DESIGN-NNN`,
  or `/codefall-implement <bead>` at tier 0.

## Other modes

Invoking this skill on an existing design does one of four things. Ask which if it is not obvious.

- **Promote** `Draft` to `Ready`, or **reopen** `Ready` to `Draft` when no beads exist yet.
- **Revise** a design and reconcile its graph, per `reference/revising.md`. A design the code has
  moved past is revised, not labelled.
- **Archive** a design: set `Status: Archived`, add `**Replaced by:**` if something took its place,
  move the file to `docs/designs/archive/`, and report its open beads to the user rather than
  closing them.
- **Add tasks** to an existing design — new rows appended to the table, each with its bead, and
  the callout dated. Retired local IDs stay retired.

Every one of these is the user's decision. Report the state and offer; never transition a design on
your own initiative.

## Rules

- **Nothing is written without the user confirming it first** — the document, the ADR, and at tier 0
  the beads.
- **Scale the artifact to the work.** Tier 0 is a real outcome, not a failure to write a document.
- **Never fill a heading.** A conditional section with nothing behind it is deleted, heading and all.
- **The design says how, never what.** What a consumer observes belongs to `codefall-specify`; a
  design that restates it duplicates a document that will change without it.
- **Design within the project's ADRs.** Changing the stance is a superseding ADR, said out loud,
  never a quiet exception; a ratified ADR is never rewritten.
- **Identifiers are append-only** — design numbers, and local task IDs within a design.
- **Beads is authoritative for work state once the tasks exist.** The Task Plan keeps the tasks,
  edges, and refs it decided, and never grows a status column.
- **Every task bead carries acceptance criteria** — checkable, citing spec requirement IDs where
  they trace, and naming the test case where the task is verified through the wired product. They
  are what `codefall-implement` verifies before closing the bead.
- **A task that introduces infrastructure, a dependency, a migration, or generated code names the
  local-script change in its criteria.** The scripts stay current at the point of introduction,
  never as a follow-up.
- **Verify the graph before marking the table created.** `bd ready` and `bd dep cycles`, against
  its roots.
- **Every bead write is pushed**: step 9, and every revision.
- **A removed task's bead is reported, never closed silently.** Work may already have happened
  against it.
- **A ticket must not change under someone holding it.** An untouched bead is edited whatever
  changed; a bead someone is holding is replaced when the work already done would no longer count.
- **Status describes the document, never the work.** Work state belongs to Beads.
- **Never record a hop you can derive.** A design carries its spec, or its vision when there is no
  spec — not both.
- **Research goes inline**, attached to the decision it informed. No sibling research files.
- **Push back once, then defer** — on the tier, on the approach, on the cut. The user knows the
  system and you may be wrong.
- **Never overwrite a file that has drifted.** Show the difference and ask.
