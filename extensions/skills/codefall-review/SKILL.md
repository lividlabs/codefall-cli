---
name: codefall-review
description: Review a target and say what is wrong with it — a commit, a branch, a pull request, or uncommitted work; or a concept, spec, design, or ADR checked against the document upstream of it — reviewed by this session's model or by another harness, with every finding written to .codefall/reviews/. Changes nothing.
argument-hint: "[a commit, branch, PR, doc identifier, or nothing for uncommitted work] [via=codex|claude|opencode|gemini[:model]] [focus=<lens,…>]"
disable-model-invocation: true
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Write
  - Bash
---

# Review

Read a target, judge it, and write down what is wrong with it.

The output is **findings** — under `.codefall/reviews/`, and in the report this skill ends with.
A finding says what is wrong, where, under what conditions it matters, and often how it could be
fixed. What it never does is fix it.

**Reviewing is not fixing.** This skill changes nothing it reviews: not a line of code, not a
character of a document, not a bead, not a tracker issue. A reviewer who edits the work has stopped
being a reviewer, and the person who owns the work loses the choice that makes a review worth
having. The one thing it writes is its own findings, under `.codefall/reviews/` and nowhere else.

A finding may carry a **proposed diff**. That is a recommendation, the same as a suggested change on
a pull request: written by the reviewer, applied by whoever owns the code, or not.

Plugin paths in this document — the ones that start with `../`, and `scripts/` beside this file —
are relative to this skill's directory, the one holding this `SKILL.md`. Resolve them from where the
file lives; they are not relative to the user's project.

## Scope — judge, never change

| In scope | Out of scope | Whose |
| --- | --- | --- |
| Saying what is wrong with a target | Making it right | the user, or `codefall-implement` |
| A proposed diff, as a recommendation | Applying that diff | the user |
| Code: a commit, branch, PR, or uncommitted work | Regression campaigns, retesting in a fresh context | `test` |
| Docs: a concept, spec, design, or ADR | Writing or revising any of them | the owning verb |
| Findings under `.codefall/reviews/` | Beads, tracker issues, the decision log | `codefall-design`, `codefall-implement` |
| A verdict on work someone else built | A verdict on work this session built | nobody — see below |

The bottom row is the reason this verb exists separately. `codefall-implement` runs the harness's own
`simplify`, `code-review`, and `security-review` passes over each bead's diff, and its own rules say
those "are checks inside implement, not review — implement renders no verdict on its own work, and
independent review stays the `review` verb's." Independence is the product. A reviewer that also
fixes what it found grades its own work on the next round, and the second reading goes through the
same blind spots that made the first one worth doing.

**Filing a finding as a bead is somebody else's call.** The findings file is the record; turning one
into work is a decision about the graph, and the graph belongs to `codefall-design`. Report what is
there and stop.

## Targets

The argument decides the target, and its shape decides the kind. Codefall's identifiers are
distinctive enough to read directly, so nothing is asked that the argument already answered.

| Argument | Target | Resolved by |
| --- | --- | --- |
| *(none)* | uncommitted work | `git diff`, `git diff --cached`, `git status --short` for untracked files |
| a SHA, long or short | that commit | `git show <sha>` |
| a branch name | this branch against it | `git diff <branch>...HEAD` |
| a PR number, or a github.com pull URL | that pull request | `gh pr view <ref>`, `gh pr diff <ref>` |
| `CONCEPT-NNN`, or a path under `docs/concepts/` | that concept | the file |
| `SPEC-NNN` | that spec | the file |
| `DESIGN-NNN` | that design | the file |
| `ADR-NNN`, `ADR-BASE-NN`, `ADR-<PREFIX>-NN` | that ADR | the file |

Resolve a bare identifier by globbing its directory — `docs/specs/SPEC-004-*.md` — and stop if it
matches nothing or more than one. **Never guess at a near miss.** An identifier that resolves to no
file is the user's typo or a document that was never written, and both are worth saying out loud.

An argument that is genuinely ambiguous — a branch named like a SHA, a path that could be either —
is the one case worth a question. Ask which was meant and take the answer.

