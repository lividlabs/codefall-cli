# codefall-extension — operative rules

Terse on purpose: the why lives in the linked docs. This is the same convention the extension
installs into scaffolded projects — an `AGENTS.md` holds the operative rules and links back, and
never restates the reasoning.

## ADRs are history

- A ratified ADR is never rewritten — not in a user's project, not in this repo. A revision lands
  as a **new, superseding ADR**; the only in-place edit ever made to an existing ADR is flipping
  its Status line to `Superseded by <id> — <date>`.
- Successor naming: the current template's name when a rename is involved, an edition suffix
  (`ADR-BASE-02.2`) when it isn't. Mechanics: `skills/<verb>/SKILL.md`, step 5.
- This binds every verb that touches ADRs — `codefall-graft` today, `codefall-design` and the rest as they land.

## Skills

- The rules for writing and changing a skill — shape, length, what goes where, supporting files,
  templates — live in [`skills/AGENTS.md`](skills/AGENTS.md). Length is governed by
  [ADR-004](../docs/adrs/ADR-004-skill-length-guidelines.md).

## Hooks

- `hooks/shared/` holds the script every harness's guard runs, and `hooks/<harness>/` holds the
  definition `codefall init` registers — never copied as files. Hooks are the one place a skill's
  relative-path rule does not apply.
- **Every script in `hooks/shared/` is named `codefall-<what-it-does>.sh`.** The prefix is not
  decoration: `init` recognises its own registration in a file it shares with the project's hooks by
  the script a command names, and replaces that entry on an upgrade rather than appending beside it.
  A script renamed out of the prefix would leave every project's old registration running for ever.

## Prose

- Banned jargon, everywhere — docs, skills, ADRs, commit messages, PR bodies: *arm*,
  *load-bearing*, *honest* and *honestly*, and *fork* unless it means a GitHub fork. Plain words
  instead: option, important, accurate, explicit.

## Workflow

- All work happens on a branch or worktree, never directly on `main` — a ruleset forbids pushing
  it anyway. Pull `main` before branching. Every change lands through a PR.
- Stacked branches are fine, and the right way to break large work into smaller reviewable
  pieces. Check `gh pr list` before basing new work.
- Documentation is part of the work, not a follow-up. Before calling a change done, check
  `README.md`, `AGENTS.md`, `docs/`, and every skill the change touches, and land the updates as
  their own commit in the same PR.
- Conventional commits, with a body that says *why*. release-please reads the types
  (`release-please-config.json` maps them to changelog sections).
- release-please owns `CHANGELOG.md` → the repo-root CHANGELOG it releases alongside — and the
  extension's version markers come from the release-please PR alone.

## Testing a skill locally

- Codefall copies the embedded tree from the `extensions/` directory — point `init` at a sandbox
  checkout with the CLI binary built from the same working tree.
