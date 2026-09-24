# When a design changes after its beads exist

Read for the Revise mode. A row and its bead share a number — `Tn` is `<prefix>-DESIGN-NNN-Tn` —
so a later run edits the graph the table already describes rather than duplicating it. The user
edits the table; this mode brings the beads level with it.

**Read the current beads first.** `bd show <prefix>-DESIGN-NNN-Tn --json` for every row, and
`bd list --parent <prefix>-DESIGN-NNN` for anything the table no longer names. The comparison is
against the graph itself, not against a record of what the document used to say.

A design that still carries a mapping line in place of a table was created before the table stayed.
Rebuild the table from the beads — title, blockers, and design ref from each `bd show` — and rewrite
the callout to the current form before going row by row.

Then, row by row:

| Case | What happens |
| --- | --- |
| The row changed | Edited or replaced, decided by whether anyone is holding the bead — below |
| The row is new | A new local ID, the next after the highest ever used, and a new bead named `<prefix>-DESIGN-NNN-Tn` |
| An edge changed | `bd dep add` or `bd dep remove`, then re-verify with `bd ready` and `bd dep cycles` |
| The row is gone | Its local ID goes on the `Retired:` line under the table. **Report the bead and let the user choose** — close it, or leave it open because work already happened against it |

**A removed row is never closed silently.** Somebody may be holding that ticket, and a design edit
is not evidence that its work stopped mattering.

## When a task row changes

Read the bead first — `bd show <id> --json` reports its status, assignee, and comment count.

**Untouched** — `open`, unassigned, no comments — is edited, whatever changed. Nobody has acted on
the old wording, so the size of the change does not matter.

**Touched** — claimed, in progress, commented on, or closed — asks one question: would the work done
against the old wording still be correct and sufficient under the new wording? Yes, edit it. No,
replace it: `bd rename <id> <id>-superseded` to move the old bead out of the way, create the new
bead under the plain `<id>`, and `bd supersede <id>-superseded --with <id>`. The row does not
change: the current bead always holds the plain ID.

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
`T2`; only the bead behind it moves, and the plain ID names the new one.

**A row that has become two separate tasks is a removal plus two additions**, not a replacement, and
the removal goes through the report-don't-close rule above.

**Then date the callout and push.** Add today to the callout — `Created in Beads <date>, revised
<today>, as …` — and `bd dolt push` once, after every edit, replacement, and edge change, so the
graph a teammate reads is the one the table now describes.

Say per row which you did and why, in the report.
