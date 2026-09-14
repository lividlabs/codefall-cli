---
name: codefall-review
description: Review something and fix what the user accepts — uncommitted work, a branch, an open pull request, a path, a document, or a description of what to look at. The reviewer is a subagent or another harness; this session triages the findings with the user and applies the ones they take. Every finding and what was decided about it is written to .codefall/reviews/.
argument-hint: "[what to review — nothing for uncommitted work] [via=codex|claude|opencode|gemini[:model]]"
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

# Review

Read something, say what is wrong with it, decide with the user what to do about each finding, do
it, and write down what was found and decided.

One invocation is one complete review. Nothing carries over: a second review of the same thing is a
new review, with its own file.

**The reviewer and the fixer are different contexts.** The review runs in a subagent or another
harness; the triage and the fixes happen in this session. That separation is the point — a model
that both finds a problem and fixes it grades its own work, and the second reading goes through the
same blind spots that made the first one worth doing.

Paths that start with `../`, and `scripts/` beside this file, are relative to this skill's
directory, not the user's project.

## Targets

The argument's shape decides what is being reviewed.

| Argument | Target | Resolved by |
| --- | --- | --- |
| *(none)* | uncommitted work | `git diff`, `git diff --cached`, `git status --short` for untracked files |
| a branch name | that branch | `git diff <base>...<branch>` against the default branch |
| a PR number or github.com pull URL | that pull request | `gh pr view <ref> --json state,headRefName,headRefOid,baseRefOid`, `gh pr diff <ref>` |
| a path to a file or directory | that path as it stands | the files under it |
| `CONCEPT-NNN`, `SPEC-NNN`, `DESIGN-NNN`, `ADR-NNN`, `ADR-BASE-NN`, `ADR-<PREFIX>-NN`, or a path under `docs/` | that document | the file, plus the document upstream of it |
| anything else — prose describing what to look at | that code | a search, confirmed with the user |

Resolve a bare identifier by globbing its directory — `docs/specs/SPEC-004-*.md` — and stop if it
matches nothing or more than one. Never guess at a near miss.

### What is not reviewable

Three things are refused rather than attempted, and the refusal says which:

- **A merged or closed pull request, and a merged branch.** The work has landed and the code has
  moved on. Review what is live.
- **A superseded ADR, or an archived concept or spec.** Immutable history. There is nothing a
  finding could lead to.
- **A specific commit.** Out of scope for this verb.

### A prose argument

A prose argument is a scope and sometimes a narrowing. Resolve the scope by searching — the
description names components, behaviours, or domain terms, and those map to files.

**An ambiguous scope is interviewed, not guessed.** Search, show what you found, and ask what the
search could not settle — which of two components was meant, whether the boundary includes its
callers, whether a second subsystem matching the same terms is in or out. Search again with the
answer. Repeat until the file list is one the user recognises. Each round costs a question; guessing
costs the whole review.

Do not review until the user confirms the file list. Stop if the search finds nothing — say so and
ask for a different description rather than widening on your own.

## The confirmation

Every target is confirmed before anything is reviewed. Name what was resolved and which lenses will
run, and offer to drop some:

```
Resolved SPEC-004 to docs/specs/SPEC-004-trip-sharing.md,
checked against CONCEPT-002-trip-sharing.md.

Lenses: structure, status, trace, criteria, precision, stories.

Review now, or exclude any of these?
```

Every lens that applies to the target runs unless the user drops it here. A prose argument that
narrows — "review the auth code for security issues" — shows the reduced list, so a misread
narrowing is caught before the review rather than after it.

## What gets read

**Diffs alone are not enough.** Code that looks wrong in isolation is often correct given what
surrounds it, and code that looks fine in a diff is often wrong given what it replaced.

**A target with a diff** — uncommitted work, a branch, a pull request:

1. The diff, by the command in the table above.
2. Every modified file in full. Control flow, error handling, and the patterns the file already
   follows decide whether a change fits.
3. Every untracked file in full. A new file has no diff; its whole content is the change.
4. The project's `AGENTS.md`, root and scoped, and `docs/adrs/`. These are what a convention finding
   cites. A project with neither is reviewed against its own surrounding code, and the report says
   so.
5. The design behind the work when there is one. A branch named `feat/bd-unz-…` carries a bead ID;
   `bd show <id> --json` gives `spec_id`, the design document's path. The design's `Related` line
   reaches the spec; the spec's `**Concept:**` row reaches the concept. No bead, no design: review
   against conventions alone and say so.

**A path or prose target**: every file in the resolved set, in full, plus item 4 above. There is no
diff, so there is no "what changed" — the whole of what was named is in scope.

**A document target**: the document and the document upstream of it. The question a document review
asks is whether the target faithfully refines what came before it, and what it added that nobody
asked for.

