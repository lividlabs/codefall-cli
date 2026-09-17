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
  consequence rather than a new decision, and not the TUI question a later ADR would settle.
  `schemas/settings.schema.json` is the published definition of `.codefall/settings.json`; the Go
  validator is hand-written in `domain/` so no schema library ships in the binary, and a test holds
  the two equal. One correction to ADR-002: Fang v1.0.0 does import `charm.land/lipgloss/v2`, so its
  note that Fang imports neither Lip Gloss major is out of date. Harmless — it is the v2 major, and
  the v1 generation stays out of the graph.
- **Second component, 2026-08-27.** `internal/initcmd/` is `codefall init`, the thing that creates
  what doctor checks. The package is named `initcmd` rather than `init` because `init` cannot be
  imported unaliased — Go rejects `import ".../internal/init"` because `init` must be a func — and
  `initcmd` keeps the directory named after the command it holds, the maintainer's choice on review
  over the earlier `setup`. The command is `init`, unaffected.
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
  `internal/initcmd/internal/application/` compiled and failed `golangci-lint run` on the
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
- **Init's Beads section, 2026-08-27.** codefall ships its own `AGENTS.md` section about Beads and
  `codefall init` writes it, as the fifth step and the last. `--skip-agents` is what makes this
  necessary: bd would append
  [`beads-section-minimal.md`](https://github.com/gastownhall/beads/blob/6c124203e771433a3550c348771a5b5e27fd3c21/internal/templates/agents/defaults/beads-section-minimal.md)
  to `AGENTS.md` and `CLAUDE.md`, and that text describes a setup codefall does not create. There is
  no invocation that keeps bd's context and codefall's own words in the same file — `--agents-template`
  applies only to an `AGENTS.md` that does not exist yet, and bd has no template hook for `CLAUDE.md`
  at all — so codefall writes the section itself. It lives at
  `internal/initcmd/internal/domain/beads_section.md`, embedded with `//go:embed` and exposed as
  `domain.BeadsSection`: the words codefall says are a fact about codefall, and `embed` is the
  standard library, which the domain layer's allow-list already permits. How the section is spliced
  into somebody's file is the application layer's, along with every other encoding.
  The section opens and closes with HTML comment markers, `<!-- BEGIN CODEFALL BEADS -->` and
  `<!-- END CODEFALL BEADS -->`, which is what makes a second run replace it in place rather than
  append a second copy — everything on either side of the pair survives byte for byte, so the
  project's own words are never this step's to rewrite. A file with an opening marker and no closing
  one stops the run: where codefall's words end and the project's resume is not something to guess
  at. A file with no markers keeps what it says and gains the section at the end.
  The step runs after `bd init` rather than before it, because bd stages `AGENTS.md` and `CLAUDE.md`
  when they exist and commits what it staged under its own message. Editing them afterwards leaves
  the change uncommitted, which is where it belongs: the section is the author's to commit.
  `CLAUDE.md` is created only for the `claude-code` harness and only when it is missing, holding one
  line that points at `AGENTS.md`. That is the file convention the codefall plugin's `scaffold` skill
  states — `AGENTS.md` holds the rules, and a harness that reads `CLAUDE.md` is pointed at them —
  and it is what this repository's own `CLAUDE.md` says. A `CLAUDE.md` that already exists is left
  alone whatever it holds, because writing a pointer over somebody's rules throws them away rather
  than pointing at them.

- **Shared modules, 2026-08-27.** `internal/shared/` exists, with the two modules the second
  component turned out to have copied from the first. `internal/shared/ui/` owns the palette, the
  styles, the marks, the colour-profile writer, and the spinner runner; a component maps its own
  vocabulary onto a `ui.Tone` and owns nothing else about how a line looks — doctor maps a check's
  status, initcmd maps a step's outcome and keeps its own dash for a skipped step, because nothing
  is wrong there. `ui.RunWithSpinner` is one entry point covering both paths, so neither command
  re-implements the branch between a terminal and a pipe: it takes the writer, draws finished lines
  under the spinner and writes them again once Bubble Tea has cleared its frames, and writes them as
  they arrive when there is nothing to spin on. `internal/shared/process/` owns running an external
  tool and reaching the file system, including the `0o755`/`0o644` modes. Each component's
  `infrastructure/` is now an adapter over it: the gateway interface and the result type stay the
  consumer's to declare (ADR-GO-01), so the two `CommandRunner` interfaces and the two
  `CommandResult` structs remain, and what the adapter adds is the translation.
  What deliberately stays duplicated is `firstLine` and the settings-format constants in both
  `domain/` packages: the layer rules keep `domain/` and `application/` off shared modules
  altogether, and moving those would mean either widening those allow-lists or moving the constants
  out of the layer that owns them. Both copies are pinned by each component's own schema test
  against `schemas/settings.schema.json`, which is what makes the duplication safe rather than
  merely tolerated. (Reversed the same day by **Pure shared modules** below, which widened the
  allow-lists for modules that do not need widening — the option not seen here.)
  The `shared-modules` rule now matches files and was proven, once per deny entry;
  proving it needs a throwaway `internal/shared/proof/` package, because a real shared module
  importing a component is an import cycle and so a compile error rather than the lint failure
  ADR-GO-02 asks for. `domain/` and `application/` were re-proven against an `internal/shared/ui`
  import the same way.

- **Pure shared modules, 2026-08-27.** `domain/` and `application/` may now import a shared module
  that is itself pure — one under `internal/shared/` importing only the standard library and
  `samber/mo`, the same imports `domain/` already has — which is what lets the settings format and
  the small helpers the two components had copied become one definition. The blanket denial that
  produced those copies was aimed at `ui` and `process`, which import Charm and `os/exec`, and it
  could not tell them apart from a module that adds nothing to the inner rings' dependency surface;
  each pure module is now named four times in `.golangci.yml` — the two allow-lists plus the
  `pure-shared-modules` rule's `files` and `allow` entries, the second of which is what also lets it
  import another pure module — so the property the permission rests on is checked rather than claimed. Recorded as [`ADR-003`](adrs/ADR-003-pure-shared-modules.md), a new ADR rather
  than an amendment to ADR-GO-02, which it refines and does not supersede.
  Two modules moved under it. `internal/shared/settings/` is now the one definition of the
  `.codefall/settings.json` format — the version, the schema id, the tracker names, the repository
  pattern, the field tables, `Validate`, and the field-level `ParseTracker`, `ValidateRepo`, and
  `ValidateProject` that a form calls while somebody is still typing. Doctor's `domain/settings.go`
  is gone; initcmd's keeps the `Settings` value object and `NewSettings`, which is where the
  combinations that make sense for a run are decided, and takes every constant and rule from the
  shared module. `Harness` stayed initcmd's, in its own file: nothing else in the project has an
  opinion about which harnesses can be set up. `internal/shared/text/` holds `FirstLine`, which both
  application layers had written identically. The two facade-level schema tests became one
  `schema_test.go` beside the format it pins, carrying every assertion either of them made — the
  reason it could not live there before was that a `domain` test cannot import `os`, and a pure
  shared module is not an inner layer. Names changed with the move, because the package name now
  carries what the prefixes used to say: `SettingsVersion` and `SettingsSchemaID` are
  `settings.Version` and `settings.SchemaID`, `ValidateSettings` is `settings.Validate`, and
  `RequiredSettingsFields` is `settings.RequiredFields`.
  Both halves of the new configuration were proven the way ADR-GO-02 requires, each compiling and
  failing `golangci-lint run`: `charm.land/lipgloss/v2` imported inside `internal/shared/settings/`
  failed the `pure-shared-modules` rule, and `internal/shared/ui` imported from
  `internal/initcmd/internal/domain/` failed the `domain-layer` rule — while that same package's
  import of `internal/shared/settings`, which is what the whole change rests on, passes.
- **Where init gets the repository, 2026-08-28.** `--issues-repo` is a flag nobody should have to
  type in the repository it names. Init asks two sources for it, in order: `gh repo view --json
  nameWithOwner`, and then `git remote get-url origin`. gh goes first because it answers with the
  repository as GitHub knows it today, which is right after a rename and for a fork whose remote
  still names the upstream; the remote goes second because it needs neither authentication nor
  network, and git is already required on every run while gh is required on none. Before the
  fallback, a machine without gh — or with gh installed and signed out — had to be told what the
  `.git/config` in front of it already said. The answer is used the same way either source produced
  it: it pre-fills the survey field when somebody is answering, and stands in for the flag when
  nobody is. A remote is only an answer when its host is `github.com` and what it points at passes
  the same `ValidateRepo` the flag passes — anything else, including an SSH host alias like
  `github.com-work` that a multi-account setup uses, is None and the flag is how the person says so.
  Only `origin` is consulted: a project whose GitHub repository is under some other remote name is
  one its owner knows better than init does.

- **Unified hooks, 2026-09-11.** `extensions/hooks/` is now the one source of truth for what init
  registers with each harness: `hooks/shared/` holds the script every guard runs
  (`codefall-block-merge-to-main.sh`, `--antigravity` selects the stdout-JSON deny contract that
  harness wants), and `hooks/<harness>/` holds the definition — `claude` and `codex` in the
  shared event shape (`PreToolUse` guard plus `bd prime --hook-json` on `SessionStart`),
  `antigravity` keyed by hook name with the guard only, and `opencode` as the plugin file that
  delegates the guard to the shared script and primes `session.created` sessions through
  `session.prompt` with `noReply: true`. Script paths resolve at the repository root two ways:
  Claude and Codex quote `$(git rev-parse --show-toplevel)`, which both harnesses run through a
  shell, and Antigravity uses the workspace-relative `./.agents/…` its own examples use — its
  command execution is not documented to expand `$(…)`, so nothing there depends on it. Either way
  the definitions carry literal commands and nothing rewrites paths.
  The extension step copies everything except those four definition directories — it is handed the
  same list the hook step reads — so `.claude/` gets `hooks/shared/` and `.agents/` gets the same,
  and a definition file lands only where the harness actually reads it. A copied script is made
  executable where it lands: the embedded tree carries no modes to copy, and a hook command names
  the script by path. The step's table
  (`hookSpecs`) holds each harness's source path, destination, and one of two formats: merge or
  copy. The merge is one recursive fold — objects in, arrays appended, scalars the destination
  already says kept — so Claude's `settings.json` keeps its permissions and Codex's `hooks.json`
  keeps its `description`, an Antigravity hook the project disabled stays disabled, and a key
  holding null says nothing the same as an absent key. Arrays
  dedupe per entry on the matcher together with the command string, the command found recursively
  because the formats keep it at different depths: the same command under another matcher still
  leaves the event unguarded, so it joins rather than counting as present. The harness's entry
  joins the project's under the same event key. OpenCode
  copies its plugin instead and skips when the bytes match; a harness with no table entry is
  skipped rather than errored. `domain.BeadsHookEvent`/`BeadsHookCommand` went away: what init
  installs is a property of the definition files now, and the format each file wants lives nowhere
  but the embedded tree. The `ExtensionFetcher` interface is renamed `ExtensionSource`: one
  gateway role for the embedded payload, fetched whole for the extension step, read one file at a
  time for the hook step.
  What an upgrade deliberately does not do is remove files the tree no longer ships. A project
  installed by the previous version keeps `.claude/hooks/hooks.json` and the un-namespaced
  `.claude/hooks/block-merge-to-main.sh`; the flat `hooks.json` was never a location Claude Code
  reads, so the leftover is inert, and the new registration is the only one that runs. Pruning
  manifest-recorded files the tree has since dropped is a separate change with its own rules —
  a manifest that means "mine to delete" is a stronger claim than one that means "mine to have
  written" — and is not made here.

- **Installing below the repository root, 2026-09-11.** `codefall init` run below the root of a
  git work tree asks, as its first question, whether to install in the working directory or at the
  root; `--location here|root` answers without asking, and a run with neither and no terminal
  fails naming the flag. At the root, and outside a work tree, there is no question. The case this
  serves is a large monorepo where one team wants codefall in its own part of the tree without
  setting it up for everyone else. Before this, a subdirectory run passed preflight, which asks only
  whether the directory is inside a work tree, installed everything under the subdirectory, and
  registered Claude and Codex guards naming the root's `.claude/` and `.agents/`, where nothing had
  landed — so every Bash call raised a hook error and the guard never ran.
  The **Unified hooks** entry's "nothing rewrites paths" no longer holds for those two: the hook
  step asks git for the install directory's path below the root (`rev-parse --show-prefix`) and
  writes it after `$(git rev-parse --show-toplevel)/` in every command that names the root. An
  install at the root has an empty prefix, so it registers exactly the command the definition ships
  and a rerun still recognises it. A prefix holding a character that keeps its meaning inside double
  quotes (`"`, `$`, backtick, backslash) is refused by preflight rather than escaped. Antigravity's
  `./.agents/…` is workspace-relative already and is unchanged; the OpenCode plugin now finds the
  guard from its own directory (`import.meta.dir`) rather than from the worktree root. Which
  directory each harness has to be started in to read a subdirectory install is that harness's own
  configuration discovery, and is not verified here for any of them.
  Beads needs no second database: `bd info` searches upward, so a subdirectory of a repository
  whose root has one reports that root database and the beads step skips, as it does anywhere else
  Beads is already initialised. A subdirectory that is the first install in its repository does get
  a database of its own, and there the run adds `--skip-hooks`. `bd init` otherwise points the
  clone's `core.hooksPath` at its own `.beads/hooks` — one setting for the whole repository, which a
  subdirectory install has no business claiming from everyone working in that clone, and which bd
  1.2.2 fills in with a path under the root even when it wrote the database below it, so the
  directory it names need not exist. Verified both ways against bd 1.2.2: without the flag a
  repository's own `core.hooksPath` is replaced by a path that does not exist, and with it the
  setting is left as it was. At the root the setting is bd's to make and the flag is not passed.

- **What the manifest asserts, 2026-09-11.** `.codefall/manifest.json` now means one thing: a run
  that finished, for the harness it names, at the version it names. It is written after the last
  step rather than in the extension step, and the gate that reports "already up to date" compares
  the harness as well as the version.
  Both halves were wrong in the same way — the gate answered from a record that claimed more than
  it knew. Written in step two, the manifest stamped the current version with the hook and
  AGENTS.md steps still to come: a run that failed at either left a project half set up and a
  record saying it was current, and the next run reported there was nothing to do. Reproduced
  against a project whose `.claude/settings.json` held `"hooks"` as a string: the run failed at the
  hook step, the rerun printed "already up to date", and the guard was never registered. Comparing
  the version alone did the same to a second harness: a project installed for `claude-code` and
  then run with `--harness antigravity` was skipped with no `.agents/` ever written, though the
  manifest had recorded the harness all along.
  The file list the manifest records is now carried out of the extension step as a local of the
  run, not a field of the use case, which every run of the process shares. A failed run leaves no
  manifest at all, so the next run repeats every step, which they are all built to tolerate — and
  the upgrade question is asked only when the version moves, not when the harness differs.
  What this does not do is record more than one harness. A project set up for two keeps the harness
  of the most recent finished run, so installing for the other one always does its work again.
  Recording a harness per install is a larger change to what the file is for, and the from/to
  comparison an upgrade reports does not need it.

- **What the hook merge calls the same entry, 2026-09-11.** The merge knew an entry by its matcher
  together with the literal command strings it ran, and appended anything else. Four things follow
  from that, and all four are fixed together because they are one question: what makes a
  registration codefall's own.
  A shipped command that changes read as an unrelated registration, so an upgrade appended the new
  one and left the old one running beside it — and no code path could ever remove it, since nothing
  recorded which entries were codefall's. The subdirectory install above makes this concrete: a
  project registered before it carries a guard naming the repository root, and the upgrade would add
  the corrected command without taking the broken one away. **An entry naming a codefall script
  now replaces the destination's entry naming that same script under the same matcher.** The
  script's file name is the identity, which is what the `codefall-` prefix is for; the rule is
  stated in `extensions/AGENTS.md`, beside the tree it governs. The alternative was recording every
  registered entry in `.codefall/manifest.json`, which would have made that file mean "mine to
  delete" — a stronger claim than "mine to have written", and one the **Unified hooks** entry had
  already set aside as its own change. It stays set aside.
  A project that primes Beads under a narrower `SessionStart` matcher had codefall's match-all entry
  appended beside it, which primed on every other session source and twice on the one the project
  chose. **A match-all entry is now covered by any entry running its commands under any matcher.**
  This restores what the pre-`623b53d` `hasBeadsHook` did, in terms that hold for any harness rather
  than for one command. A narrower incoming matcher still joins: the guard registered for `Edit`
  leaves `Bash` unguarded, and that is what the test with both matchers pins.
  An entry whose commands were only partly registered was appended whole, running the command the
  destination already had twice. **It now joins with only its unregistered handlers**, and an entry
  that keeps its command on itself rather than in a handler list joins whole or not at all, because
  there is nothing to split.
  **A destination holding an object or an array where the definition has a scalar is now reported**
  rather than quietly kept, which is what the same disagreement in the other direction already did.

- **Adding the execute bit rather than setting the mode, 2026-09-11.** A copied script was chmodded
  to `0755` on every run, which took a decision away from the project: a script tightened to its own
  user came back world-readable and world-executable at the next upgrade, with no line in the report
  to say so. `MakeExecutable` now adds execute wherever the file is already readable — `0600`
  becomes `0700`, `0644` becomes `0755` — and does nothing at all when the mode already allows it,
  so a rerun with nothing to change cannot fail on a directory that refuses `chmod`. The helper also
  stops passing `dirMode` for a file, which was the right bits under the wrong name.
  Which files get the bit does not change: every `.sh` in the tree, not only `hooks/shared/`. The
  hook commands name the guard by path, and `codefall-implement`, `codefall-specify` and
  `codefall-design` each invoke `shared/preflight.sh` by path — a script nobody can run is a skill
  that fails halfway through.
  Two tests were restored to pinning what they claim. The extension step's test now asserts the
  exclude list it hands `Fetch`, so a fetch that stopped excluding the per-harness definitions
  fails rather than silently shipping every harness's definition into every project. The refused
  merge's error table now also asserts the destination is byte-for-byte untouched, because the merge
  builds its document in place before a later key can fail it, and the error alone never said the
  file survived. Both were proven against the mutation each describes.

- **Harnesses are shared vocabulary, 2026-09-16.** `internal/shared/harness/` is now the one
  definition of which coding harnesses codefall can set up and where each one reads skills, as a
  third pure shared module. This reverses the note in **Shared modules, 2026-08-27** that `Harness`
  stayed initcmd's because nothing else in the project had an opinion about which harnesses can be
  set up: doctor is about to have one. Once a project records the harnesses it chose in
  `.codefall/settings.json`, doctor reads that field and reports on the directories it names, and
  then neither component can be the one that decides where those directories are.
  What moved is the five names, the sorted roster, and the destination directory per harness:
  `Harnesses()` and `ParseHarness` are `harness.All` and `harness.Parse`, and `extensionDestDirs` is
  `harness.SkillsDir`, an Option because a name `Parse` has already refused has no directory to
  report (ADR-GO-03). That one map is also the roster, so a harness codefall can set up is a harness
  with somewhere to install and the two cannot disagree. What stayed in initcmd is `hookSpecs` —
  which file in the embedded tree a definition comes from, and whether it is merged or copied, is how
  init registers a hook rather than what a harness is. Claude Code's hook destination is now the
  literal `.claude/settings.json` its three siblings already used, rather than being composed from
  the skills directory that moved.
  Nothing a user can see changes here: `--harness` still defaults to `claude-code`, and a run still
  installs for exactly one harness. This is the first of four changes that together let a project
  choose several, and it lands on its own so the move is reviewable without the behaviour change on
  top of it.
  Both halves of the new module's configuration were proven the way ADR-GO-02 requires. A
  `charm.land/lipgloss/v2` import in `internal/shared/harness/` compiled and failed
  `golangci-lint run` on the `pure-shared-modules` rule, and the same run left a deliberate
  `harness` import from `internal/initcmd/internal/domain/` unreported, which is what proves the
  `domain-layer` allow entry names the module. The `application-layer` entry needs no throwaway
  proof: the extension and hook steps import the module, and lint passes.

- **A run installs for a set of harnesses, 2026-09-16.** `Request.Harness` is `Request.Harnesses`,
  and the extension and hook steps each do their work once per harness the run is for. This
  supersedes the closing note of **What the manifest asserts, 2026-09-11**, which set recording a
  harness per install aside as a larger change to what the file is for. This is that change.
  `.codefall/manifest.json` now holds an entry per harness — the version that installed it and the
  files it wrote — and a run updates the entries for the harnesses it installed for while leaving
  every other entry as it was. A project set up for two keeps both records, where the file naming one
  harness lost the first record as soon as the second install finished, and the next run for that
  first harness repeated work it had already done. The recorded paths are relative to the directory
  init installed in rather than to the harness's own skills directory: four of the five harnesses
  read `.agents/`, so a bare `skills/…` would not say which directory it landed in.
  Every step reads the set through one function that sorts it and drops repeats, so the order a flag
  or a survey happened to collect them in cannot change what a step does or what it reports. The
  extension step copies once per directory rather than once per harness, because four of them share
  one, and every harness that reads a directory records the files in it. The hook step registers with
  each harness in turn and folds the outcomes: Done when any registration changed something, with
  every harness's own sentence in the detail. A run for one harness reports exactly what it reported
  before, which is what keeps this change invisible. The `CLAUDE.md` pointer is written whenever
  Claude Code is among the harnesses rather than when it is the harness. A run for no harness is
  refused by preflight, where nothing has been touched yet, rather than by six steps in turn
  reporting they had nothing to do.
  The no-op gate compares per harness: a rerun is a no-op when every harness it is for is recorded at
  this binary's version, and the upgrade question is asked only when one of them is recorded at a
  different version — a harness with no record at all is work to do rather than an upgrade.
  Nothing a user can see changes yet, because presentation still passes the one harness `--harness`
  names and that flag still defaults to `claude-code`. Choosing several is the next change in the
  stack. A manifest written before this one decodes to an empty record, so the next run repeats every
  step, which they are all built to tolerate; no project has a manifest yet, so nothing is migrated.

- **The project says which harnesses it uses, 2026-09-16.** `.codefall/settings.json` records a
  required `harnesses` array, `init` asks which harnesses to set up, and `--harness` no longer
  defaults to `claude-code`. The default was the whole of the old decision: a pflag carrying one
  cannot be told apart from one the user typed, so every run that did not pass the flag installed for
  Claude Code silently, and the survey had no harness question at all. Which harnesses a project uses
  is not something codefall can work out — four of the five share one directory, and codefall writes
  those directories itself — so it is asked, and the answer is recorded.
  The field lives in the shared settings module beside the tracker, because both components need it:
  init writes it and doctor is about to report on it (ADR-003). The validator refuses an empty array
  rather than reading it as "none", and refuses a name codefall cannot set up; the schema says the
  same two things with `minItems` and an enum, and the schema test holds the two definitions equal.
  This is also `internal/shared/harness/`'s first consumer outside initcmd: `settings` imports it to
  validate the names, which is the pure-module-importing-a-pure-module case ADR-003 allows and the
  fourth `.golangci.yml` entry permits.
  The survey asks for the harnesses first, as a multi-select with nothing selected and at least one
  answer required. Nothing is preselected on purpose: a preselected option is the same decision made
  on someone's behalf that the flag's default was. A scripted run passes `--harness`, repeated or
  comma-separated, and a run with no terminal and no flag fails naming it, which is the rule every
  other prompted value already follows (ADR-002).
  A rerun does not ask again, because the harnesses are in the settings and init reads them back — so
  the flag is needed the first time, or to add one. Settings that record none, which is what a file
  written before the field existed looks like, are refused with the flag named rather than filled in
  with a guess.
  This breaks twice over: a settings file without `harnesses` fails doctor's completeness check until
  init is run again, and a scripted run that relied on the flag's default now fails naming the flag.
  `codefall create` inherits the second, because it adopts init's flags. No project has a settings
  file yet, so nothing is migrated.

- **Doctor reports on the harnesses the settings name, 2026-09-16.** A fourth category, one check:
  codefall's extension is installed for every harness `.codefall/settings.json` records. It fails
  rather than warns, because a harness the project chose and codefall was never run for has no skills
  and no hooks there — every codefall verb is missing in a harness somebody is using — and the remedy
  is `codefall init`.
  What it stats is `hooks/shared/` under each harness's own skills directory, a path the shared
  harness module names and the initcmd facade test pins against the real embedded tree. Nothing but
  codefall writes that directory, where `.claude/` and `.agents/` exist in plenty of projects that
  have never run codefall — and `.agents/` is read by four of the five harnesses, so its presence
  identifies none of them. The check is deliberately not keyed to a skill's name: the skills are
  still being renamed, and a check that broke on a rename would fail on projects that are set up
  correctly.
  It reads the settings itself rather than taking them from the settings group, and runs only when
  that group found the file complete: settings doctor has already complained about say nothing about
  which harnesses were chosen, and a second complaint about the same file would be noise. A skipped
  check is absent from the report, which is what every other prerequisite failure already does. No
  presentation change was needed — `Sections()` groups by first appearance, so a new category reports
  itself.
  What this entry did not do is warn about a harness codefall finds installed that the settings do
  not name. That needs doctor to read `.codefall/manifest.json`, which means promoting the manifest
  format to a shared module, and codefall's stance is that it never removes what it wrote, so the
  remedy is somebody's to carry out by hand. **Doctor reports an install the settings dropped,
  2026-09-16** below is that change.

- **Doctor reports an install the settings dropped, 2026-09-16.** A second check in the Harnesses
  category: the manifest records an install for a harness `.codefall/settings.json` no longer names.
  A project set up for two harnesses that later drops one keeps every file codefall wrote for it,
  because codefall only ever writes what it owns and never deletes — so the skills for a dropped
  harness stay where an agent reading that directory will find them, and the hook codefall registered
  for it still runs.
  `internal/shared/manifest/` is now the one definition of the `.codefall/manifest.json` format, a
  fourth pure shared module, for the reason the settings format is one: initcmd writes that file and
  doctor reads it, and a file initcmd wrote that doctor could not read would be the worst bug in
  either (ADR-003). initcmd's private `manifest` and `harnessInstall` structs and its
  `domain.ManifestName` are gone; `Decode`, `Encode`, `Recorded` and `Versions` belong to the format,
  and `Installation` stays initcmd's because what the upgrade gate compares is initcmd's business. The
  format itself is unchanged, so no project's file needs anything done to it.
  The check warns rather than fails, because nothing stops working. Its remedy points at the manifest
  rather than naming a directory to delete: the manifest is what records which files codefall wrote
  for that harness, and harnesses share directories — four of the five read `.agents/` — so the
  directory a dropped harness read may still be another's. Which of those files are safe to remove is
  the reader's judgement, not doctor's.
  A manifest that is missing, unreadable, or no longer a manifest records no install that could be
  left over, so the check is absent from the report rather than complaining a second time: doctor has
  no check for the manifest's own shape, the install check has already reported whatever is missing on
  disk, and init fails on a file it cannot read the next time it writes one.
  The module's four `.golangci.yml` entries were proven the way ADR-GO-02 requires. A
  `charm.land/lipgloss/v2` import in `internal/shared/manifest/` compiled and failed
  `golangci-lint run` on the `pure-shared-modules` rule, and the same run left a deliberate `manifest`
  import from `internal/initcmd/internal/domain/` unreported. That second proof matters here in a way
  it did not for `harness`: no `domain/` package imports the manifest format, so nothing else would
  have exercised the entry.

## Open

- **UI composition.** Half settled by **Shared modules, 2026-08-27** above: the theme, the styles,
  the colour-profile writer, and the spinner runner now live in `internal/shared/ui/`, and rule 4
  holds there — proven, not assumed. What is still open is the other half: reusable Bubble Tea
  models (list, table, status bar) generic over the data they show, with each component's
  `presentation/` binding its own data, and a shell owning arrangement and navigation (`main` for
  the command tree; `main` or an `internal/tui/` component for a TUI — unsettled). A facade
  re-exports a view type by alias when the shell must name it. Alternatives seen and not taken: one
  shared module holding all presentation (breaks rule 4); a shell rendering generic widgets from
  contracts alone (widens every facade). Graduates to a later ADR with the first component that has a
  view, or a committed TUI, whichever comes first.

## Parking lot

Product detail heard during scaffolding but deliberately not acted on — it belongs to `specify` and
`architect`, not here.

- (empty — the project was scaffolded before its capabilities were described)
