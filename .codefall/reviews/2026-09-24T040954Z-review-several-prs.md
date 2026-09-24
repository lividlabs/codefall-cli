# Review: review-several-prs

**Reviewed:** 2026-09-24T04:09:54Z · **Reviewer:** muse/muse-spark · **Revision:** 3a0f9122a555 (base 069c90f5bb4c)
**Lenses:** correctness, behaviour, conventions, docs, simplify

## Findings

### 1. The schema's revision description omits the range target even though kind, key, base, and the findings-file table all define it. — `minor` · `correctness` · fixed
`extensions/skills/codefall-review/findings.schema.json:29-31`



```diff
--- a/extensions/skills/codefall-review/findings.schema.json
+++ b/extensions/skills/codefall-review/findings.schema.json
@@
-          "description": "The state reviewed, and the target's own rather than the session's: branch tip for a branch, headRefOid for a pull request, the last commit that touched a document, HEAD otherwise."
+          "description": "The state reviewed, and the target's own rather than the session's: branch tip for a branch, headRefOid for a pull request, <to> for a range, the last commit that touched a document, HEAD otherwise."
```

### 2. The fix-landing table adds a range row but Getting there still only tells the session how to fetch and check out branch and pull-request targets. — `minor` · `correctness` · fixed
`extensions/skills/codefall-review/SKILL.md:200-202`

When the range tip is not already checked out, which is the normal case when reviewing a stack from the default branch.

```diff
--- a/extensions/skills/codefall-review/SKILL.md
+++ b/extensions/skills/codefall-review/SKILL.md
@@
-**Getting there.** A branch or PR target that is not already checked out is fetched and checked out
-before any fix is applied — `git fetch origin` then `git checkout <branch>`, taking the branch name
-from `headRefName` for a pull request. A new worktree is `git worktree add` off the default branch.
+**Getting there.** A branch, PR, or range target that is not already checked out is fetched and checked out
+before any fix is applied — `git fetch origin` then `git checkout <branch>`, taking the branch name
+from `headRefName` for a pull request and the branch whose tip is `<to>` for a range. A new worktree is `git worktree add` off the default branch.
```

### 3. The stack-as-range form never says which commit <from> is, so it stops matching the top-branch form once the default branch has moved. — `important` · `correctness` · fixed
`extensions/skills/codefall-review/SKILL.md:66-69`

When the default branch advanced after the stack branched and the reviewer passes its tip as <from>; the two-dot diff then includes unrelated default-branch commits that the top branch's three-dot diff excludes.

```diff
--- a/extensions/skills/codefall-review/SKILL.md
+++ b/extensions/skills/codefall-review/SKILL.md
@@
-**Several pull requests** — `codefall-implement` leaves one per task. A stack, where each pull
-request is based on the one below it, is one target: the top branch, or the range from the stack's
-base to its tip, holds every pull request's diff. Pull requests against the default branch share
-nothing and are one invocation each, in the order the implement report listed them.
+**Several pull requests** — `codefall-implement` leaves one per task. A stack, where each pull
+request is based on the one below it, is one target: the top branch, or the range from the stack's
+base — the merge-base of its tip with the default branch — to its tip, holds every pull request's diff. Pull requests against the default branch share
+nothing and are one invocation each, in the order the implement report listed them.
```

### 4. The chain table's review row still enumerates uncommitted work, branch, PR, path, and document with no range, as do its installed copy in extensions/shared/workflow.md and the Reviews section and skills table in README.md (all amendable project documents, the shared copy via workflow-sync from docs/workflow.md). — `minor` · `docs` · fixed
`docs/workflow.md:51-51`



```diff
--- a/docs/workflow.md
+++ b/docs/workflow.md
@@
-|| `review` | anything live: uncommitted work, a branch, a PR, a path, a document | `.codefall/reviews/`, a JSON and Markdown pair per review; fixes on the target's branch for the findings the user takes | the human |
+|| `review` | anything live: uncommitted work, a branch, a PR, a commit range, a path, a document | `.codefall/reviews/`, a JSON and Markdown pair per review; fixes on the target's branch for the findings the user takes | the human |
```

## Not checked

- Whether a range exactly one commit apart coherently escapes the single-commit refusal, given the range fix rule requires <to> to be a live branch tip while a bare commit need not be one; no session evidence was examined.
- Whether a range review covering stacked pull requests should post anything under reference/posting.md, which only describes the open-pull-request target; the skill is silent.
- Full extensions/scripts/skill-health.sh --strict was not executed; line and token counts were estimated from body word and character counts and the schema was validated by JSON parse only.