| Target | Upstream | Found by |
| --- | --- | --- |
| concept | none | — |
| spec | its concept, when it has one | the `**Concept:**` header row |
| design | its spec, or its concept when there is no spec | the `**Related:**` row |
| ADR | the design that cites it, and every other accepted ADR | `grep -l 'ADR-NNN' docs/designs/`, plus `docs/adrs/` |

Links point one way, deliberately: a design records its spec and a spec records no design. Reaching
a design from a spec or an ADR is a grep for the identifier. A concept has no upstream and is
reviewed for internal consistency alone.

## The lenses

**Code:**

| Lens | The question |
| --- | --- |
| `correctness` | Logic errors, off-by-one, wrong conditionals, missing guards, unreachable paths, null and empty and boundary inputs |
| `failures` | Errors swallowed, caught and ignored, logged and continued past, or returned and never checked |
| `behaviour` | Anything a consumer would notice that changed, especially where it looks unintentional |
| `tests` | Does the work carry the tests it needs, and do they test behaviour rather than implementation |
| `types` | Where new types appear: do they make invalid states unrepresentable, or push that job onto every caller |
| `conventions` | `AGENTS.md` and the ADRs — boundary rules first, since those are the encoded architecture |
| `comments` | Do the comments say what the code does, and does anything the work touched now lie |
| `docs` | Does the code contradict a document that describes it — `README.md`, `AGENTS.md`, a design, an ADR — and is that document one the project may amend or one that is immutable |
| `simplify` | Reuse it should have used, dead code, nesting that early returns would flatten |
| `security` | Injection, authn/authz bypass, data exposure, secrets in the diff |

**A `docs` finding is not always a documentation fix.** An ADR contradicted by the code is never a
docs edit — ADRs are immutable, and the only in-place change allowed is flipping Status to
`Superseded by <id> — <date>`. Such a finding says which it is: the code is wrong, or the decision
changed and nobody wrote the superseding ADR. A `README.md` or an `AGENTS.md` that has fallen behind
is an ordinary fix.

**`simplify` is a lens, not a separate verb.** `codefall-implement` already ran the host's own
simplify pass over each bead's diff, so a simplification finding on that code is either something
that pass missed or code implement never saw. Say which.

**Documents.** Every document target gets `structure` and `status`, plus the lenses for its kind.

| Target | Lens | The question |
| --- | --- | --- |
| all | `structure` | Required sections present; no empty headings; no template guidance left in; identifiers written in full |
| all | `status` | Status line is one of the document's own values with a real date; `Archived` is under `archive/`; a superseded ADR names one that exists |
| concept | `floor` | A problem stated, who feels it, and why now — not a solution wearing a requirement's clothes |
| concept | `scope` | What this is not, stated. No `## Non-goals` means it has not said where it stops |
| concept | `testable` | Could what it asks for be tested at all, in principle? Not how — whether |
| spec | `trace` | Does not run without a concept. Reports a spec requirement with no concept parent as **unframed**, and a concept requirement no spec requirement reaches as **unaddressed** |
| spec | `criteria` | EARS: one of the six patterns, a named subject, `IF … THEN` for failure behaviour, `SHALL` only inside criteria, no endpoints/tables/components/queues/libraries |
| spec | `precision` | Ambiguity a reader could resolve two ways; undefined terms; missing non-functional requirements |
| spec | `stories` | Every requirement has a user story with its `so that` clause |
| design | `coverage` | Every spec requirement has a home in the design, or the design says why not |
| design | `decisions` | Hard-to-reverse choices with no ADR; assumptions the spec does not guarantee; a conditional section present with nothing behind it |
| design | `plan` | The Task Plan is in exactly one of its two forms, never both |
| adr | `alternatives` | Alternatives genuinely weighed, not asserted and dismissed in a clause |
| adr | `consequences` | Consequences stated, including the ones that cost something |
| adr | `coherence` | No contradiction with another accepted ADR; project decisions numbered bare `ADR-NNN` rather than continuing an inherited sequence |

**Neither half of `trace` is a defect.** A spec requirement with no concept parent is often
legitimate — a precondition the concept never anticipated. An unaddressed concept requirement may
mean the concept should catch up. The finding asks the question; it does not assert the spec is
wrong.

**An ADR edited after ratification is itself a serious finding.**

## Calibration

The difference between a review worth reading and one worth skipping is what it declines to say.

- **Be certain before calling something a bug.** Investigate. If still unsure, say so — that is
  useful, and it is not a finding.
- **Review the target, nothing else.** Where the target has a diff, the scope is the changed lines,
  and pre-existing code the diff did not touch is out of bounds. Where it does not — a path, a
  document — the scope is the whole of what was named. Either way, a finding about something the
  target does not cover is out of scope however true it is.
