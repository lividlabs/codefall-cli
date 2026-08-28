# Decision log

The scratchpad for decisions in flight. Settled decisions graduate to an ADR in `docs/adrs/` and are
struck from `Open`; this file never holds the reasoning that belongs in an ADR.

## Locked

Decided at scaffold, 2026-08-16.

- **One surface.** `codefall-cli` is a single Go CLI — no backing service, no UI, no seam. Profile:
  `go`. Inherited: ADR-BASE-01, ADR-BASE-02, ADR-GO-01, ADR-GO-02, ADR-GO-03.
- **Module path** — `github.com/lividlabs/codefall-cli`. No git remote set yet.
- **One app, one package.** Single Go module at the repo root; `cmd/codefall/` is the only binary.
  Codefall has no house opinion on repo layout, and one app doesn't need one.
- **Package-by-component**, chosen over ports-and-adapters even though the capabilities are unnamed —
  see the "For this project" section added to ADR-BASE-02.
- **No components yet, and no example one.** `internal/` is empty by choice: a placeholder component
  was scaffolded, reviewed, and deleted as ceremony that would be replaced wholesale in the next
  release. `main.go` prints the banner directly and wires nothing.
- **No dependencies.** `go.mod` requires nothing. `samber/do` (ADR-GO-01) and `samber/mo`
  (ADR-GO-03) are governing decisions, not yet imports — they arrive with the code that needs them.
- **Boundary enforcement is wired**, ahead of the code it will govern: `internal/` facades
  (compiler) plus `depguard` strict allow-lists in `.golangci.yml` (ADR-GO-02), both proven against
  deliberate violations at scaffold time.
- **Contracts cross facades, not entities.** A facade returns a data shape declared in the
  component's root package, never a `domain` type — recorded as `ADR-001`. Go's `internal/` rule
  does not cover this: it restricts naming a package, not holding a value, so a leaked entity
  compiles for a caller that cannot import it. Neither the compiler nor `depguard` can catch it.
- **`ADR-BASE-03` (extraction readiness) is deliberately not part of this project's set.** Its gate
  is whether a surface could ever be split into a service, and a CLI cannot. Four of its five rules
  have no referent here; the fifth is `ADR-001`. If a backing service ever appears, revisit.
- follows codefall `0.4.0` as of `2026-08-19`
- grafted codefall `0.4.0` on `2026-08-19` — took the `ADR-BASE-02` revision (re-applied in place,
  carrying the `For this project` amendment forward; no supersession, as no code had been written
  against it) and the `AGENTS.md` workflow defaults. Skipped `ADR-BASE-03` as inapplicable.
  `ADR-BASE-01` and `ADR-GO-01/02/03` were already current.
- **ADR-GO-01 amended, 2026-08-19.** Its claim that the composition root "is the only place that
  names concrete implementations" contradicted the layout: `cmd/` cannot import
  `internal/<component>/internal/*`. Concretes are named in each component's registration function;
  `main` calls those functions and resolves facades only. The same fix is owed to the plugin
  templates (go and typescript-react profiles).
- **CLI libraries, 2026-08-27.** Cobra for the command tree, run through Fang; Lip Gloss v2 for
  styled output where it helps; Huh v2 for prompts. Charm's v2 generation only, on the
  `charm.land/…/v2` paths — the two majors do not mix. Recorded as `ADR-002`. Bubble Tea arrives
  transitively with Huh and is the answer if a TUI is ever needed; that is a consequence of ADR-002,
  not a new decision. None of it is imported yet — each library arrives with the code that first
  calls it.
