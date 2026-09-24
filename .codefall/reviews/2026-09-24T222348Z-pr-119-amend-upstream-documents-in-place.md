# Review: pr-119-amend-upstream-documents-in-place

**Reviewed:** 2026-09-24T22:23:48Z · **Reviewer:** muse/muse-spark-1.3-contributor · **Revision:** f8241a514acf (base d42464475556)
**Lenses:** correctness, behaviour, conventions, docs, simplify, precision, structure, status, alternatives, consequences, coherence

## Findings

### 1. The wave-boundary mirror tells the root to regenerate a worker-amended requirement's tracker issue without saying which copy of the spec to read, so regenerating from the primary checkout silently omits the new criterion. — `important` · `correctness` · fixed
`extensions/skills/codefall-implement/reference/beads.md:135-137`

When a worker amends a spec in its own branch: the primary checkout never carries worker-branch edits, and the worker's amended list holds only document path, section, and summary — not the criterion text the Refreshing sequence regenerates the issue body from.

```diff
--- a/extensions/skills/codefall-implement/reference/beads.md
+++ b/extensions/skills/codefall-implement/reference/beads.md
@@ -132,7 +132,7 @@
 - **A spec was amended.** Regenerate the requirement's tracker issue, the existing-requirement case
-  of the *Refreshing* sequence in the spec's tracker profile, as the mirror reference beside this
-  file says.
+  of the *Refreshing* sequence in the spec's tracker profile, as the mirror reference beside this
+  file says — reading the amended text from the worker's branch (`git show origin/<branch>:<path>`) or its surviving worktree, never the primary checkout, which does not carry it.
 - **The close reason** names the amendment beside what was verified, so the bead's record says the
   document moved with the work.
```

### 2. README states that only a move-work disagreement is filed as a design-revision bead, while ADR-008 and the skills also file beads when the document is frozen or the user declines the amendment. — `minor` · `docs` · fixed
`README.md:369-373`

When a finder cannot amend for a reason other than moving work: an Active vision or ratified ADR found wrong, or an amendment the user declines at confirmation — all of which ADR-008 and reference/beads.md file as design-revision beads. The same over-narrowing appears at README.md:316 ("but only for a disagreement that moves work").

```diff
--- a/README.md
+++ b/README.md
@@ -313,9 +313,10 @@
 own, because someone may still be working it. A request to revise can also arrive from downstream,
-as a `design-revision` bead, but only for a disagreement that moves work: `implement` and `review`
+as a `design-revision` bead, when the finder cannot amend it itself: `implement` and `review`
 amend the design's text themselves when the code disagrees with it, in their own pull request, and
-file the bead when the fix would change a task row or a criterion a bead cites.
+file the bead when the fix would move work, the document is frozen, or the user declined the amendment.
@@ -369,8 +370,8 @@
 finish its task despite the design's text, or the spec's, disagreeing with the code amends that text
 in its own branch, names the amendment in its pull request, and the root re-mirrors a spec change to
-its tracker issue. Only a disagreement that would move work — a task row, a criterion a bead cites —
-is filed as a `design-revision` bead, and the close-out names those beads separately from code
+its tracker issue. A disagreement the finder cannot amend — one that would move work (a task row, a criterion a bead cites), touch a frozen document, or the user declined —
+is filed as a `design-revision` bead, and the close-out names those beads separately from code
 follow-ups and tells you to run `design` on that document. A disagreement the task cannot finish
```

### 3. The reference promises a Draft spec is amended in the run, but step 2 of the design SKILL stops any run framed by a Draft spec, so the two instructions cannot both be followed. — `minor` · `correctness` · fixed
`extensions/skills/codefall-design/reference/beads.md:66-75`

When codefall-design is invoked against a spec whose Status is Draft: the step-2 gate (SKILL.md:192-199) stops the run before it ever reaches the amendment step the beads reference describes.

```diff
--- a/extensions/skills/codefall-design/reference/beads.md
+++ b/extensions/skills/codefall-design/reference/beads.md
@@ -71,7 +71,7 @@
 `../../codefall-specify/trackers/<name>/PROFILE.md`, the existing-requirement case alone. On no, the
 criterion is written marked `derived`, which from here on means the spec was offered it and
-declined. A spec that is `Draft` is amended the same way; one that is `Archived` is not a target
-for design at all. Nothing is patched into the case.
+declined. A spec that is `Draft` is amended the same way, once the run is past the step-2 readiness gate — a run framed by a `Draft` spec still stops there; one that is `Archived` is not a target
+for design at all. Nothing is patched into the case.
```

### 4. Step 8 says every taken amendment is "mirrored", but a vision amendment has no tracker mirror, so a run amending a vision has no mirror step to perform. — `minor` · `precision` · fixed
`extensions/skills/codefall-design/SKILL.md:311`

When a design run amends a vision (the case ADR-008 names explicitly) alongside or instead of a spec: the Upstream section defines regenerating the requirement issue only for spec amendments.

```diff
--- a/extensions/skills/codefall-design/SKILL.md
+++ b/extensions/skills/codefall-design/SKILL.md
@@ -308,7 +308,7 @@
 is a tier 0 run with no ADR, which writes no file.
 
 Write `docs/designs/DESIGN-NNN-slug.md` with the Task Plan **staged**, the ADR if there is one,
-`docs/designs/AGENTS.md` if it was missing, and the amendments the user took, each mirrored.
+`docs/designs/AGENTS.md` if it was missing, and the amendments the user took — spec amendments mirrored to the tracker per above, vision amendments committed beside the spec.
 
 Write the document before creating the beads, so a failed creation leaves a resumable run.
```

## Not checked

- `comments`, `tests`, `types`, `failures`, `local`, `security` did not run: the diff is skill procedure and documentation, with no code.
