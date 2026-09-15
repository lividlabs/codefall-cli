# Beads: creating and verifying the graph

Read before creating the graph (step 9).

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
