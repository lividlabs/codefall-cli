# codefall-implement — where it came from

What this took from elsewhere, what it deliberately did not, and why the rules are shaped the way
they are. None of this is instruction — the skill is the instruction. This exists so nobody re-adds
something that was removed on purpose.

## Taken

**From the dev-implement skill pair** that preceded this extension: the worker-prompt pattern with
strategy encoded in `BASE_REF`/`PR_TARGET`, the go-block, wave execution with verified worker
results, the recovery table, and the humans-merge-`main` rule with its hook.

**From Kiro**: the one-task-at-a-time worker discipline and read-everything-first; drift
reconciliation at resume, which became the recovery table.

**From spec-kit**: parallel means file-disjoint, nothing else.

**From beads' own docs**: close-at-done with gates carrying the merge seam.

## Dropped

**From the dev-implement pair**: `MAX_STACK_DEPTH = 4` — the depth cap solved a cosmetic problem
and a 12-PR stack works; the classification that sent multiple independent chains to an epic branch
— chains parallelize as stacks; the two-skill split — one skill decides scope at run time; and the
board scripts — beads replaced the project board.

**From Kiro and spec-kit**: checklist gates and the converge audit — review is another verb.

## Adapted

**Gates attach to a landed child bead**, not to dependents and not to the epic. A stacked dependent
deliberately builds on its parent's branch, so gating it would block exactly what stacking allows.
`bd gate create` rejects an epic, and an epic already refuses to close while children are open, so
one extra child task carries the gates and the epic follows it.

**Workers do not self-claim with `bd ready --claim`**, diverging from beads' multi-agent docs. That
pattern arbitrates races between peer agents pulling a shared queue; this skill's topology is
dispatcher and workers, the root assigns work top-down, and the default embedded database is
single-writer besides. That is why every `bd` write is the root's: architecture, not etiquette.

## Why there is no depth cap

The cost of a deep stack is cosmetic — the top PR's three-dot diff shows unmerged ancestors until
the chain drains bottom-up. It is disclosed at the go gate, not capped against. Any DAG serializes
into a single stack by topological order, so the single stack is always available and costs only
wall-clock, never structure.

## Why workers are verified, never trusted

A worker can run its whole verification and stop without pushing, and the harness still reports it
completed. So link N+1 launches only after the root has confirmed link N's branch is on the remote
and its PR exists.

## Why the go gate is the only stop

After go, waves proceed on their own and failures are the only mid-run stop. That absence is what
makes an overnight run possible; the user can always interrupt. Background workers cannot answer
permission prompts, which is why the gate states the permissions condition and offers single-task
mode when it fails.

## Why closed means done, not merged

That is beads' own semantics — "closing a beads issue means 'work is done' but the code may still
be on a feature branch" — and it is what lets a stacked dependent start the moment its parent's
branch is pushed. The merge is tracked separately by the gates.

## Why the mirror is coarse

Requirement sub-issues close with the parent, never individually, because beads carry design refs,
not requirement IDs, and a mirror that guesses is worse than one that is coarse.
