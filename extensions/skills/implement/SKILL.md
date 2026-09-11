---
name: implement
description: Execute the work design put into the graph — claim ready beads, build each task in its own worktree with tests as part of done, verify against the bead's acceptance criteria and the project's own checks, open pull requests, and walk the dependency graph in parallel waves until the frontier is empty. Never merges to main.
argument-hint: "[a bead, an epic, a design, or nothing to pick from ready work]"
disable-model-invocation: true
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Write
  - Edit
  - Bash
  - Agent
---

# Implement

Walk the graph `design` created. Claim what is ready, build it, verify it, open a pull request, and
let each close unblock the next task until the frontier is empty.

Implementing is not merging. A run ends at open pull requests and a reported merge order — **a human
performs every merge to `main`, and this skill never does**, in any mode, under any instruction
short of the user editing this file. Merges into an epic branch are the one exception, because the
epic branch exists to fan work back in and the human gate sits at its aggregate PR.

Plugin paths in this document — the ones that start with `../` — are relative to this skill's
directory, the one holding this `SKILL.md`. Resolve them from where the file lives; they are not
relative to the user's project.

## Scope — build, not decide

| In scope | Out of scope | Whose |
| --- | --- | --- |
| Executing tasks from the graph | What the tasks are, or their edges | `design` |
| Branches, worktrees, commits, PRs | Merging anything to `main` | the user |
| Every test the current work needs | Regression campaigns and fresh-context retesting | `test` |
| Harness checks on its own diffs | Independent review and verdicts | `review` |
| Bead lifecycle: claim, close, discovered work | Creating or re-cutting the task graph | `design` |
| The concept's `Active` transition | Any other document transition | the owning verb |
| Mirroring work state to the spec's tracker issue | The mirror's lifecycle and labels | `specify` |

The line that matters most is the top one. **A task that turns out to be wrong is reported, not
redesigned.** When the design's cut does not survive contact with the code, say what you found and
hand the graph back to `design` — quietly building something else leaves a graph that lies.

Two boundaries inside testing, decided deliberately:

- **Implement writes every test the current work needs** — planned by the design's Testing Strategy
  or discovered mid-task, unit, integration, and end-to-end alike. Discovering a missing test is not
  a reason to reopen `design`; write it.
- **The `test` verb owns what comes after the work lands**: regression passes, coverage campaigns,
  agentic testing in a fresh context. Implement's tests prove this change; `test` re-proves the
  system.

## One bead or the graph

The argument fixes the scope; the skill never infers it.

| Invocation | Scope |
| --- | --- |
| `/implement bd-abc` | That bead, alone |
| `/implement bd-a2g` (an epic) | The epic's whole graph, until its ready set is empty or a gate stops the run |
| `/implement DESIGN-007` | The design's epic, via the mapping line in its Task Plan |
| `/implement` | Show ready work grouped by epic and ask |

A design whose Task Plan still says `Staged. Not yet in Beads` has no graph to walk. Refuse and
point at `/design` — its step 9 creates the beads.

**Tier 0 is not a separate mode.** A bare bug bead is single-bead scope with a shorter reading list;
the lighter path falls out of the context rules, not out of a flag.

**Scope and execution style are different decisions.** Picking an epic decides *what*; whether it
runs as one serialized stack or parallel workers is decided afterward from the graph's shape, at the
go gate.

## What gets read

Per bead, in order, before any plan is formed:

1. **The bead** — `bd show <id> --json`: the description, the Design ref, the acceptance criteria,
   and `spec_id`, which `design` set to the design document's path.
2. **The design document** — Overview and Architecture always; the specific section the Design ref
   names (that column exists so this skill reads the relevant fifteen lines, not the whole
   document); Hard Constraints; Technical Context and Testing Strategy when present.
3. **The spec**, one hop up the design's `Related` line — its acceptance criteria are the
   externally observable contract. The concept only when there is no spec.
