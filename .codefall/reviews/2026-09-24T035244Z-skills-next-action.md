# Review: skills-next-action

**Reviewed:** 2026-09-24T03:52:44Z · **Reviewer:** muse/muse-spark · **Revision:** 091b516376a4 (base 5b77115dd10f)
**Lenses:** correctness, behaviour, conventions, docs, simplify

## Findings

### 1. The graft next-action line drops the boundary-enforcement obligation the report just listed as owed, and 'Nothing else' denies it. — `important` · `correctness` · fixed
`extensions/skills/codefall-graft/SKILL.md:255-256`

After an adoption run, where the owed bullet names the hand-merges plus the boundary-enforcement obligation.

```diff
--- a/extensions/skills/codefall-graft/SKILL.md
+++ b/extensions/skills/codefall-graft/SKILL.md
@@
-- **Last, what the user does next**: merge the pull request, then the hand-merges above. Nothing
-  else.
+- **Last, what the user does next**: merge the pull request, then what is owed above. Nothing
+  else.
```

### 2. The implement next-action sends the user to `/codefall-test DESIGN-NNN`, an argument shape the test skill defines no target for and stops on instead of running. — `important` · `correctness` · fixed
`extensions/skills/codefall-implement/SKILL.md:326-327`

Whenever the user follows the line on a design whose bead named a case; the test skill's Targets table accepts suites, changed, a suite name, <area>/<slug>, or a case path, never a design identifier.

```diff
--- a/extensions/skills/codefall-implement/SKILL.md
+++ b/extensions/skills/codefall-implement/SKILL.md
@@
 - **Last, what the user does next**: review the pull requests, merge them in the order above, then
-  `/codefall-test DESIGN-NNN` where a bead named a case.
+  `/codefall-test <area>/<slug>` for each case a bead named, where one was named.
```

### 3. The review next-action tells the user to commit the findings files with uncommitted work, while the skill's own findings-file rule leaves them unstaged in that case. — `important` · `correctness` · fixed
`extensions/skills/codefall-review/SKILL.md:260-262`

On a review whose target is uncommitted work.

```diff
--- a/extensions/skills/codefall-review/SKILL.md
+++ b/extensions/skills/codefall-review/SKILL.md
@@
    worktree the fixes landed on if one was created. **End with what the user does next**: on a pull
-   request, merge it; on uncommitted work, commit the findings files with the work; otherwise
-   nothing is pending.
+   request, merge it; otherwise nothing is pending, and on uncommitted work the findings files stay
+   unstaged.
```

### 4. The section's last line is the old 'Do not create issues. Do not start a specification.' prohibition, not the next action the new AGENTS.md rule requires. — `minor` · `conventions` · fixed
`extensions/skills/codefall-envision/SKILL.md:238-238`



```diff
--- a/extensions/skills/codefall-envision/SKILL.md
+++ b/extensions/skills/codefall-envision/SKILL.md
@@
+Do not create issues. Do not start a specification.
+
 Report the path, the identifier, the status, every open question the document carries, the branch,
 and the pull request if one was opened. **End with what the user does next**: merge the pull
 request, then `/codefall-specify VISION-NNN`.
-
-Do not create issues. Do not start a specification.
```

### 5. The section's last line is the old 'Do not create issues. Do not start a design.' prohibition, not the next action the new AGENTS.md rule requires. — `minor` · `conventions` · fixed
`extensions/skills/codefall-mock-up/SKILL.md:332-332`



```diff
--- a/extensions/skills/codefall-mock-up/SKILL.md
+++ b/extensions/skills/codefall-mock-up/SKILL.md
@@
+Do not create issues. Do not start a design.
+
 Report the directories, every file with what it shows, what the mockup was matched against, options
 offered and which was taken, states deliberately not made and why, anything assumed because the user
 did not answer, any labels cleared, the branch, and the pull request if one was opened. **End with
 what the user does next**: merge the pull request; then `/codefall-design SPEC-NNN` where a spec was
 waiting on this mockup, otherwise `/codefall-specify`.
-
-Do not create issues. Do not start a design.
```

### 6. The section's last line is the 'Do not merge. Do not wait for merges.' directive, not the next action the new AGENTS.md rule requires. — `minor` · `conventions` · fixed
`extensions/skills/codefall-implement/SKILL.md:329-329`



```diff
--- a/extensions/skills/codefall-implement/SKILL.md
+++ b/extensions/skills/codefall-implement/SKILL.md
@@
 ### 8. Report and stop
 
+Do not merge. Do not wait for merges. The next session's `bd gate check` finishes the story.
+
 - Every bead built, with PR, branch, and what its close reason verified.
 - The merge order, bottom-up per stack, and what is blocked on the user.
 - Discovered work filed.
 - The tracker mirror's state, the vision transition if one fired.
 - The worktree list, with the cleanup offer.
 - Final `bd dolt push`.
 - **Last, what the user does next**: review the pull requests, merge them in the order above, then
   `/codefall-test DESIGN-NNN` where a bead named a case.
-
-Do not merge. Do not wait for merges. The next session's `bd gate check` finishes the story.
```

## Not checked

- Whether envision's unconditional next action (/codefall-specify VISION-NNN) is right when the vision was left Draft (user stopping and coming back) or when the project has not been scaffolded yet, since the workflow table has scaffold requiring a vision.
- Whether scaffold's unconditional next action (then /codefall-envision) is right when scaffold ran from an existing vision, which would send the user back to re-envision work already framed.
- Whether the new AGENTS.md last-line rule is meant to cover other-modes and early-exit reports (envision/specify/design other modes, review's three early exits), which gained no next-action line.
- Whether review's 'otherwise nothing is pending' holds for branch, document, and path targets where fixes landed on a branch or worktree but no push or pull request was opened.