- **First component, 2026-08-27.** `internal/doctor/` is the first capability under `internal/`:
  `codefall doctor` reports whether a directory has `.codefall/settings.json`, Beads, and an
  authenticated `gh` with the scopes codefall needs. It reports and never repairs — it prints the
  remedy for every unmet check and `codefall init` will be the thing that runs them. The composition
  root now exists: `cmd/codefall/main.go` builds the injector, calls `doctor.Register`, mounts
  `doctor.Command`, and hands the tree to `fang.Execute`. The layer rules were re-proven the way
  ADR-GO-02 requires — a facade reach-around from `main` failed `go build`, and a Cobra import from
  `internal/doctor/internal/application/` compiled and failed `golangci-lint run` on the
  `application-layer` rule. Facade shape, settling ADR-002's open question:
  `Register(do.Injector)` plus `Command(do.Injector) *cobra.Command`, with
  `Commands(do.Injector) []*cobra.Command` when a component has several top-level commands. No
  report contract crosses the facade yet, because nothing outside doctor consumes one; add
  `doctor.Report`/`doctor.CheckResult` with the first external consumer, likely `init` (ADR-001).
  Now imported: Cobra, Fang, `samber/do`, `samber/mo`, `charm.land/lipgloss/v2`, and
  `github.com/charmbracelet/colorprofile`. Huh still waits for the first prompt.
  Bubble Tea and Bubbles v2 joined with doctor's spinner, in `presentation/` only — an ADR-002
  consequence rather than a new decision, and not the TUI question ADR-003 would settle.
  `schemas/settings.schema.json` is the published definition of `.codefall/settings.json`; the Go
  validator is hand-written in `domain/` so no schema library ships in the binary, and a test holds
  the two equal. One correction to ADR-002: Fang v1.0.0 does import `charm.land/lipgloss/v2`, so its
  note that Fang imports neither Lip Gloss major is out of date. Harmless — it is the v2 major, and
  the v1 generation stays out of the graph.
- **Second component, 2026-08-27.** `internal/setup/` is `codefall init`, the thing that creates
  what doctor checks. The package is named `setup` rather than `init` because `package init` is
  legal but importing it is not: Go rejects `import ".../internal/init"` outright, since `init` must
  be a func, so the import would need an alias. One aliased import in the whole codebase is a worse
  trade than naming the directory after what it does. The command is `init`, unaffected.
  A run is an ordered list of steps, each reporting Done or Skipped through an `Observer` the
  presentation layer implements twice: a spinner with the running step's title when stdout is a
  terminal, plain lines everywhere else. This change carries one step, `settings`; installing the
  Claude Code plugin and running `bd init` are the next two, appended to the same list. Settings
  that already exist are left alone and the step reports Skipped — `--force` rewrites them — so init
  is safe to run again once it has more steps than this. `--harness` is accepted and validated now,
  with `claude-code` the only value, because the steps after this one are the ones that need it.
  Huh v2 is now imported, the first prompt in the project (ADR-002). The tracker survey lists Jira
  and Linear with "(not yet available)" in the label and refuses them in the field's own validation:
  Huh v2 has no disabled option, and a tracker that is not offered at all reads as a tracker nobody
  thought of. The layer rules were re-proven the way ADR-GO-02 requires — a Cobra import from
  `internal/setup/internal/application/` compiled and failed `golangci-lint run` on the
  `application-layer` rule. Still no report contract crosses a facade: init does not consume
  doctor's.