4. **ADRs** — the ones on the design's `Related` line plus the project's `docs/adrs/` baseline.
   Never rewritten; a conflict between a task and an ADR goes back to `design` as a superseding-ADR
   conversation, never a quiet exception.
5. **The project's `AGENTS.md`**, root and scoped — workflow rules, verify commands, conventions.
6. **Mockups** under `docs/mockups/` when referenced. A working mockup is still a drawing: it
   proves an interaction and gets rebuilt in the app's stack. Never copy its markup.

A tier-0 bead has no document behind it; the list collapses to bead + `AGENTS.md` + ADRs, which is
why `design` writes tier-0 beads self-sufficient.

## Landing strategies

Three ways work reaches `main`. The graph's shape picks one; the user can overrule it at the go
gate. The project's `AGENTS.md` constrains the choice before either speaks, `CUSTOMIZE.md` tunes it
within those constraints, and a conflict between the two is drift — show the difference and ask,
never silently pick.

| Strategy | When | Shape |
| --- | --- | --- |
| **Parallel stacks** (default) | Independent chains with disjoint predicted file scopes | Each chain stacks toward `main`; chains run in parallel |
| **Single stack** | Overlapping file scopes, uncertainty, or fan-in without a need for parallelism | The whole graph in topological order, one branch atop another |
| **Epic branch** | Fan-in across chains, or work that must not land on `main` in increments — *and* parallelism matters | Workers branch off `epic/<id>-<slug>`, PRs target it, one aggregate PR to `main` |

**There is no depth cap.** Depth alone never forces the epic branch. The cost of a deep stack is
cosmetic — the top PR's three-dot diff shows unmerged ancestors until the chain drains bottom-up —
and it is disclosed at the go gate, not capped against.

**Any DAG serializes into a single stack.** Topological order makes one branch-atop-branch chain out
of any graph, so the single stack is always available and costs only wall-clock, never structure.
That is the fallback whenever parallel stacks cannot be shown safe.

**Fan-in is the epic branch's real trigger.** A bead whose parents sit on different branches has no
single stack base. Run it serially as one topological stack, or give the chains an integration
branch — never pause mid-run waiting for a human to merge the parents; an unattended run must not
block on a person.

### Predicting whether stacks can run in parallel

Parallel stacks require **predicted-disjoint file scopes**, derived before the go gate: each chain's
Design refs name components, components own files, and disjoint components usually mean disjoint
files.

**Hotspot files count as overlap by default.** Lockfiles, generated artifacts, migrations, shared
type barrels, route and DI registries — cross-cutting files are where parallel work actually
collides. Two chains that both touch one are serialized unless the overlap can be shown harmless.
The burden runs toward serializing: a single stack is always correct.

**Prediction is verified at integration, never trusted.** After building completes, restack the
stack bottoms onto current `main` and re-run verification before reporting the merge order. A
conflict found here is resolved deliberately, in the open — not by an automated fix-up.

### The diagrams

The go gate renders the plan as a branch diagram built from the actual graph. Stacked:

```
main ──┬── bd-a1 ── bd-a2 ── bd-a3 ── bd-a4    stack A · parser chain    [src/parse/**]
       └── bd-b1 ── bd-b2                      stack B · CLI chain       [src/cli/**]
             2 stacks in parallel · merges drain bottom-up · A and B independent
```

Epic branch:

```
main ── epic/bd-e7-stage-context
          ├── bd-t1 ─┐
          ├── bd-t2 ─┼── bd-t4    fan-in: t4 needs t1 + t2
          └── bd-t3 ─┘            wave 1: t1 t2 t3 · wave 2: t4
```

## The go gate

One approval, before any work starts. Everything the run will do, in one block:

- the landing strategy and its one-line reason ("linear chain of 7, no fan-in → single stack");
- the branch diagram;
- the waves, and how many workers run concurrently in each — **there is no default cap**; the wave
  is sized by the graph and the file scopes, and the user trims it here if it is too wide;
