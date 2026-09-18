codefall
--------

**Opinionated skills for the software development lifecycle.**

This directory is the extension. Everything above it is packaging.

| Skill | Does |
| --- | --- |
| [`codefall-conceptualize`](skills/codefall-conceptualize/SKILL.md) | Get an idea onto paper before anyone specifies or scaffolds it: a numbered concept document under `docs/concepts/` carrying the problem, the rough shape of an answer, and what nobody has decided yet. |
| [`codefall-scaffold`](skills/codefall-scaffold/SKILL.md) | Start a new project on the Clean + package-by-component stance: ratified ADRs, scoped `AGENTS.md`, optionally project files and boundary lint. |
| [`codefall-graft`](skills/codefall-graft/SKILL.md) | Bring a scaffolded project's docs up to date with the current templates: report what changed since its version, with per-file provenance, and apply only what the user takes. Also handles first-time adoption of the stance. |
| [`codefall-specify`](skills/codefall-specify/SKILL.md) | Turn a feature idea into a specification another session can implement: a spec document under `docs/specs/` holding requirements with EARS acceptance criteria, mirrored to the issue tracker. |
| [`codefall-mock-up`](skills/codefall-mock-up/SKILL.md) | Get the visual surface of a feature into the repository under `docs/mockups/`: import what a design tool exported, or make the mockup here, matching the app's own design system. |
| [`codefall-design`](skills/codefall-design/SKILL.md) | Decide how a feature gets built and put the work into the graph: a design document under `docs/designs/` scaled to the change, ADRs for hard-to-reverse choices, and the tasks in Beads with their dependency edges. |
| [`codefall-implement`](skills/codefall-implement/SKILL.md) | Execute the graph: claim ready beads, build each in an isolated worker worktree with tests as part of done, verify against acceptance criteria, open PRs, and walk the waves until the frontier is empty. Never merges to `main`. |
| [`codefall-review`](skills/codefall-review/SKILL.md) | Review something and fix what the user accepts: uncommitted work, a branch, an open PR, a path, a document, or a description of what to look at. A subagent or another harness reviews, the session triages with the user and applies what they take, and every finding is committed under `.codefall/reviews/`. |
| [`codefall-test`](skills/codefall-test/SKILL.md) | Run what the project declares: every suite, the subset the changed files reach, a named subset, or one test case in its `spec` or `agentic` modality. A spec case runs through the project's own runner; an agentic case is driven step by step and judged against the case's criteria. Every run is reported under `.codefall/tests/`, and findings are triaged rather than turned into edits that make a run pass. |
| [`codefall-equip`](skills/codefall-equip/SKILL.md) | Equip a project with the two local-environment scripts `codefall-refresh` runs — `start` and `update` — by finding what the project already has or drafting them from what the repository shows, then declaring them under `local` in `.codefall/settings.json`. Builds and rebuilds; scaffold and implement follow it as the procedure. |
| [`codefall-refresh`](skills/codefall-refresh/SKILL.md) | Bring the checkout and the local environment current: fetch, fast-forward the default branch when that is safe, run the declared `start` and `update`, record the commit the environment now matches, and turn a failure into a sentence that says what to do. Safe to run at any time. |

Skills are explicitly invoked and carry `disable-model-invocation: true`, so none fire on their own.
The extension also ships hooks per harness, defined under [`hooks/`](hooks/): a `PreToolUse` guard
that denies merges and pushes to the default branch everywhere, plus a `SessionStart` prime on what
Beads knows for the harnesses that have the event (Claude Code and Codex). The shared script the
guards run lives in `hooks/shared/`; `codefall init` merges the definitions or copies the plugin.

`codefall init` copies three of the directories here and leaves the rest. `skills/` goes into each
chosen harness's own skills directory, because that is the only part a harness finds by convention.
`hooks/shared/` and [`shared/`](shared/) go into the project's `.codefall/`, once per install
whatever harnesses were chosen, because every path that reaches them is codefall's at both ends — a
skill names a shared file `../../../.codefall/shared/<file>`, and every harness's hook definition
names `.codefall/hooks/shared/…`. The per-harness definitions under `hooks/<harness>/` are read
straight from the binary and never copied. This file, [`AGENTS.md`](AGENTS.md),
[`skills/AGENTS.md`](skills/AGENTS.md), [`docs/`](docs/), and each skill's `NOTES.md` are written for
someone working on codefall and are installed nowhere.
[ADR-006](../docs/adrs/ADR-006-install-layout.md) records the layout.

See the [repository README](../README.md) for the architectural stance, the surface catalog, and
installation instructions.

## License

[MIT](../LICENSE) — Copyright (c) 2026 Livid Labs, LLC, authored by Dave Jensen.

The templates under `skills/codefall-scaffold/templates/`, and everything `codefall-scaffold` copies
from them into your project, are additionally available under [0BSD](../LICENSE): no attribution, no
notice, no obligation. Your architecture documents are yours.
