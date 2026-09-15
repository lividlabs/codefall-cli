# The spec's tracker issue

When the spec behind the work has a GitHub mirror, its **parent issue** walks the work's state, at
the granularity GitHub can express. Read at step 6, at the first claim, and at step 8.

| Moment | With a Project board | Without one |
| --- | --- | --- |
| First claim | Status → **In Progress** | — |
| Work built, PRs open | Status → **In Review** | — |
| Every PR merged, epic closed | Status → **Done**, issue closed, children closed with it | Issue closed, children with it |

- Requirement sub-issues close **with the parent, never individually**.
- Every PR body carries one line — `Relates to #<spec-issue>` — so the mirror cross-links the work
  as it happens.
- Board IDs are per-installation and never stored in this skill: discover them at run time
  (`gh project list`, `gh project field-list`), or read them from `CUSTOMIZE.md` when the project
  has pinned them there. A pinned ID that fails means the board changed — rediscover, show the
  difference, and ask before updating the file.