`focus=<lens,…>` narrows the review to the named lenses. `via=` chooses the reviewer; see
[Who reviews](#who-reviews).

## What gets read

**Diffs alone are not enough.** Code that looks wrong in isolation is often correct given what
surrounds it, and code that looks fine in a diff is often wrong given what it replaced. Read the
whole of what changed before judging any of it.

### For a code target

1. **The diff**, by the command in the target table.
2. **Every modified file in full** — not the hunks, the file. Control flow, error handling, and the
   patterns the file already follows are what decide whether a change fits.
3. **Every untracked file in full.** A new file has no diff; its whole content is the change.
4. **The project's `AGENTS.md`**, root and scoped, and `docs/adrs/`. These are the conventions a
   finding about convention has to cite. A project with neither is reviewed against itself — the
   patterns in the surrounding code — and the review says that is what it did.
5. **The design behind the work, when there is one.** A branch named `feat/bd-unz-…` carries a bead
   ID; `bd show <id> --json` gives its `spec_id`, which `codefall-design` set to the design
   document's path. From there the design's `Related` line reaches the spec, and the spec's
   `**Concept:**` row reaches the concept. No bead, no design: review against conventions alone and
   say so.

### For a document target

The target, and **the document upstream of it**. The upstream is what makes a document review
possible at all: on its own, a specification can only be checked for internal consistency, and the
question worth asking is whether it faithfully refines what came before it and what it added that
nobody asked for.

| Target | Upstream | How it is found |
| --- | --- | --- |
| concept | none | — |
| spec | its concept | the `**Concept:**` header row; absent when no concept framed the work |
| design | its spec, or its concept when there is no spec | the `**Related:**` row's `spec:` or `concept:` label |
| ADR | the design that cites it, and every other accepted ADR | `grep -l 'ADR-NNN' docs/designs/`, plus all of `docs/adrs/` |

**The links only point one way, deliberately.** A design records `spec: SPEC-004`; the spec records
no design, because `codefall-design`'s rule is to never record a hop it can derive. An ADR records
nothing upstream at all. So reaching a design from a spec, or a design from an ADR, is a grep for
the identifier — which is exactly what the full-identifier convention exists to make work.

A concept has no upstream and is reviewed for internal consistency alone. That makes its lens set
shorter, not weaker.

## The lenses

One question per pass. A reviewer looking for everything at once finds less than several passes each
looking for one thing, which is why the decomposition is written down rather than left to judgment.

`focus=` runs a subset. With no `focus=`, every lens that applies to the target runs.

### Code lenses

| Lens | The question |
| --- | --- |
| `correctness` | Logic errors, off-by-one, wrong conditionals, missing guards, unreachable paths, null and empty and boundary inputs |
| `failures` | Errors swallowed, caught and ignored, logged and continued past, or returned and never checked |
| `tests` | Does the change carry the tests it needs, and do they test behaviour rather than implementation |
| `types` | Where new types appear: do they make invalid states unrepresentable, or push that job onto every caller |
| `comments` | Do the comments say what the code does, and does anything the change touched now lie |
| `conventions` | The project's `AGENTS.md` and its ADRs — the boundary rules first, since those are the encoded architecture |
| `simplify` | Reuse it should have used, dead code, nesting that early returns would flatten |
| `behaviour` | Anything a consumer would notice that changed, especially where the change looks unintentional |
| `security` | Injection, authentication and authorization bypass, data exposure, secrets in the diff |

**`simplify` is a lens here, not a separate verb.** Harnesses that ship one merged it into their own
review for the same reason, and the harnesses that never had one get it this way. Note what
`codefall-implement` already did, though: it ran the host's `simplify` over each bead's diff before
the PR existed. A simplification finding on that code is either something that pass missed or a
target implement never saw — a commit from elsewhere, a PR from another person. Say which.

### Document lenses

Every document target also gets `structure` and `status`, which are the mechanical checks, and the
lens named for its own content, which is the reading.

| Target | Lens | The question |
| --- | --- | --- |
| all | `structure` | The required sections are present; no section is an empty heading; no template guidance survived into the emitted document; identifiers are written in full |
| all | `status` | The Status line is one of the document's own values with a real date; an `Archived` document is under `archive/`; a superseded ADR names one that exists |
| concept | `floor` | A problem stated, who feels it, and why now — not a solution wearing a requirement's clothes. `## Problem` and `## Proposed shape` are both mandatory |
| concept | `scope` | What this is not, stated. A concept with no `## Non-goals` has not said where it stops |
| concept | `testable` | Could what the concept asks for be tested at all, in principle? Not how — `codefall-specify` writes the criteria — but whether anything here could ever be shown to hold |
| spec | `trace` | Every requirement traces to something the concept asked for; every concept requirement reaches a spec requirement or is named out of scope |
| spec | `criteria` | EARS: one of the six patterns, a named subject in every criterion, `IF … THEN` for failure behaviour rather than `WHEN`, `SHALL` only inside criteria, and no endpoints, tables, components, queues, or libraries anywhere in them |
| spec | `precision` | Ambiguity a reader could resolve two ways; terms used but never defined; non-functional requirements the work needs and nobody wrote down |
| spec | `stories` | Every requirement has a user story, and the `so that` clause is present — it is mandatory, and it is the clause that gets dropped |
| design | `coverage` | Every spec requirement has a home in the design, or the design says why it does not |
| design | `decisions` | Choices that are hard to reverse and got no ADR; assumptions the spec does not actually guarantee; a conditional section present with nothing behind it |
| design | `plan` | The Task Plan is in exactly one of its two forms — staged table, or the collapsed mapping line — and never both |
| adr | `alternatives` | Alternatives genuinely weighed in Context, not asserted and dismissed in a clause |
| adr | `consequences` | Consequences stated, including the ones that cost something |
| adr | `coherence` | No contradiction with another accepted ADR; a project decision numbered bare `ADR-NNN` rather than continuing an inherited sequence |

**An ADR that was edited after ratification is itself a finding**, and a serious one. A ratified ADR
is never rewritten — the only in-place edit the rules allow is flipping its Status line to
`Superseded by <id> — <date>`. A revision that landed as an edit rather than a successor has erased
the decision it replaced.

## Calibration

The difference between a review worth reading and one worth skipping is what it declines to say.

- **Be certain before calling something a bug.** If you are not sure, investigate. If you still are
  not sure, write that you are not sure — it is a useful thing to say, and it is not a finding.
- **Only review what changed.** Pre-existing code that the diff did not touch is not in scope, however
  much it deserves attention.
- **No hypothetical edge cases.** An edge case is a finding when there is a realistic scenario that
  reaches it. Name the scenario or drop the finding.
- **State the conditions up front.** A finding that only matters on an empty input, or under
  concurrency, or on a cold cache, says so in its first sentence. The conditions are most of the
  severity.
- **Do not be a zealot about style.** Verify the code is actually in violation of a convention the
  project actually holds. Some violations are the simplest available option and are fine. Excessive
  nesting is a legitimate finding regardless.
- **Do not overstate severity**, and do not inflate the count. Ten findings where three are real
  makes the three harder to act on.
- **No flattery.** No "great job", no "thanks for", no summary of what the change does well. The
  report is what is wrong with it.

Severity is `blocker`, `important`, or `minor`. A **blocker** is wrong and will be observed. An
**important** finding is wrong under conditions that will occur. A **minor** finding is worth fixing
and costs nothing to leave. If the project has a `REVIEW.md` at its root, read it and calibrate to
what it says — it is the project's own statement of what it cares about, and it wins.

## Who reviews

The reviewer is this session's model, or another harness running headless.

**A second model is the point of `via=`.** The value of an outside reviewer is that it has different
blind spots, and that value disappears when the same model reviews its own output.

**There is no default.** With no `via=`, ask who should review: this session, or one of the external
harnesses. Offer only the ones actually on `PATH` — an option that would be refused is not an option
— and check with `command -v` rather than assuming.

```
via=codex            via=codex:gpt-5-codex
via=claude           via=claude:claude-opus-5
via=opencode         via=opencode:anthropic/claude-sonnet-5
via=gemini           via=gemini:gemini-3-pro
```

An external reviewer runs through `scripts/review-via.sh`, beside this file. The skill writes a
self-contained **review packet** — the diff or document, the full files, the conventions, the
upstream artifact, the lenses, and the findings schema — and the shim runs the harness's headless
mode against it and returns findings in the same shape the host would have produced.

Three rules govern that boundary, and none of them is negotiable:

- **The external reviewer runs read-only.** Every harness the shim supports has a read-only or
  planning mode, and the shim uses it. A reviewer is not permitted to edit, and a subprocess editing
  the tree would also bypass the host's `PreToolUse` hooks and its checkpoints, so nothing it did
  would be guarded or reversible.
- **The external reviewer returns findings, never applied changes.** It may propose a diff. So may
  this session. Neither applies one.
- **A shim that fails is reported, not worked around.** A missing CLI, an authentication failure, a
  timeout, a non-zero exit: say which harness failed and what it said, and offer to review with this
  session instead. Never quietly fall back.

## Findings

### Where they live

```
.codefall/reviews/<target-key>/round-NNN/
  findings.json    # the record: what a later round reconciles against, what a gate could check
  summary.md       # the reading: what the report says, in prose
```

`<target-key>` is the commit SHA, the branch name slugged, `pr-<number>`, `uncommitted`, or the
document path slugged. Rounds are numbered from `001` so a second review of the same target
reconciles against the first rather than duplicating it.

**These are committed.** Findings persist so that patterns across reviews are visible — the same
class of mistake, three reviews running, is worth more than any one of the three.

**They are hidden from codebase search** by a `.ignore` file beside the `.codefall/` directory:

```
# .ignore — read by ripgrep, not by git
.codefall/reviews/
```

Ripgrep honours `.ignore` and git does not read it, so the directory is tracked and committed while
every harness that searches through ripgrep skips it. `codefall init` writes that file, in the
directory it installed into — the repository root for most projects, and the subdirectory itself
where codefall was installed below the root. Ripgrep reads `.ignore` files hierarchically, so one
beside `.codefall/` covers it either way. The hiding is not airtight — `git grep`, `find`, and `cat`
still reach it — and it does not need to be: it keeps old findings out of unrelated searches, which
is the whole objective.

Reviews of uncommitted work live in the worktree that holds that work, which is what keeps a finding
anchored to the file state it was written against.

### The shape of a finding

| Field | What it holds |
| --- | --- |
| `id` | Stable within the target: `<lens>-<nn>` |
| `lens` | Which pass produced it |
| `severity` | `blocker`, `important`, or `minor` |
| `location` | File and line range for code; document path and heading for a document |
| `claim` | What is wrong, in one sentence |
| `conditions` | When it manifests. Empty means always |
| `patch` | A proposed diff. Optional, and a recommendation only |
| `status` | `open`, `resolved`, `stale`, or `dismissed` |

`findings.schema.json`, beside this file, is that shape as JSON Schema. It is what the shim hands an
external harness so its output arrives in the same form rather than as prose to be parsed.

### Rounds and staleness

A re-review reads the previous round before it starts, and reconciles rather than restating:

| Found | Status |
| --- | --- |
| The claim still holds at the same location | `open`, carried forward with its original `id` |
| What the claim named is no longer true | `resolved` |
| The anchored span has changed and the claim can no longer be checked against it | `stale` |
| The user marked it so by hand | `dismissed`, and never revived |

**Never revive a dismissed finding.** Somebody looked at it and decided. Re-raising it as a new
finding on the next round is the same argument with a new number.

## Posting to a pull request

Off unless the project turned it on. `.codefall/settings.json` carries it:

```json
{ "review": { "postToPullRequest": false } }
```

When it is `true` and the target is a pull request, the summary posts as **one comment** —
`gh pr comment <ref> --body-file <summary.md>` — and the report says it did. One comment, never a
review and never inline threads: a comment is visible, undoable, and does not block anyone's merge.

When it is `false` or absent, nothing is posted and nothing is asked. The findings are on disk and in
the report, and posting them is the user's to do.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

### 1. Resolve the target

Per [Targets](#targets). Infer from the argument's shape; ask only where it is genuinely ambiguous.
Name what you resolved — the file, the SHA and its subject, the PR and its title — so the user can
redirect before anything is read.

Stop if the target does not exist. Stop if a code target's diff is empty: there is nothing to review,
and saying so is the correct outcome.

Read `.codefall/settings.json` for the `review` block, and `.codefall/skills/codefall-review/CUSTOMIZE.md`
if it is there.

### 2. Read

The full list in [What gets read](#what-gets-read). Read before judging, and read all of it — the
temptation to review the diff alone is the failure this step exists to prevent.

### 3. Choose the reviewer

Per [Who reviews](#who-reviews). `via=` answers it; with no `via=`, ask, offering only the harnesses
on `PATH` alongside this session.

### 4. Review

Run each lens that applies to the target, one pass per lens, per [The lenses](#the-lenses).
`focus=` narrows the set. Apply [Calibration](#calibration) to every candidate finding before it
becomes one.

For an external reviewer: write the packet, run `scripts/review-via.sh`, and take its findings. A
failure there is reported, not routed around.

### 5. Write the findings

Read the previous round for this target if there is one, reconcile per
[Rounds and staleness](#rounds-and-staleness), and write `findings.json` and `summary.md` into the
next round directory. Add the `.ignore` entry beside `.codefall/` if `codefall init` has not.

This is the only write this skill performs. Nothing under `.codefall/reviews/` is the reviewed
target, and nothing outside it is written at all.

### 6. Report

- The target, and who reviewed it.
- The findings, most severe first, each with its location and the conditions under which it matters.
- What the previous round said, when there was one: resolved, still open, gone stale.
- Where the findings were written.
- Whether anything was posted to a pull request.
- What could not be checked, and why — no design behind the branch, no upstream document, a lens
  skipped for lack of context. A gap named is worth more than a gap papered over.

Then stop. Do not offer to fix anything, and do not fix anything if asked in the same breath — that
is `codefall-implement`'s work, or the user's, and it starts with a fresh invocation.

## Rules

- **Nothing in the reviewed target is ever written.** Not code, not a document, not a bead, not a
  tracker issue. The only writes are `.codefall/reviews/**` and the `.ignore` beside it.
- **A proposed diff is a recommendation.** Write it into the finding; never apply it, and never
  offer to.
- **Independence is the product.** The reviewer never grades work it produced.
- **Be certain, or say you are not.** Uncertainty stated is useful; uncertainty dressed as a finding
  is not.
- **Only what changed** — for code targets, the diff's lines and nothing else.
- **Every finding carries the conditions under which it manifests.**
- **No flattery, and no filler findings** to make a review look thorough.
- **One pass per lens.** The decomposition is the method, not a formatting choice.
- **Read the whole file, never only the hunk.**
- **An external reviewer runs read-only and returns findings.** A shim failure is reported, never
  worked around.
- **There is no default reviewer.** With no `via=`, ask, and offer only what is installed.
- **A dismissed finding stays dismissed.**
- **Posting to a pull request is off unless the project turned it on**, and is one comment when it is
  on.
- **An identifier that resolves to nothing is a stop**, not an invitation to guess at the nearest
  match.

## Lineage

What this took from elsewhere, and what it deliberately did not.

**From OpenCode's built-in review**: target inference from the argument's shape, the rule that diffs
alone are not enough, and most of [Calibration](#calibration) — be certain, only review changed
lines, no hypothetical edge cases, do not be a zealot about style, state the conditions, no flattery.
That template is the best short statement of review discipline in the open, and there was no reason
to write a worse one. **Added:** document targets, the lens decomposition, and structured findings,
none of which it has.

**From Anthropic's `pr-review-toolkit`**: one lens per pass rather than one reviewer looking for
everything, and the specific lenses its agents cover — silent failures, test coverage, type design,
comment accuracy, simplification. **Dropped:** the aggregate "Strengths" and "Positive Observations"
sections, which are flattery with a heading.

**From the cross-model review tools** — Codex-as-second-opinion wrappers and the loop harnesses:
shelling out to another harness's headless mode rather than trying to switch providers inside the
host, because no harness offers a per-subagent provider switch and a session-wide base URL override
changes the host model too. **Diverged, deliberately:** the external reviewer never edits, even where
its CLI can. A subprocess writing files bypasses the `PreToolUse` hooks that are codefall's
enforcement and the checkpoints that make a session reversible, so a patch it produced would be
neither guarded nor undoable.

**From `codefall-graft`**: the report as the deliverable, and the discipline of stopping after it.

**Dropped from the original design of this verb:** applying fixes for code targets and inline
`> [!REVIEW]` callouts for document targets. Both were changes to the reviewed work, and a reviewer
that changes the work is not one.
