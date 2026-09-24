# When a design changes after its beads exist

Read for the Revise mode. The mapping line says which bead each local ID became, so a later run
updates the graph rather than duplicating it.

**Read the current beads first.** `bd show <id> --json` for each bead in the mapping. The comparison
is against the graph itself, not against a record of what the document used to say.

Then, row by row:

| Case | What happens |
| --- | --- |
| The row changed | Edited or replaced, decided by whether anyone is holding the bead — below |
| The row is new | A new local ID, a new bead named `<prefix>-DESIGN-NNN-Tn`, appended to the mapping line |
| An edge changed | `bd dep add` or `bd dep remove`, then re-verify with `bd ready` and `bd dep cycles` |
| The row is gone | Its local ID retires. **Report the bead and let the user choose** — close it, or leave it open because work already happened against it |

**A removed row is never closed silently.** Somebody may be holding that ticket, and a design edit
is not evidence that its work stopped mattering.

## When a task row changes

Read the bead first — `bd show <id> --json` reports its status, assignee, and comment count.

**Untouched** — `open`, unassigned, no comments — is edited, whatever changed. Nobody has acted on
the old wording, so the size of the change does not matter.

**Touched** — claimed, in progress, commented on, or closed — asks one question: would the work done
against the old wording still be correct and sufficient under the new wording? Yes, edit it. No,
replace it: `bd rename <id> <id>-superseded` to move the old bead out of the way, create the new
bead under the plain `<id>`, and `bd supersede <id>-superseded --with <id>`. The mapping line does
not change: the current bead always holds the plain ID.

A ticket must not change under someone holding it. Where nobody is holding it, editing costs
nothing.

Worked, so two sessions apply it the same way:

| Old row | New row | Bead state | Result |
| --- | --- | --- | --- |
| Wire context load into `/scaffold` | same, with a Design ref added | any | **edit** — nothing about the work changed |
| Add `StageContext` type | Add `StageContext` type + serde | open, unclaimed | **edit** — nobody started; the bead becomes the bigger task |
| Add `StageContext` type | Add `StageContext` type + serde | closed | **replace** — the type exists, the serde does not, and editing a closed bead leaves it unbuilt |
| Emit context on `/envision` exit | Emit context on every verb's exit | claimed | **replace** — they are building one thing and would silently owe five |

`bd supersede <old> --with <new>` closes the old bead with a reference to its replacement, so the
record of what the task used to be survives. **The local ID does not change** — `T2` still means
`T2`; only the bead behind it moves, and the mapping line records the new one.

**A row that has become two separate tasks is a removal plus two additions**, not a replacement, and
the removal goes through the report-don't-close rule above.

**Push when the last row is done.** `bd dolt push` once, after every edit, replacement, and edge
change, so the graph a teammate reads is the one the mapping line now describes.

Say per row which you did and why, in the report.