- the model proposed per bead, and one session-level effort recommendation as the exact command —
  "recommend `/effort high` before go." When one bead wants far more than the rest, propose it as
  its own batch instead of running everything expensive;
- what will be claimed in beads, and — when a concept sits behind the work — that go flips it to
  `Active`;
- the permissions condition: background workers cannot answer permission prompts, so the session
  must allow edits and Bash without prompting, or the run offers single-task mode instead.

**Single-bead scope shrinks the gate to a plan approval**: the files to touch, the approach, the
test plan, one branch. Approve, then it builds.

**After go, waves proceed on their own.** Failures are the only mid-run stop. That absence is
deliberate — it is what makes an overnight run possible — and the user can always interrupt.

Model choice is the root's, made at the gate, not `design`'s: task understanding and the model
lineup both shift between design day and implement day. A bead already carrying execution metadata
is taken as a recommendation and shown in the table; **implement never writes or edits bead
metadata** — what actually ran is recorded in a `bd comment`, beside the PR link.

## Workers

One background agent per bead, each in its own worktree — `Agent` with `isolation: 'worktree'` —
prompted from `worker-prompt.md` beside this file, rendered by substituting `{{BEAD_ID}}`,
`{{TITLE}}`, `{{BODY}}`, `{{ACCEPTANCE}}`, `{{DESIGN_REF}}`, `{{BRANCH}}`, `{{BASE_REF}}`,
`{{PR_TARGET}}`, `{{VERIFY_COMMANDS}}`, `{{RELATES_LINE}}`, and `{{REPO}}`.

**Single-bead scope also gets a worktree.** The primary checkout stays free for the user — that is
the point of always isolating, and it is why `bd` runs with `-C` against the primary checkout when
the session's working directory is elsewhere.

The worker is strategy-blind: stacked versus epic is fully encoded in `BASE_REF` and `PR_TARGET`,
and the worker never needs to know which is in play. Branch names are root-supplied and
deterministic — `feat/<bead>-<slug>` — so a stacked link's `BASE_REF` can name its predecessor's
branch before that worker exists.

**Worktree isolation seeds from the repo's default HEAD**, not from `BASE_REF`. The worker's first
act is therefore its own `git fetch` followed by `git checkout -b {{BRANCH}} origin/{{BASE_REF}}` —
off the remote-tracking ref, because git's one-branch-one-worktree rule makes a plain checkout fail
when another worktree holds that branch.

**Chains are strictly sequential; parallelism is across chains.** Link N+1 launches only after the
root has verified two facts about link N: the branch is on the remote, and the PR exists. A worker
can run its whole verification and stop without pushing — the harness still reports it completed —
so the root checks, never trusts.

**Workers never run `bd`.** Everything a worker needs from the tracker travels in its prompt, and
every tracker write is the root's. The default beads setup is an embedded single-writer database;
this rule is architecture, not etiquette.

**Failure handling.** A worker returning `{"status": "failure"}` gets **one automatic retry**: a
fresh worker at higher effort or a stronger model, with the failure reason folded into its prompt.
A second failure stops the chain and escalates — fix it by hand, skip the bead (unclaim it, note
why), or abort the run. Nothing retries more than once on its own. A worker that is alive but
stalled is resumed with a message, not replaced.

The worker's final message is exactly one JSON object:

```json
{"bead": "bd-s58", "status": "success", "pr": 102, "branch": "feat/bd-s58-wire-context",
 "discovered": [{"title": "Parser drops trailing comma", "context": "…", "from": "bd-s58"}]}
```

or `{"bead": "…", "status": "failure", "reason": "…"}`. The `discovered` list is how tangent work
reaches the root, which files it — the worker's diff stays scoped to its bead.

**Worktrees survive the run.** Cleanup is an offer at close-out — never automatic, and never for a
worktree whose PR is still open.