- **No hypothetical edge cases.** Name the realistic scenario that reaches it, or drop it.
- **State the conditions up front.** Empty input, concurrency, cold cache — the first sentence says
  when it matters. The conditions are most of the severity.
- **Do not be a zealot about style.** Verify the project actually holds the convention. Some
  violations are the simplest option and are fine. Excessive nesting is a finding regardless.
- **Do not overstate severity or inflate the count.** Ten findings where three are real makes the
  three harder to act on.
- **No flattery.** No summary of what the work does well.

Severity: a **blocker** is wrong and will be observed; **important** is wrong under conditions that
will occur; **minor** is worth fixing and costs nothing to leave. A `REVIEW.md` at the project root,
if present, states what the project cares about and wins.

## Who reviews

Three options, and `via=` picks the third:

1. **A subagent of this harness.** The default. Separates the reviewer from the fixer at no cost.
2. **This session.** The reviewer is then also the fixer, which is acceptable only when the work
   under review came from somewhere else. Say so in the report.
3. **Another harness**, named by `via=`. A different model has different blind spots, which is the
   reason to reach for one.

```
via=codex            via=codex:gpt-5-codex
via=claude           via=claude:claude-opus-5
via=opencode         via=opencode:anthropic/claude-sonnet-5
via=gemini           via=gemini:gemini-3-pro
```

### Lens groups

The review runs as **four subagents in parallel**, each reading the material once and asking one
coherent set of questions. Several passes each looking for one thing find more than one pass looking
for everything, and four is where that stops paying for itself.

| Group | Lenses |
| --- | --- |
| 1 | `correctness`, `failures`, `behaviour` |
| 2 | `tests`, `types` |
| 3 | `conventions`, `comments`, `docs`, `simplify` |
| 4 | `security` |

A document target runs two groups: the mechanical checks — `structure` and `status` — and the
reading, which is everything else for that kind.

A group whose lenses were all dropped at the confirmation does not run. Each returns JSON against
`findings.schema.json`; this session merges them into one findings array.

### Another harness

`scripts/review-via.sh [--model <model>] <harness> <prompt-file> <schema-file> <out-file>` runs the
harness's headless read-only mode in the repository, so the reviewer reads the files itself. Leaving
`--model` out is how the harness's own default is taken — which is what `via=codex` with no model
means. The
prompt file carries the target, the lenses, the calibration rules, and the schema. Codex and Claude
Code also take the schema as a flag — `--output-schema` and `--json-schema` — which makes their
output conform by construction rather than by request.

- The external reviewer runs read-only. It proposes; it never edits.
- A failure — missing CLI, auth error, timeout, non-zero exit — is reported with the harness name
  and its output, with an offer to review with this session instead. Never fall back silently.

## Where the fixes go

**The target decides, not where the user is standing.** Reviewing PR #51 from a feature branch puts
the fixes on #51's branch, not the one you happen to be on.

| Target | Fixes land on |
| --- | --- |
| Uncommitted work | The working tree, in place |
| A branch | That branch |
| An open pull request | That PR's branch |
| A document or path, when something is already checked out for it | There |
| A document or path on the default branch | A new worktree, branched from the default branch |

The last row is the only case that creates anything, and it creates it because there is nothing to
land on — not because of where the session started. A worktree unless the user asks for a plain
branch.

Uncommitted work is never moved. A new worktree cannot contain the changes in this one, so those
fixes are applied where the changes are.

## Triage and fixes

Present the findings as one numbered list, most severe first, each with its location, its claim, and
the conditions under which it matters. Then ask which to fix.

Every finding leaves triage with a status:

| Status | Meaning |
| --- | --- |
| `fixed` | The user took it; the fix is applied in this invocation |
| `dismissed` | The user rejected it, with a reason |
| `deferred` | Real, but not now |

Apply the accepted fixes with this session's tools, code and documents alike. A finding's proposed
`patch` is a starting point, not a script — apply the intent, matching the surrounding code. Fixes
are not re-reviewed here; run the review again if that matters.

## The findings file

Two files per invocation, sharing one stem, both committed:

```
.codefall/reviews/<YYYY-MM-DDTHHMMSSZ>-<target-key>.json
.codefall/reviews/<YYYY-MM-DDTHHMMSSZ>-<target-key>.md
```

`<target-key>` carries a slug wherever there is one to take: `pr-51-init-below-root` from a PR
title, `SPEC-004-trip-sharing` from a document, `feat-bd-unz-stage-context` from a branch,
`src-fulfillment` from a path, the resolved common root for a prose target or `adhoc` when the files
share none, and `uncommitted` when there is no subject at all.

The `.json` is the record; the `.md` is the same review written to be read. Both hold what was
reviewed and at which revision, who reviewed it, which lenses ran, what could not be checked, and
every finding with its status. `findings.schema.json` beside this file is the shape.

