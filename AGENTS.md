# codefall

Opinionated skills for the software development lifecycle, plus the CLI that installs them. The repo
holds two components; each scopes its own operative rules:

- `cli/` — the Go installer. Rules: [cli/AGENTS.md](cli/AGENTS.md)
- `extensions/` — the skills, hooks, and shared scripts that `codefall init` installs. Rules:
  [extensions/AGENTS.md](extensions/AGENTS.md)

Decisions for the whole repository live in [`docs/adrs/`](docs/adrs/); the in-flight scratchpad is
[`docs/decision-log.md`](docs/decision-log.md). Provenance for the scaffolded documents is
`.codefall/scaffold.json`.