- **Init's plugin step, 2026-08-27.** `codefall init` installs the codefall plugin for Claude Code
  as its second step, by running the harness's own CLI — `claude plugin marketplace add
  lividlabs/codefall-plugin --scope project` and then `claude plugin install codefall@codefall
  --scope project -y`. Both are project scope, so `.claude/settings.json` carries the marketplace
  declaration and the enabled plugin and everyone who clones the repository gets them; a user-scope
  install would work for whoever ran init and for nobody else. What is already done is read out of
  that file rather than asked of `claude plugin list --json`, whose `enabled` field is computed from
  the caller's working directory and stamped onto every row, so it misreports. The CLI merges into
  the file, preserving keys it does not own, which is why init runs the CLI rather than writing the
  file itself. The plugin's identifiers are domain constants because they are facts about codefall;
  the CLI's name and argument shapes are application, because they are how a tool is asked. The step
  switches on the harness with `claude-code` as its only case — presentation has already refused
  every other value, so the default is unreachable and is where the next harness lands. Every run
  now starts with a preflight that checks the tools it will reach for are on PATH, before the first
  step: a tool that turns up missing halfway through leaves the project half set up, which is the
  one state init exists to avoid. It reports every missing tool at once rather than the first, and
  `bd` and git join `claude` there with the Beads step. Every tool init sets up is required on every
  run, including a run where every step would skip, because a run that cannot act is not a run that
  finished. The step's two halves are decided separately: the marketplace is declared whenever the
  project does not already declare it, and the plugin installed whenever it is not already enabled,
  because a plugin enabled with no project-scope marketplace beside it is exactly what a user-scope
  install leaves behind — it works for whoever ran it and for nobody who clones the repository.
- **Init's Beads steps, 2026-08-27.** `codefall init` finishes by initialising Beads and giving
  Claude Code the hook that primes a session with what Beads knows — the third and fourth steps, and
  the last of them. The invocation is `bd init --non-interactive --skip-agents`, which writes
  `.beads/` and appends a Dolt block to `.gitignore` and nothing else. `--skip-agents` is what keeps
  it to that: without it bd appends its own section to `AGENTS.md` and `CLAUDE.md` and writes
  `.claude/settings.json`, `.codex/`, and `.agents/`, and codefall owns the first two. bd has no
  template hook for `CLAUDE.md` at all — `--agents-template` applies only when `AGENTS.md` does not
  exist yet — so there is no invocation that keeps bd's context and codefall's own words in the same
  file. The text bd would have appended is
  [`beads-section-minimal.md`](https://github.com/gastownhall/beads/blob/6c124203e771433a3550c348771a5b5e27fd3c21/internal/templates/agents/defaults/beads-section-minimal.md);
  whoever wants codefall's own `AGENTS.md` to say the same things starts there.
  The session hook is how a Claude Code session gets that context instead: a `SessionStart` entry
  running `bd prime --hook-json`, the hook bd installs for the same purpose, written by init
  rather than by `bd setup claude` — that command cannot be told to leave `CLAUDE.md` alone. `bd
  setup claude --check` accepts what init writes (`✓ Project hooks installed`) and complains only
  about the missing `CLAUDE.md` section, which is the intended difference. The hook step decodes
  `.claude/settings.json` into a plain object so that every key the file has survives being written
  back; key order is the one thing that does not, because `encoding/json` sorts it. It is written
  back through a `json.Encoder` with `SetEscapeHTML(false)`, because `json.Marshal` escapes `<`, `>`
  and `&` and would silently rewrite a permission rule like `Bash(a && b)` in a file codefall does
  not own. It runs after the plugin step because both write that file, and the harness CLI's own
  merge goes first.
  Three guards run in preflight, because bd init decides things for itself that nobody asked.
  It commits what it wrote with `git commit --no-verify` and no pathspec, and before that it stages
  `.gitignore`, `AGENTS.md`, `CLAUDE.md`, `.claude/settings.json`, `.codex`, and `.agents` by name
  when they exist — so both anything already in the index and any uncommitted change to one of those
  paths lands in a commit that says it initialised Beads, under bd's message rather than its
  author's. `--stealth` is the only flag that stops the commit and it also stops tracking `.beads/`,
  which is not the trade. So a run refuses to start while `git diff --cached --quiet` reports a
  staged index, and refuses again when `git status --porcelain` over exactly those six paths reports
  anything — staged, unstaged, or untracked — naming the paths git reported. Both run only when bd
  init is going to run at all, because otherwise what the directory has waiting is nobody's business
  but whoever left it there. And bd init run outside a repository silently runs `git init` first, so
  a run also refuses a directory that `git rev-parse --is-inside-work-tree` does not answer `true`
  in — the word, not the exit code, because inside a bare repository and inside `.git` itself git
  prints `false` and exits 0. The path guard runs in preflight rather than later so that the plugin
  step's own write to `.claude/settings.json` is not what it catches: that write is codefall's to
  make and bd committing it is harmless, and the difference between it and somebody's uncommitted
  edit to the same file is exactly when the question is asked.
  `bd info` is the question that decides whether there is anything to do, the same question doctor
  asks and for the same reason — `BEADS_DIR` relocates `.beads/`, so looking for the directory asks
  something else. Re-running `bd init` where it has already run is an error rather than a no-op, so
  asking first is what keeps init safe to run again.

## Open

- **UI composition.** Shared UI widgets — theme, styles, the colour-profile writer, key maps,
  reusable Bubble Tea models (list, table, status bar) — under `internal/shared/ui/`, generic over
  the data they show; rule 4 means that module never imports a business component. Each component's
  `presentation/` binds its own data to a shared widget, and a shell owns arrangement and navigation
  (`main` for the command tree; `main` or an `internal/tui/` component for a TUI — unsettled). A
  facade re-exports a view type by alias when the shell must name it. Alternatives seen and not
  taken: one shared module holding all presentation (breaks rule 4); a shell rendering generic
  widgets from contracts alone (widens every facade). Graduates to `ADR-003` with the first
  component that has a view, or a committed TUI, whichever comes first.

## Parking lot

Product detail heard during scaffolding but deliberately not acted on — it belongs to `specify` and
`architect`, not here.

- (empty — the project was scaffolded before its capabilities were described)
