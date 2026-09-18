# Beads: creating and verifying the graph

Read before creating the graph (step 9).

## Contents

- What gets created
- The test case in a bead's criteria
- The plan file
- Verify the graph

## What gets created

| Bead | One per | Type | Body |
| --- | --- | --- | --- |
| Epic | design document | `epic` | The Overview, and the path to the document |
| Task | Task Plan row | `task`, or `bug` where it is one | What to do, the design ref, and acceptance criteria |

**Every task bead carries acceptance criteria** — two to five checkable statements, drawn from the
spec requirements the task serves and the Hard Constraints that bind it. A criterion that traces to
a spec requirement cites it (`R3: context survives a restart`). Write them as checks, not restated
task text; `codefall-implement` verifies them before closing the bead.

Tier 0 has no epic. One or two beads are created directly, with the reproduction, the cause, and
acceptance criteria in the body — the bead must be self-sufficient.

## The test case in a bead's criteria

A task whose work is verified through the wired product — the real interface, against the real
services, as a user meets it — carries a test case, and the bead's acceptance criteria are where it
is written. Nothing is created here: `codefall-implement` writes the case file from these criteria
before it writes the code, and `codefall-test` runs it.

Three things go in the criteria.

- **The case**, as `<area>/<slug>` — a product area and a hyphenated slug. That is the path the file
  takes under the testing root's `test-cases/`, and it is the case's identifier.
- **The modalities**, `spec`, `agentic`, or both. `agentic` only where verifying an outcome needs
  judgement — generated prose, conversational behaviour, visual wrongness a locator cannot assess.
  `spec` otherwise. Neither where unit tests verify the task, and a driver being available is not a
  reason to add one.
- **Every criterion the case will hold**, each carrying its source. A criterion that restates a spec
  criterion cites it in full — `SPEC-003-REQ-01-AC-01`, the form `codefall-specify` writes and the
  form `grep` finds. A criterion that goes further is marked `derived`, cites the requirement it
  elaborates — `SPEC-003-REQ-01` — and says in one line what it adds: a boundary, a negative path,
  error handling. A case covers every criterion of the requirement it tests, or records why one was
  dropped.

```
Case: checkout/place-order · modalities: spec, agentic
1. An order placed with a valid card reaches a confirmation naming the order id.
   (SPEC-003-REQ-01-AC-01)
2. A declined card leaves the cart intact and states why, without charging.
   (derived from SPEC-003-REQ-01 — the negative path; the spec names only the accepted card.)
```

**The implementation is never a source.** Not the code, not clicking through the application to see
what it does, not a pull request's own text. Reading code for mechanics — a control's role, its
accessible name — is allowed; reading it to decide what should happen is not.

**A gap the criteria expose goes back to the spec.** Where the spec is silent or ambiguous the
criterion is still written, marked `derived`, and the gap goes in this run's report for
`codefall-specify` to add a criterion. Spec numbering is append-only, so the addition takes the next
number and nothing already cited moves. Nothing is patched into the case.

**The user's approval of the task plan is the sign-off** the derived criteria need, so every such
criterion is shown at step 7 and nothing later asks again.

The case-file format `codefall-implement` writes to is
`../../codefall-test/reference/case-file.md`.

Every bead gets `--spec-id` set to the design document's path. The mapping line in the document is
the primary record; `spec_id` finds the beads again if that line is lost.

## The plan file

`bd create --graph` takes a plan file and creates every node and edge in one call, then prints the
mapping the document needs. The plan's node keys **are** the Task Plan's local IDs.

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

## Verify the graph

Creation is not verification, and a reversed edge is invisible in the creation output.

```bash
bd ready          # the unblocked set
bd dep cycles     # must find none
```

**The ready set must be exactly the rows whose Depends on column is `—`**, plus the epic. If a task
with prerequisites is ready, or a root task is not, the edges went in backwards — fix them before
collapsing the table, while the local IDs still line up with what you sent.
