# Workers

One background agent per bead, each in its own worktree. Read at step 6, before rendering the first
prompt.

## Launching

`Agent` with `isolation: 'worktree'`, prompted from `../worker-prompt.md`, rendered by substituting
`{{BEAD_ID}}`, `{{TITLE}}`, `{{BODY}}`, `{{ACCEPTANCE}}`, `{{DESIGN_REF}}`, `{{BRANCH}}`,
`{{BASE_REF}}`, `{{PR_TARGET}}`, `{{VERIFY_COMMANDS}}`, `{{RELATES_LINE}}`, and `{{REPO}}`.

- **Single-bead scope also gets a worktree.** The primary checkout stays free for the user.
- **The worker is strategy-blind.** Stacked versus epic is fully encoded in `BASE_REF` and
  `PR_TARGET`. Branch names are root-supplied and deterministic — `feat/<bead>-<slug>` — so a
  stacked link's `BASE_REF` can name its predecessor's branch before that worker exists.
- **Worktree isolation seeds from the repo's default HEAD**, not from `BASE_REF`. The worker's
  first act is its own `git fetch` followed by `git checkout -b {{BRANCH}} origin/{{BASE_REF}}` —
  off the remote-tracking ref, because git's one-branch-one-worktree rule makes a plain checkout
  fail when another worktree holds that branch.
- **Workers never run `bd`.** Everything a worker needs from the tracker travels in its prompt, and
  every tracker write is the root's.

## Chains

**Chains are strictly sequential; parallelism is across chains.** Link N+1 launches only after the
root has verified two facts about link N: the branch is on the remote, and the PR exists. The root
checks, never trusts.

## Results

The worker's final message is exactly one JSON object:

```json
{"bead": "bd-s58", "status": "success", "pr": 102, "branch": "feat/bd-s58-wire-context",
 "discovered": [{"title": "Parser drops trailing comma", "context": "…", "from": "bd-s58"}]}
```

or `{"bead": "…", "status": "failure", "reason": "…"}`. The `discovered` list is how tangent work
reaches the root, which files it — the worker's diff stays scoped to its bead.

## Failure handling

A worker returning `{"status": "failure"}` gets **one automatic retry**: a fresh worker at higher
effort or a stronger model, with the failure reason folded into its prompt. A second failure stops
the chain and escalates — fix it by hand, skip the bead (unclaim it, note why), or abort the run.
Nothing retries more than once on its own. A worker that is alive but stalled is resumed with a
message, not replaced.

## Worktrees survive the run

Cleanup is an offer at close-out — never automatic, and never for a worktree whose PR is still open.
