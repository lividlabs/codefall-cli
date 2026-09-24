# Workers

One background agent per bead, each in its own worktree. Read at step 6, before rendering the first
prompt.

## Launching

`Agent` with `isolation: 'worktree'`, prompted from `../worker-prompt.md`, rendered by substituting
`{{BEAD_ID}}`, `{{TITLE}}`, `{{BODY}}`, `{{ACCEPTANCE}}`, `{{DESIGN_REF}}`, `{{BRANCH}}`,
`{{BASE_REF}}`, `{{PR_TARGET}}`, `{{VERIFY_COMMANDS}}`, `{{RELATES_LINE}}`, `{{REPO}}`,
`{{TESTING_ROOT}}`, and `{{CASE_FILE_FORMAT}}`.

**The two testing placeholders are resolved by the root, not the worker.**

| Placeholder | What the root renders |
| --- | --- |
| `{{TESTING_ROOT}}` | `test.dir` from `.codefall/settings.json`, as the worker sees it from the repository root — `testing` |
| `{{CASE_FILE_FORMAT}}` | the absolute path of `codefall-test`'s `reference/case-file.md` in the skills directory this run was invoked from — `<repo>/.claude/skills/codefall-test/reference/case-file.md`, or `.agents/skills/…` under a harness that mirrors the tree |

A worker reads the format by that absolute path because it cannot resolve a skill-relative one: it
runs in the project, not in the skills directory. Render both on every prompt; a bead whose criteria
name no case simply never reads them.

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
{"bead": "booking-DESIGN-007-T2", "status": "success", "pr": 102,
 "branch": "feat/booking-DESIGN-007-T2-wire-context",
 "discovered": [{"kind": "code", "title": "Parser drops trailing comma", "context": "…",
                 "from": "booking-DESIGN-007-T2"},
                {"kind": "design", "title": "§ Architecture names a StageStore the code replaced",
                 "context": "…", "from": "booking-DESIGN-007-T2"}]}
```

or `{"bead": "…", "status": "failure", "reason": "…"}`. The `discovered` list is how tangent work
reaches the root, which files it — the worker's diff stays scoped to its bead. `kind` is `code` or
`design` and picks the form under *Discovered work* in `beads.md`, so the root files each item
without re-reading the worker's prose.

## Failure handling

A worker returning `{"status": "failure"}` gets **one automatic retry**: a fresh worker at higher
effort or a stronger model, with the failure reason folded into its prompt. A second failure stops
the chain and escalates — fix it by hand, skip the bead (unclaim it, note why), or abort the run.
Nothing retries more than once on its own. A worker that is alive but stalled is resumed with a
message, not replaced.

## Worktrees survive the run

Cleanup is an offer at close-out — never automatic, and never for a worktree whose PR is still open.
