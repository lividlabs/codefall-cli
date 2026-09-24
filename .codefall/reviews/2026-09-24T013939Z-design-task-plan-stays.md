# Review: design-task-plan-stays

**Reviewed:** 2026-09-24T01:39:39Z · **Reviewer:** muse/muse-spark · **Revision:** 602a4a40542d (base 26fc5d78ae3e)
**Lenses:** correctness, behaviour, conventions, docs, simplify

## Findings

### 1. README.md still describes the Task Plan table as replaced by an ID-mapping line on approval, contradicting the change that keeps the table. — `important` · `behaviour` · fixed
`README.md:299-303`

Whenever a reader reads the README section on the task plan at the reviewed revision.

README.md still describes the Task Plan table as replaced by an ID-mapping line on approval, contradicting the change that keeps the table. Fixed in the commit `docs(readme): describe the Task Plan table staying after creation`, which landed after the review started and which the reviewer did not see.

```diff
--- a/README.md
+++ b/README.md
@@ -296,9 +296,10 @@
-Once you approve it, those rows become beads and the table is replaced by the line that records what
-became what — `Epic: booking-DESIGN-007 · T1→booking-DESIGN-007-T1 · T2→booking-DESIGN-007-T2`. The
-bead IDs are the document's own numbering behind the project's Beads prefix, so a person can read
-and say them. Beads is authoritative from that moment, and a
-duplicate task list left behind in a git-tracked file would drift from it. The mapping is what lets a
-later run update the graph instead of duplicating it.
+Once you approve it, those rows become beads and the table stays. The callout above it records the
+date and the epic, one bead per row as `<prefix>-DESIGN-NNN-Tn`. Beads is authoritative for work
+state from that moment; the table holds only the tasks, edges, and design refs, and a later run
+revises the design rather than duplicating the graph.
```

## Not checked

- Whether the older docs/decision-log.md bullets that name the mapping line and the collapse (lines 1194, 1215, 1223) should be amended or left as append-only history; the branch appends a new bullet without touching them, matching the precedent where the envision rename left the earlier conceptualize wording in place, so they were read but not filed.