**The files are written three times** — after the review, after triage, after the fixes. An
interrupted session resumes from them rather than starting over.

**`revision` is the target's, not the session's.** Standing on the default branch reviewing a pull
request, HEAD is the wrong answer.

| Target | `revision` |
| --- | --- |
| uncommitted | HEAD SHA, plus `dirty` — a short hash of `git diff HEAD` |
| branch | the branch's tip SHA, plus `base` — the merge-base with the default branch |
| pull request | `headRefOid`, plus `base` — `baseRefOid` |
| document | the SHA of the last commit that touched the file; if it is modified in the working tree, HEAD plus a `dirty` hash |
| path or prose | HEAD SHA of the checkout the review ran in, plus `dirty` when the tree is not clean |

### Kept out of codebase search

A `.ignore` file beside `.codefall/` holds one line:

```
.codefall/reviews/
```

Ripgrep honours `.ignore` and git does not, so the directory is committed while every harness that
searches through ripgrep skips it. `git grep`, `find`, and `cat` still reach the directory, which is
fine — the objective is keeping old findings out of unrelated searches, not secrecy.

`codefall init` writes the line, appending to an existing `.ignore` rather than replacing it, and
`codefall doctor` warns when it is missing rather than failing: nothing stops working without it.

**Check it before the review, and offer.** The line is easy to lose — a project set up before this
verb existed never got one, and a person editing `.ignore` can drop it. This skill is where that
costs something, because it is about to write findings into a directory nothing is hiding, so it is
where the offer belongs:

> `.ignore` doesn't list `.codefall/reviews/`, so findings from this review will show up in
> codebase searches. Add the line?

On yes, append it — the file may hold entries that are the project's own, so append, never replace.
On no, carry on and say nothing further; it is one line in a file that belongs to the user. Say
nothing at all when the line is already there.

## Posting to a pull request

Off unless the project turned it on. `.codefall/settings.json` carries it:

```json
{ "review": { "postToPullRequest": false } }
```

When it is `true` and the target is an open pull request, the findings post after triage as **one
review** — a single `POST /repos/{owner}/{repo}/pulls/{n}/reviews` with `event=COMMENT` and a
`comments` array, so each finding becomes an inline thread on the line it is about.

GitHub rejects an inline comment on a line outside the diff, and rejects the whole call if any
comment is out of bounds. So a finding that anchors inside a diff hunk becomes an inline thread, and
everything else — a whole-file concern, a missing test, a `docs` finding about another file — goes
in the review body. Never post a finding the user dismissed.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

1. **Resolve and confirm.** Resolve the target per [Targets](#targets); refuse what is not
   reviewable. Interview a prose scope until the file list is recognised. Read
   `.codefall/settings.json`'s `review` block and `.codefall/skills/codefall-review/CUSTOMIZE.md`
   if present. Check the `.ignore` entry and offer to add it if it is missing, per
   [Kept out of codebase search](#kept-out-of-codebase-search). Present
   [the confirmation](#the-confirmation) and wait.
2. **Read.** Everything in [What gets read](#what-gets-read).
3. **Review.** Four subagents in parallel by lens group, or `scripts/review-via.sh` when `via=` was
   given. Every candidate finding checked against [Calibration](#calibration) before it becomes one.
4. **Write.** Merge the groups' JSON and write the findings files, every finding `open`.
5. **Triage.** Present the list, take a decision on each, update the files. Post to the pull request
   if that is enabled and the target is one.
6. **Fix.** Apply what was taken, in the place [Where the fixes go](#where-the-fixes-go) names.
   Update the files.
7. **Report.** The target and reviewer; what was found, most severe first; what was fixed,
   dismissed, deferred; what could not be checked and why; where the files are; and the branch or
   worktree the fixes landed on if one was created.

## Rules

- **The reviewer never fixes.** Review in a subagent or another harness; triage and fix in this
  session.
- **Nothing is changed before triage.** The user decides what gets fixed.
- **Confirm the scope before reviewing it.** A review of the wrong files costs the whole run.
- **Be certain, or say you are not.**
- **Review the target, nothing else.**
- **Every finding carries the conditions under which it manifests.**
- **No flattery, no filler findings.**
- **Read the whole file, never only the hunk.**
- **An external reviewer runs read-only.** A failure is reported, never worked around.
- **The target decides where fixes land**, never where the session started.
- **Uncommitted work is never moved to a worktree.**
- **Every finding ends with a status**, and the files are written even when nothing was fixed.
- **Refuse what is not reviewable** — a merged or closed pull request, a merged branch, a superseded
  ADR, an archived document, a specific commit — and say which.
- **An identifier that resolves to nothing is a stop**, not a guess.
- **Never post a dismissed finding to a pull request.**
- **The `.ignore` entry is offered, never added unasked**, and appended rather than written over.
