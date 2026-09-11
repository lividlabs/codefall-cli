codefall
--------

**Opinionated skills for the software development lifecycle.**

This directory is the installable plugin. Everything above it — the marketplace catalog, the release
tooling, the repo README — is packaging and does not ship to users.

| Skill | Does |
| --- | --- |
| [`conceptualize`](skills/conceptualize/SKILL.md) | Get an idea onto paper before anyone specifies or scaffolds it: a numbered concept document under `docs/concepts/` carrying the problem, the rough shape of an answer, and what nobody has decided yet. |
| [`scaffold`](skills/scaffold/SKILL.md) | Start a new project on the Clean + package-by-component stance: ratified ADRs, scoped `AGENTS.md`, optionally project files and boundary lint. |
| [`graft`](skills/graft/SKILL.md) | Bring a scaffolded project's docs up to date with the current templates: report what changed since its version, with per-file provenance, and apply only what the user takes. Also handles first-time adoption of the stance. |
| [`specify`](skills/specify/SKILL.md) | Turn a feature idea into a specification another session can implement: a spec document under `docs/specs/` holding requirements with EARS acceptance criteria, mirrored to the issue tracker. |
| [`mock-up`](skills/mock-up/SKILL.md) | Get the visual surface of a feature into the repository under `docs/mockups/`: import what a design tool exported, or make the mockup here, matching the app's own design system. |
| [`design`](skills/design/SKILL.md) | Decide how a feature gets built and put the work into the graph: a design document under `docs/designs/` scaled to the change, ADRs for hard-to-reverse choices, and the tasks in Beads with their dependency edges. |
| [`implement`](skills/implement/SKILL.md) | Execute the graph: claim ready beads, build each in an isolated worker worktree with tests as part of done, verify against acceptance criteria, open PRs, and walk the waves until the frontier is empty. Never merges to `main`. |

Skills are explicitly invoked and carry `disable-model-invocation: true`, so none fire on their own.
The plugin also ships one hook: a `PreToolUse` guard that denies merges and pushes to the default
branch, because a human performs every merge to `main`.

See the [repository README](../../README.md) for the architectural stance, the surface catalog, and
installation instructions.

## License

[MIT](LICENSE) — Copyright (c) 2026 Livid Labs, LLC, authored by Dave Jensen.

The templates under `skills/scaffold/templates/`, and everything `scaffold` copies from them into
your project, are additionally available under [0BSD](LICENSE): no attribution, no notice, no
obligation. Your architecture documents are yours.
