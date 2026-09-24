# Review: feat-design-revision-beads

**Reviewed:** 2026-09-24T04:35:51Z · **Reviewer:** muse/muse-spark-1.3-contributor · **Revision:** e765d710926b (base d61516ae3d9e)
**Lenses:** conventions, docs, simplify, precision

## Findings

### 1. Both new README paragraphs state the revision-bead trigger as design-vs-code only, omitting the design-vs-spec case every operative file includes. — `minor` · `docs` · fixed
`README.md:311-314`

When a task or review finds the design disagreeing with the spec while matching the code: implement/SKILL.md, reference/beads.md, worker-prompt.md, revising.md, review/SKILL.md, and both workflow documents all file a bead, but a reader of README.md lines 311-314 or 364-368 would conclude none is filed.

```diff
--- a/README.md
+++ b/README.md
@@ design section
-downstream: `implement` and `review` file a `design-revision` bead when the design's text and the
-code disagree, and `design` reads those at its start, so the run that finds a design wrong is not
-the only place that knows it.
+downstream: `implement` and `review` file a `design-revision` bead when the design's text disagrees
+with the code or with the spec, and `design` reads those at its start, so the run that finds a
+design wrong is not the only place that knows it.
@@ implement section
-its task despite the design's text disagreeing with the code reports the disagreement, and the run
-files it as a `design-revision` bead against the document.
+its task despite the design's text disagreeing with the code or with the spec reports the
+disagreement, and the run files it as a `design-revision` bead against the document.
```

### 2. The relative clause 'which `codefall-design` reads on its next run' attaches to `reference/beads.md`, saying the design skill reads implement's reference file instead of the filed bead. — `minor` · `precision` · fixed
`extensions/skills/codefall-implement/SKILL.md:58-62`

Always, for any reader parsing the sentence; `codefall-design` lists revision beads via `bd list`, and nothing in its skill reads implement's beads.md.

```diff
--- a/extensions/skills/codefall-implement/SKILL.md
+++ b/extensions/skills/codefall-implement/SKILL.md
@@
-smaller disagreement the task can finish under — the design's text against the code, or against
-the spec — is filed as a `design-revision` bead per `reference/beads.md`, which `codefall-design`
-reads on its next run.
+smaller disagreement the task can finish under — the design's text against the code, or against
+the spec — is filed as a `design-revision` bead per `reference/beads.md`; `codefall-design`
+lists such beads on its next run.
```

### 3. 'They decide between Revise and Add tasks' gives the mode decision to the revision requests, while the skill elsewhere reserves it to the user. — `minor` · `precision` · fixed
`extensions/skills/codefall-design/SKILL.md:211-212`

When open revision requests point at pure new work suited to Add tasks: a reader may take the requests as forcing Revise, against Other modes ('Ask which if it is not obvious', 'never transition a design on your own initiative').

```diff
--- a/extensions/skills/codefall-design/SKILL.md
+++ b/extensions/skills/codefall-design/SKILL.md
@@
-Report what you found before designing — open revision requests first; they decide between
-Revise and Add tasks. If the work already exists, say so and stop.
+Report what you found before designing — open revision requests first; use them to decide between
+Revise and Add tasks. If the work already exists, say so and stop.
```

## Not checked

- Whether `bd list -l design-revision --spec <path>` (codefall-design/SKILL.md step 3, revising.md) and `bd create --spec-id <path>` (implement reference/beads.md, agents/sections/beads.md, codefall-review/SKILL.md) address the same bead field: the diff uses both spellings, the bd CLI reference was out of scope, and a possible spelling inconsistency is therefore not called.
- Settled by the session after the review: `bd list --spec` filters on `spec_id`, the field `bd create --spec-id` sets, per `bd list --help` and `bd create --help`.
