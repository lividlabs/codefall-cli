# ADR-008: Upstream Amendments

## Status

Accepted — 2026-09-24

## Context

The chain leaves a document at every step, and each later step reads the ones before it: `specify`
reads the vision, `design` reads the spec and the vision, `implement` and `review` read the design
and the spec. A later step regularly finds an earlier document wrong or incomplete. The design's
Architecture names a type the code has replaced. A spec is silent on a negative path a test case
needs a criterion for. A vision's constraints section names an environment the project has left.

Until this record, every such finding was handed to a later run of the document's own verb, and
the hand-off had four shapes. `design` reported a spec gap "for `codefall-specify`" and marked the
criterion `derived` in the bead; ADR-007 called that a round trip and accepted its cost.
`implement` filed a `design-revision` bead when a worker found the design's text disagreeing with
the code or the spec, and `design`'s Revise mode closed it on a later run. `review` offered the same
bead for a deferred finding the design caused, even though its triage already applies document fixes
on the target's branch. `specify` had no path for a vision it found wrong.

The cost of that hand-off is not the round trip. It is that between the finding and the later run
the document is known to be wrong and nothing on the page says so. A reader of the design sees
`StageStore`; the truth is in a bead they have not listed. A reader of the spec sees four criteria;
the fifth is in a design run's report. Work continues against the wrong text, and when the owning
verb finally runs, its user has to reconstruct what the earlier run knew. Several of the affected
documents have to be amended together, which a bead per document made harder still.

What the hand-off was protecting is real, and it is narrower than the hand-off. Two things
genuinely belong to the owning verb: a change that alters the task graph, because `implement` never
re-cuts what `design` created and a bead's criteria are approved as a plan; and a document whose
lifecycle forbids in-place edits, an `Active` vision or a ratified ADR. Neither is about which verb
is running. A text amendment to a `Ready` spec is the same edit whether `specify` or `design` makes
it, and the ADR that already applied this reasoning is the one `design` writes for a stance change
in its own run.

## Decision

### The rule

A verb that finds an upstream document wrong or incomplete amends it in the same run, when three
things hold:

1. **The document's lifecycle allows the edit.** A `Draft` or `Ready` spec, design, or vision is
   edited in place. An `Active` vision is frozen and is revised only by a new vision through
   `codefall-envision`. A ratified ADR is never rewritten; a stance change is a superseding ADR,
   which `codefall-design` already writes in its own run.
2. **The amendment is text, and moves no work.** It changes no task row in a design's Task Plan,
   retires no criterion a bead already cites, and rewords none. A spec amendment appends, which the
   append-only numbering makes safe: a new criterion takes the next number and nothing already
   cited moves.
3. **The user takes it**, at the confirmation the verb already holds. `design` shows the spec
   amendment with the document at draft-and-confirm; `review` offers it at triage as a fix;
   `specify` shows the vision amendment with the spec; `implement`'s go gate says that workers
   amend a design's or spec's text where the code disagrees, and the pull request is where the user
   sees it.

When any of the three fails, the finding is filed as a `design-revision` bead against the design,
exactly as before, and the owning verb's Revise mode settles it.

### Distance does not matter

The rule applies to any document above the current step. `implement` finding the spec wrong is the
same case as `implement` finding the design wrong; `design` finding the vision wrong is the same
case as `design` finding the spec wrong. The tests are the document's lifecycle and whether work
moves, never how many hops up the document sits.

### All layers or none

Every document between the change and the current step that restates the point is amended in the
same run, or none is and the whole chain goes in the bead with each affected document named. A spec
changed while the design still restates the old text leaves two documents disagreeing with no bead
saying so, which is worse than the hand-off this record replaces.

### Where the amendment lands

On the verb's own branch, in the verb's own pull request, and named in its report. `design`'s
branch carries the spec amendment beside the design. A worker's branch carries the design or spec
amendment beside the code, and the pull request body names it. `review` edits on the target's
branch, where its other document fixes already land. `specify`'s branch carries the vision
amendment beside the spec. Nothing about landing changes: the branch, the commit, and the offered
pull request are the ones `landing.md` already describes.

### The tracker mirror follows the amendment

A spec amendment is re-mirrored to the requirement's tracker issue by the run that made it, through
`codefall-specify`'s tracker profile. Under `implement`, the root does this at the wave boundary,
because every tracker write is the root's.

### What stays a bead

- A change that adds, removes, or rewords a task row, or changes the criteria a bead cites. That is
  graph reconciliation, and `codefall-design`'s Revise mode does it, now including the case where
  the document found wrong is the spec: the spec is amended, re-mirrored, and the bead's criteria
  updated together.
- A disagreement a worker could not finish its task under. That is a failure result and an
  escalation, as before, never a revision bead.
- A conflict between a task and an ADR. That reaches `codefall-design` as a superseding-ADR
  conversation, as before.
- An amendment the user declines. For a spec gap, the bead's criterion stays `derived`, which is
  what `derived` now means: the spec was offered the criterion and did not take it.

### Concurrency under `implement`

Two workers in one wave may each amend the same document. The root checks at the wave boundary; where
two amendments touch the same section it applies one and says so, and the restack at integration
reports any conflict that remains rather than resolving it silently.

## Consequences

- **A document is wrong for one pull request, not for an unknown number of runs.** The amendment
  is reviewable with the work that needed it, and the reader of the document after the merge sees
  the current text.
- **`design-revision` beads become rare and specific.** Each one now says why it could not be an
  amendment: the graph moves, the document is frozen, or the user declined. A run that files one
  is saying something, where before it was the only path.
- **ADR-007's round-trip consequence is revised.** The gap a case's criteria expose still goes to
  the spec, and the spec is still amended by appending; what changes is that `codefall-design`
  appends it in its own run, and the bead cites the real criterion. ADR-007 is not rewritten and
  not superseded; this record names the one consequence it changes.
- **`derived` narrows.** It marked every criterion beyond the spec; it now marks the ones the spec
  declined, and a bead written after this record cites a real identifier for the rest.
- **Verbs write to documents they do not own.** The rule is bounded by the three tests, but a
  reader of `docs/specs/` can no longer assume `codefall-specify` wrote every line. The commit
  history says which verb did, and each verb's report names the amendment.
- **Workers gain a file outside their predicted scope.** A design or spec document under
  `docs/` is a small, prose-only conflict surface, and the wave-boundary check is what keeps two
  workers from racing on it. A conflict that reaches integration is reported, as any other is.
- **`implement` re-mirrors a spec.** The root already updates the parent issue's state; appending a
  criterion to a requirement's issue is one more write through the same profile.

## Related

- ADR-007, *Test Cases* — the `derived` marker and the round-trip consequence this record revises,
  and the append-only numbering it relies on.
- `docs/workflow.md`, *The chain* — where the rule is stated for a project.
- `extensions/skills/codefall-design/reference/revising.md` — the Revise mode that settles what
  stays a bead.
- `extensions/AGENTS.md`, *ADRs are history* — why a ratified ADR is the one document no verb amends.
- `docs/decision-log.md`, *Upstream amendments, 2026-09-24* — the alternatives seen and not taken.
