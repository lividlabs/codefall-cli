# codefall

Opinionated skills for the software development lifecycle, plus the CLI that installs them. The repo
holds two components; each scopes its own operative rules:

- `cli/` — the Go installer. Rules: [cli/AGENTS.md](cli/AGENTS.md)
- `extensions/` — the skills, hooks, and shared scripts that `codefall init` installs. Rules:
  [extensions/AGENTS.md](extensions/AGENTS.md)

Decisions for the whole repository live in [`docs/adrs/`](docs/adrs/); the in-flight scratchpad is
[`docs/decision-log.md`](docs/decision-log.md). Provenance for the scaffolded documents is
`.codefall/scaffold.json`.

## Workflow

These hold for both components; each component's own file adds to them.

- Every change lands through a PR, and a PR is **squash-merged**. Merge commits and rebase merges
  are turned off on GitHub, and the branch is deleted on merge.
- **The PR title becomes the commit message** on a branch of more than one commit; a single-commit
  branch keeps that commit's own title. So the title is a Conventional Commit line naming the
  highest-impact change on the branch, and it carries `!` when any commit on the branch is
  breaking. Adding a commit means reading the title again — a branch that starts as `feat:` and
  gains a breaking rename needs a new title before it merges.
- release-please reads that one header, not the commits underneath it: the squashed body holds
  every commit message, and none of them is parsed. A commit that renames a flag, a settings
  field, or anything else a user types also carries a `BREAKING CHANGE:` footer, so the rename
  reaches the changelog by a second path if the title ever fails to.