## Verification and done

A bead is done when three things are true: **its acceptance criteria hold, the project's checks are
green, and its PR is open.** Done is not merged — [Beads](#beads) covers that seam.

### Where the commands come from

This plugin is project-agnostic; implement hardcodes no build, lint, or test invocation. Resolution
order:

1. `.codefall/skills/implement/CUSTOMIZE.md` — verb-specific tuning, such as a fast subset per bead
   with the full suite reserved for pre-PR;
2. the project's `AGENTS.md` — scaffolded projects already carry the command list in their
   verification section;
3. inference from the repo (`package.json` scripts, `Makefile`, `go.mod`) — stated at the go gate,
   with an offer to record the inferred commands in `AGENTS.md` so inference happens once.

The resolved list is passed into worker prompts. Workers re-derive nothing.

### The checks

- The project's own verification commands, run until clean.
- The harness's built-in passes on the bead's own diff, where the harness provides them: `simplify`
  always; `code-review` and `security-review` when available. These are checks inside implement,
  not review — implement renders no verdict on its own work, and independent review stays the
  `review` verb's.
- The bead's acceptance criteria, checked one by one. What passed goes into the close reason. A
  bead with no acceptance field — tier 0, hand-made, pre-convention — falls back to the design's
  Hard Constraints plus the spec's criteria, and the close reason still records what was verified.

**Tests are part of done, not a follow-up.** The tests the design planned, and the tests the work
turned out to need — a gap found mid-task is written now, not filed for later and not sent back to
`design`.

## Beads

Every `bd` command in a run is the root session's, executed against the primary checkout. The
default beads setup is an embedded Dolt database — one writer at a time, no server — and its sync
channel is the repo's own git remote, under a ref plain git never shows (`refs/dolt/data`).

### Session start

```bash
bd dolt pull            # teammates' claims and closes, when a remote is wired
bd gate check           # resolve gates for PRs merged since last session
bd ready --mol <epic>   # the claimable frontier
```

A project with no Dolt remote works identically; state is this-machine-only, and the skill says so
once rather than treating it as an error. `bd dolt pull` failing for lack of a remote is that
statement's trigger, not a stop.

`bd gate check` fails the same way — safely. With `gh` missing, unauthenticated, or no GitHub
remote, every gate stays open and the command reports per-gate errors and still exits 0: a report,
not a stop, and never a falsely resolved gate. An `ESCALATE` line is different — the gate was
checked and its PR is missing — and is surfaced to the user rather than skimmed past.

### Claim, work, close

```bash
bd update <epic> <bead> --claim     # epic scope: the epic itself is claimed on the run's first claim
bd dolt push                        # the claim is visible to the team before the work, not after
```

Single-bead scope claims only its bead — the epic stays unclaimed so coworkers are free to take
siblings — and creates no landed bead; the PR-link comment is the merge trail, the same as a
standalone bead.

Work happens in git; commits carry the bead ID — `feat: add StageContext type (bd-unz)`.

```bash
bd comment <bead> "PR #101 · feat/bd-unz-stage-context · built with opus/high"
bd close <bead> -r "done: criteria R1,R2 verified, checks green, PR #101 open" --suggest-next
bd dolt push
```

**The close is the engine.** Closing a bead is what removes it as a blocker, so `--suggest-next`
prints the next wave straight from the graph — the edges `design` wired are the sequencer, and this
skill keeps no schedule of its own.

**Closed means done, not merged.** That is beads' own semantics — "closing a beads issue means
'work is done' but the code may still be on a feature branch" — and it is what lets a stacked
dependent start the moment its parent's branch is pushed. The merge is tracked separately:

### The landed bead and its gates

An epic cannot be gated directly (`bd gate create` rejects it), and an epic already refuses to
close while children are open. So an epic run's first claim also creates one extra child task —
`Land: all PRs merged to main` — with gates blocking it:

```bash
bd gate create --type=gh:pr --blocks <land-bead> --await-id=<pr-number> -r "PR #<n>"
```

`bd close` structurally refuses an issue with unsatisfied gates, so the landed bead cannot close —
and therefore the epic cannot close — until every PR is merged. A later session's `bd gate check`
clears the gates as merges land, the landed bead closes, and the epic follows. **Epic closed always
means the code is on `main`.**

**Which PRs get gates depends on the strategy.** A stacked run gates every PR as it opens — each is
its own landing on `main`. An epic-branch run gates none of its worker PRs: they target the epic
branch and the root merges them at the wave boundary, so a gate on one would resolve while nothing
is on `main`. Its one gate is created at integration, for the aggregate PR — the only landing that
matters.

A standalone bead gets no gate; its PR link in the comment is the merge trail.

### Discovered work

```bash
bd create "Parser drops trailing comma" --deps discovered-from:<bead> -p 2
```

One command, the dep rides the create, and the tangent is in the graph instead of in someone's
diff. Workers report discoveries in their result JSON; the root files them.

### Session end

Final `bd dolt push`. There is no other landing ritual — every write already landed when its
command ran, and the database is the handoff: the next session opens with `bd ready` and sees
exactly where this one stopped.

## Merges and the mirror

**The root never merges to `main`.** The plugin ships a `PreToolUse` hook that denies it
mechanically — a denial from that hook is the system working as designed, not an obstacle to route
around. What the root does merge: worker PRs into the **epic branch**, one at a time, at wave
boundaries — the epic branch exists to fan waves back in.

At close-out the run reports the **merge order** — bottom-up per stack, GitHub retargets each PR as
its base merges — and stops. After each human merge, the next PR's mergeable state is worth a
glance; a conflict there stops the drain for a deliberate fix, and `bd gate check` turns the merges
into bead state next session.

### The spec's tracker issue

When the spec behind the work has a GitHub mirror, its **parent issue** walks the work's state, at
the granularity GitHub can express:

| Moment | With a Project board | Without one |
| --- | --- | --- |
| First claim | Status → **In Progress** | — |
| Work built, PRs open | Status → **In Review** | — |
| Every PR merged, epic closed | Status → **Done**, issue closed, children closed with it | Issue closed, children with it |

Requirement sub-issues close **with the parent, never individually** — beads carry design refs, not
requirement IDs, and a mirror that guesses is worse than one that is coarse. Every PR body carries
one line — `Relates to #<spec-issue>` — so the mirror cross-links the work as it happens.

Board IDs are per-installation and never stored in this skill: discover them at run time
(`gh project list`, `gh project field-list`), or read them from `CUSTOMIZE.md` when the project has
pinned them there. A pinned ID that fails means the board changed — rediscover, show the
difference, and ask before updating the file.

## The concept transition

`conceptualize` reserves one transition for this skill: `Status: Active — <date>`, meaning work has
started. At the run's first claim, resolve the concept — the design's `Related` line to the spec,
the spec's `**Concept:**` row to the concept, or the design's `concept` label when there is no
spec — and flip its Status line. Automatically, no ceremony: this records an observable fact, which
is the carve-out the repo's rules make for it. Report it in the run output — "CONCEPT-012 →
Active."

Once, idempotently. Already `Active`: nothing to do. No concept in the lineage — a tier-0 bead, a
spec with no concept: nothing to do. Only the Status line is touched, ever.

## Picking up an interrupted run

State lives in three places — beads, git, GitHub — and a resumed session reconciles them rather
than re-running anything:

| Found | Meaning | Do |
| --- | --- | --- |
| Bead closed, PR merged | Finished and landed | `bd gate check` records it; nothing else |
| Bead closed, PR open | Done, awaiting the human | Leave it; it is in the merge order |
| Bead claimed, branch pushed, no PR | Worker stopped before `gh pr create` | Verify the branch, open the PR from the root — do not re-run the work |
| Bead claimed, no branch | Work never started or never landed anywhere | Relaunch the worker with the same rendered prompt |
| Bead unclaimed but `bd ready` says ready | Never started | Normal flow |

A stacked chain resumes from its highest link with an open PR; everything below is merged or
awaiting merge, and everything above follows the normal sequence.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

### 1. Check preconditions

Run the shared check against the user's project — the script's path resolves from this skill's
directory, and the argument is the project.

```bash
"../../shared/preflight.sh" .
```

`beads=ok` advances. Otherwise read `beads_reason`, tell the user what is missing, hand over the
command that fixes it, and stop:

| `beads_reason` | What is wrong | Give them |
| --- | --- | --- |
| `not_installed` | `bd` is not on PATH | `brew install beads` |
| `not_initialized` | this repository has no beads database | `bd init` |
| `unreadable` | bd found a database and could not read it | quote `beads_detail` |

**Never run the remedy.** `bd init` writes and commits real files; that is the user's decision.

Then read the project's `AGENTS.md` (root and scoped) and `.codefall/skills/implement/CUSTOMIZE.md`
— workflow constraints, verify commands, pinned board IDs, a standing strategy preference.

### 2. Fix the scope

Per [One bead or the graph](#one-bead-or-the-graph). With no argument, run the session-start
ritual per [Beads](#session-start) — `bd ready` unfiltered, since no epic is chosen yet — then show
the ready set grouped by epic, with title, priority, and what each unblocks, and ask. Never infer a
batch from an unprompted ready set; nothing lands unrequested.

### 3. Read

The full list in [What gets read](#what-gets-read), for every bead in scope. Reconcile the design
against the code as it stands on the base branch — cite file and line for anything the design
assumed that has moved — and default to preserving whatever the design is silent about.

### 4. Classify the landing strategy

Build the graph picture (`bd ready --mol <epic> --explain`, `bd dep tree`), find the chains and any
fan-in, derive each chain's predicted file scope from its Design refs, apply the hotspot rule. Pick
per [Landing strategies](#landing-strategies), constrained by `AGENTS.md` and `CUSTOMIZE.md`.

### 5. The go gate

Present the block per [The go gate](#the-go-gate) and wait. On go: flip the concept to `Active` if
one is behind the work. Epic scope: create the epic branch if the strategy calls for one
(`epic/<id>-<slug>` off `main`, pushed), create the landed bead, claim the epic and the first wave,
`bd dolt push`. Single-bead scope: claim the bead, `bd dolt push`, nothing else.

When the user overrules the classifier the same way twice, offer to record the preference in
`CUSTOMIZE.md` — offer, never write unasked.

### 6. Execute the waves

Per wave: render worker prompts, launch the batch, wait for results. Verify each success — branch
on the remote, PR exists, or it did not happen. One automatic retry per failed bead at higher
effort; a second failure escalates. File discovered work. Comment the PR link, close the bead with
what was verified, gate the landed bead with the new PR (stacked runs), `bd dolt push`.
`--suggest-next` names the next wave; claim it and go again. Epic branch: merge each worker PR into
the epic branch, serialized, at the wave boundary.

Single-bead scope is the same loop with one iteration, run in one worktree.

### 7. Integrate

When the frontier is empty: restack stack bottoms onto current `main`, re-run verification, resolve
nothing silently. Epic branch: open the aggregate PR to `main`, titled as a release-worthy
conventional commit, and gate the landed bead with it — `Closes` nothing; the gate owns the epic's
close.

### 8. Report and stop

- Every bead built, with PR, branch, and what its close reason verified.
- The merge order, bottom-up per stack, and what is blocked on the user.
- Discovered work filed.
- The tracker mirror's state, the concept transition if one fired.
- The worktree list, with the cleanup offer.
- Final `bd dolt push`.

Do not merge. Do not wait for merges. The next session's `bd gate check` finishes the story.

## Other modes

- **Resume** an interrupted run — per
  [Picking up an interrupted run](#picking-up-an-interrupted-run). Reconcile, then continue the
  normal loop.
- **Abandon** a run: unclaim what is claimed and unbuilt, note why on each bead, report branches
  and PRs left standing. The user decides their fate; delete nothing.
- **Drain assistance is reporting only.** After merges, `bd gate check` and the mirror update are
  welcome; performing merges is not, and the hook enforces it.

## Lineage

What this took from elsewhere, and what it deliberately did not, so nobody re-adds it.

**From the dev-implement skill pair** that preceded this plugin: the worker-prompt pattern with
strategy encoded in `BASE_REF`/`PR_TARGET`, the go-block, wave execution with verified worker
results, the recovery table, and the humans-merge-`main` rule with its hook. **Dropped:**
`MAX_STACK_DEPTH = 4` — the depth cap solved a cosmetic problem and a 12-PR stack works; the
classification that sent multiple independent chains to an epic branch — chains parallelize as
stacks; the two-skill split — one skill decides scope at run time; and the board scripts — beads
replaced the project board.

**From Kiro**: the one-task-at-a-time worker discipline and read-everything-first; drift
reconciliation at resume, which became the recovery table. **From spec-kit**: parallel means
file-disjoint, nothing else. **Dropped from both:** checklist gates and the converge audit — review
is another verb.

**From beads' own docs**: close-at-done with gates carrying the merge seam. **Adapted:** gates
attach to a landed child bead, not to dependents — a stacked dependent deliberately builds on its
parent's branch, so gating it would block exactly what stacking allows — and not to the epic, which
bd will not gate. **Diverged from beads' multi-agent docs, deliberately:** workers do not self-claim
with `bd ready --claim`. That pattern arbitrates races between peer agents pulling a shared queue;
this skill's topology is dispatcher and workers, the root assigns work top-down, and the default
embedded database is single-writer besides.

## Rules

- **A human performs every merge to `main`; this skill performs none, in any mode.** The hook
  denying one is the system working.
- **Every `bd` write is the root's, in the primary checkout.** Workers never run `bd`; their prompt
  carries what they need and their result JSON carries what they found.
- **Closed means done — criteria verified, checks green, PR open.** Merged is the gates' to say,
  and the epic closes only when they drain.
- **Publish the claim before the work.** `bd dolt push` follows every claim and every close.
- **The graph is the sequencer.** `bd ready` decides what runs next; this skill keeps no schedule
  of its own and never starts a blocked bead.
- **Scope is exactly the bead.** Tangents become `discovered-from` beads, filed by the root, never
  fixed in passing.
- **Bead IDs ride every commit message.**
- **Depth never forces the epic branch; fan-in and don't-touch-main do**, and only when parallelism
  matters — the single topological stack is always correct.
- **Hotspot files are overlap until shown otherwise.** When parallel stacks cannot be shown safe,
  serialize.
- **Verify workers, never trust them.** Branch on the remote and PR open, or it did not happen.
- **One automatic retry, then a human.** Higher effort, fresh worker, failure reason in the prompt.
- **No permission prompts mid-run.** Workers cannot answer them; the go gate states the condition
  and offers single-task mode when it fails.
- **Tests are part of done.** Planned or discovered, written now, never deferred to `test`.
- **Implement never writes bead metadata and never redesigns the graph.** Metadata is read as a
  recommendation; a wrong task goes back to `design`.
- **The mirror never guesses.** The spec's parent issue carries the ladder; requirement children
  close with it, not by inference.
- **`Active` is a fact, recorded once.** Only the concept's Status line, only at first claim, only
  when a concept exists.
- **Worktrees are cleaned up by offer, never by default**, and never under an open PR.
- **Never overwrite a file that has drifted.** Show the difference and ask.
