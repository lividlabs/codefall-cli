# ADR-007: Test Cases

## Status

Accepted — 2026-09-17

## Context

Codefall already says where an expectation is written. `codefall-specify` produces a spec whose
criteria carry identifiers — `SPEC-003-REQ-01-AC-01` is a criterion inside a requirement inside a
spec — and `codefall-design` turns those criteria into beads whose acceptance criteria
`codefall-implement` verifies one by one. What codefall does not say is what verifies the wired
product: the path through the real interface, against the real services, as a user meets it.
Implement counts tests toward done and leaves what a test is to the project.

Unit tests can stay that way. End-to-end tests cannot, for one reason. A test written after the
code, from the code, certifies what the code does, bugs included — it passes on a defect as readily
as on correct behaviour, and it passes loudly, which is worse than not existing. Every decision
below exists to keep that from happening, and the point of recording them here rather than in a
skill is that a rule about where an expectation may come from is not a procedure; it outlives
whichever verb carries it.

The design this record adopts was built elsewhere. `demo-flights`, a flight-booking application,
covered its shipped flows in a backfill and wrote what it learned as ADR 0011, *E2E test cases and
the backfill discipline*, with the working conventions in its `testing/CLAUDE.md`. That record
settled: one markdown case file as the single home for a case's inputs, prompts, and expectations;
the file path as the case identifier with no registry beside it; two modalities, a generated spec
and an agent working the same case by hand, each with its own verdict vocabulary; no mocking at any
layer; a state that cannot be forced dropped as a criterion rather than dressed up as a manual test;
and a triage vocabulary that separates a product's own nondeterminism from a flaky test. All of that
is taken as it stands.

What does not transfer is everything that knows one project. This is the split ADR-005 already drew
for the local environment: codefall serves many project shapes, so it can own the contract and the
moments at which the contract is called, and not the commands. Left out for that reason: the Sabre
CERT notes about which routes return which inventory, the stranded-booking ledger, the npm aliases,
and the TypeScript loader the model's case files are read through. Each is a fact about one
supplier, one stack, or one runner. The shape they take here is a scoped `AGENTS.md` under the
testing root that the project fills in, and a runner the project declares.

Three things are changed rather than carried.

**The criteria come from the spec.** In the model, expectations came from GitHub issues, because an
issue was the only record that predated the implementation. Codefall produces specs, so the spec's
numbered criteria are the source, cited in full, and an issue tracker is not required for a project
to write a case.

**A case may go past the spec, and says so.** Pinning a case to the spec's criteria alone makes the
spec the ceiling as well as the floor, which no tester would accept: boundaries, negative paths, and
error handling are the work. A criterion beyond the spec is marked `derived` and cites the
requirement it elaborates, so a reader can tell in one line what the spec asked for and what testing
added.

**Agentic run reports are committed.** The model git-ignored everything a run produced. A report
that records which criteria held, how many attempts each step took, and which driver ran is the only
way variance becomes countable across runs rather than remembered, and it is small prose. The bulky
output stays ignored.

Once any project has case files, the format below and the `test` block are a breaking-change surface
by the definition in the repository `AGENTS.md`: a project cannot re-derive its cases, and a renamed
frontmatter key or a moved directory is a change every such project answers by hand.

## Decision

### What a case is, and where it lives

A case is a markdown file at `<root>/test-cases/<area>/<slug>.md`, where `<root>` is the testing
root the project declares and `<area>` is a product area. Frontmatter carries `id`, `modalities`,
`variants`, and optionally `messages`; the body carries Preconditions, Variants, Criteria, Steps,
and Reporting. `messages` holds the fixed strings a run sends and applies only to a conversational
surface; a surface with no conversation has no use for it.

Variants are named input sets, and one case generates one run per variant. A difference in inputs
with the same expected outcomes is a variant; a difference in expected outcomes is a separate case.
A criterion observable only under some variants names them, and the run skips it elsewhere.

The format is documented in `codefall-test/reference/case-file.md` and owned by `codefall-test`.

### The path is the identifier

A case's id is its path minus the extension, relative to `test-cases/` — `checkout/place-order`.
There are no `CASE-01`-style identifiers and no registry: the directory of case files is the list.
The frontmatter `id` duplicates the path deliberately, so that a file moved without its frontmatter
is caught mechanically rather than at review. Every test title begins with the case id.

A case in the `spec` modality is a sibling pair: `<slug>.md` beside a file whose name begins with
the same slug, in whatever suffix the project's runner collects. The pairing is checked both ways —
a case declaring `spec` has a sibling, and a sibling has a case.

### A case precedes its spec

A case file is written before the spec generated from it, never afterwards to describe what a spec
already does. Writing it afterwards is the same mistake as writing criteria from the code, made one
file at a time.

### Two modalities, and two verdict vocabularies

`modalities` accepts `spec` and `agentic`, and nothing else. There is no `manual` modality: a case
nobody runs is a coverage claim with nothing behind it, and the label was doing the work of hiding a
state that cannot be forced.

A `spec` run reports through its runner's own pass and fail. An `agentic` run — an agent working the
case's `## Steps` through a driver — uses its own two vocabularies, and they belong to that modality
alone. A criterion is `held`, `failed`, `skipped`, or `unreachable`; a case run as a whole is
`PASS`, `FAIL`, `ERROR`, or `PARTIAL`. Per-step attempt counts are recorded even when the run
passes, because a pass that took three attempts is data about the product. Every agentic run ends
with an anomaly sweep: user-visible wrongness the run noticed that no criterion asked about, listed
beside the verdict and never folded into it.

### Where a criterion comes from

**The spec's criteria are the floor.** A criterion that restates one cites it in full —
`SPEC-003-REQ-01-AC-01`, the form `codefall-specify` writes and the form `grep` finds. A case covers
every criterion of the requirement it tests, or records why one was dropped.

**A criterion may go beyond the spec, marked `derived`.** It cites the requirement it elaborates
(`SPEC-003-REQ-01`) and says in one line what it adds: a boundary, a negative path, error handling.

**The one banned source is the implementation.** Not the code, not clicking through the application
to see what it does, and not the implementing agent's own pull request text or notes — in a codebase
built by agents, those were written by the same author as the code and are implementation wearing
documentation's clothes. Reading code for mechanics is allowed and often necessary: a control's role
or its accessible name, so that a spec can drive it. Reading code to decide what should happen is
not.

**A gap the expansion exposes goes back to the spec.** Where the spec is silent or ambiguous the
case still states the expectation as `derived`, and the gap is reported for `codefall-specify` to
add a criterion. Spec numbering is append-only, so the addition takes the next number and nothing
already cited moves. Derived criteria are shown to the user before a case file is written, at the
point described below where the user already approves the plan.

### Nothing is mocked, and an unforceable criterion is dropped

No mocking, faking, or intercepting at any layer — not the project's own API, not an upstream
service. A state that cannot be reached through the product's own interfaces is unforceable: the
criterion is dropped and the drop is recorded in the case file, so the omission reads as a decision
rather than an oversight.

### The `test` block

`.codefall/settings.json` gains an optional `test` block:

```json
"test": {
  "dir": "testing",
  "runners": ["playwright"]
}
```

`dir` is required when the block is present, and is relative to the directory holding `.codefall/`.
It defaults to `testing` and `codefall init` asks for it. `runners` is an optional array of known
names, validated against an enum — `playwright`, `go-test` — exactly as `harnesses` is, and written
by `codefall-equip` rather than by hand.

The division is by reader. **What a program reads is in settings; what an agent reads is in
`<root>/AGENTS.md`.** The runner names are in settings because programs read them: the check scripts
decide which runner-specific checks apply, and doctor decides whether the project is equipped. The
commands that run the suite are prose in `<root>/AGENTS.md`, because only an agent or a person runs
them.

A runner is chosen per surface: Playwright for a browser front end, an Electron shell, or an HTTP
API; `go test` for a Go surface. React Native, Tauri, and Flutter take no `spec` runner in this
version; they may still carry `agentic` cases wherever the harness has a driver.

### The files

- `<root>/AGENTS.md` — the scoped rules, holding the facts only the project can know: each runner on
  one line with its run-all and run-one commands, the setup and state-forcing commands, the rule for
  recording and cleaning up real side effects, and environment notes.
- `<root>/CLAUDE.md` — one line pointing at it, for a harness that reads `CLAUDE.md`.
- `<root>/README.md` — the tree in a few lines, and the case index.
- `<root>/test-cases/` — the cases.

`codefall init` creates the tree and both skeleton documents, each only when it is missing. A
skeleton is written whole or not at all: it is never appended to and never rewritten, because from
the moment it exists it is the project's document.

The root `AGENTS.md` gains a third marked section, `CODEFALL TESTING`, written by `codefall init`
exactly as the Beads and Local environment sections are — its own marker pair, the text between the
markers replaced and nothing outside them touched, the section appended at the end when the file has
no markers, the file created with it when there is no file, and the run stopped before anything is
written when an opening marker has no closing marker. It names the testing root and says that cases
are written through `codefall-implement`, run through `codefall-test`, and the harness is set up
through `codefall-equip`. `codefall-equip` does not write it, and `<root>/AGENTS.md` holds project
facts only.

### What run output is committed

Output is split by kind. Runner output, logs, traces, HTML reports, and any run-scoped account or
session are git-ignored under `<root>/.artifacts/`. An agentic run's report is committed under
`.codefall/tests/`, beside `.codefall/reviews/` and with the same `.ignore` treatment, and it
records which driver ran. Triage working notes stay git-ignored; a finding becomes an issue only on
the user's word, and a criterion is never edited to make a run pass.

### Which verb sets up, decides, writes, and runs

**`codefall-equip` sets the harness up.** Its scope widens from the local scripts to the two things
a project needs equipped, and the widening is a change to its own `SKILL.md`: ADR-005 never
restricted what equip may own. For testing it reads the `test` block and stops when there is none,
naming `codefall init`; searches for a runner configuration and end-to-end directories the project
already has; asks one question with that evidence; installs or declares the runner; writes
`test.runners`; records the runner's line in `<root>/AGENTS.md`; revises the declared `update`
command for the runner's own install; and proves the result with the runner's list command and
`check-cases.sh`. Setting a harness up is its own pull request in the user's project. ADR-005's
point-of-introduction rule stays limited to `start` and `update`, and no task's pull request sets up
a test harness along the way.

**`codefall-design` decides which tasks need a case.** It names the modalities and drafts the case's
criteria — spec-cited or `derived` — into the bead's acceptance criteria, where the user already
approves the task plan. That approval is the sign-off the derived criteria need, and the bead's
criteria are the signal every later verb reads. `agentic` is chosen only where verifying an outcome
needs judgement; `spec` otherwise; neither where the project does not need a case at all. A driver
being available is not a reason to write an agentic case.

**`codefall-implement` writes the case file.** The worker writes it from the bead's approved
criteria before it writes the code, and then the spec where the modality calls for one. Authoring in
this version is limited to implement, working forward from the spec; a backfill procedure for flows
that shipped before their cases is a later decision. When a bead in scope calls for a case and the
harness is not equipped, implement says so, names `codefall-equip`, and does not start those beads;
beads needing only unit tests proceed.

**`codefall-test` runs.** It takes three kinds of target: the project's declared suites — all of
them, the subset the changed files reach, or a named subset — resolved from the same three sources
`codefall-implement` resolves its commands from; a case in `spec` modality; and a case in `agentic`
modality. A case target takes its run command from `<root>/AGENTS.md`. An agentic run brings the
application up through the declared `start` command and names `codefall-refresh` when it is down.

**The driver is decided by the case's surface.** A browser tool for a web or Electron surface, the
shell for a command-line or HTTP surface — and the shell is always there. The agent reads its own
tool list and names the driver it found in the confirmation, rather than consulting a registry that
would go stale. One live call through the driver proves it before the run depends on it, and a
failure is reported rather than worked around. With no working driver the agentic modality is
refused for that case, with a sentence saying how a checked-in MCP server entry supplies one.

### A verb names the remedy and never applies it

No verb runs `codefall-equip` or `codefall init` on the user's behalf. This is the rule for
everything equip sets up, and doctor follows it: `test-declared` warns when the block is absent,
with `codefall init` as the remedy; `test-equipped` warns when no runner is declared, with
`codefall-equip`; `test-dir-exists` fails when a declared directory is not there, because the
project declared it and a verb will look for it.

### The check scripts

`check-cases.sh` is installed once at `.codefall/shared/check-cases.sh`, from `extensions/shared/`,
and a project's CI calls that path. It checks the frontmatter `id` against the file's path, that
`modalities` holds only the two known values, that variant names are unique, that a case declaring
`agentic` carries a `## Steps` section, and the sibling pairing in both directions. It reads the
testing root from the `test` block with jq and falls back to `testing/`, and it treats a file beside
`<slug>.md` whose name begins with the slug as that case's spec, so it needs to know no runner's
suffix.

The Playwright-specific checks sit beside it in `.codefall/shared/` and run only when `test.runners`
names `playwright`: the rule that every test title begins with its case id, read through
`playwright test --list`, and the rule against `isVisible()` on an un-narrowed locator, which throws
on a multiple match and reports the throw as an absence.

### The session-start notice

`hooks/shared/codefall-session-notice.sh` calls the shared preflight, turns the lines that need
attention into one short notice, prints nothing when everything is current, and always exits `0`.
Its fetch is bounded by a short timeout or skipped. It reports only: a hook never pulls and never
runs `update`, per ADR-005. It is registered for Claude Code, Codex, and OpenCode; Antigravity has
no session event.

The notice does not gate anything. A verb's own preflight run stays unconditional, and both always
run — the notice is early, and the verb's run is what the verb acts on. Preflight gains one line for
this, `test=undeclared`, `test=unequipped`, or `test=equipped`, read with jq and changing no exit
code.

## Consequences

- **The spec is the floor and not the ceiling.** A case that could only restate spec criteria would
  make testing a transcription exercise, and a case free to invent would drift back toward the code.
  `derived` is what holds both open: it is allowed, it is marked, and it cites what it elaborates,
  so every criterion in a case answers "where did this come from?" in its own line.
- **The expansion feeds the spec, and that costs a round trip.** A gap found while writing a case
  goes back to `codefall-specify`, which is slower than editing the case alone and is the only way
  the spec stops being wrong. Append-only numbering means the correction never disturbs a citation
  already written into a test title.
- **Backfill is not in this version.** Authoring runs forward from a spec through design and
  implement. A project adopting codefall onto code that already shipped has no route to cover that
  code until the backfill procedure lands, and until then its cases cover new work only.
- **Nothing is inferred at run time.** The root, the runners, and the commands are all declared, and
  a project that has declared none gets a report naming `codefall-equip` rather than a guess. The
  cost is one explicit step on adoption, which doctor and `codefall init`'s report both name.
- **Setting up the harness can block a task.** Because setup is its own pull request, a bead whose
  criteria call for a case stops when the project is unequipped. That is deliberate — a task's pull
  request that also installs a test runner is two changes in one review — but it means a first spec
  can reach implement and go no further until someone runs equip.
- **Two documents describe the runner and can disagree.** Settings names it and `<root>/AGENTS.md`
  carries its commands. Splitting by reader is what keeps a program from parsing prose, and the
  price is that a project can declare `playwright` in settings while `<root>/AGENTS.md` still names
  an older command. Nothing checks the pair; equip writes both in one pass, which is the only thing
  keeping them level.
- **Committed run reports are a new kind of file in a project's history.** They make variance
  countable across runs rather than remembered, and they add churn under `.codefall/tests/`; the
  `.ignore` entry keeps them out of a ripgrep-backed search the way review findings already are. The
  bulky output stays git-ignored, so what is committed is prose about a run and never the run's
  artifacts.
- **The runner-specific checks do not generalize.** The two Playwright rules earn their place
  because one of them catches a check that silently reports a failure as an absence. A second runner
  brings its own checks or none, and the shared check script is what every project gets either way.
- **The two verdict vocabularies stay apart.** `PASS`/`FAIL`/`ERROR`/`PARTIAL` describes a run a
  person's judgement stood behind; a spec has its runner's pass and fail and nothing else. Mixing
  them would suggest the two modalities produce the same kind of evidence, and they do not.
- **The case-file format and the `test` block are a breaking-change surface.** Once a project has
  case files, renaming a frontmatter key, changing what a modality means, moving the test-cases
  directory, or renaming a field in the block is a change that project answers by hand, and carries
  `!` and a `BREAKING CHANGE:` footer per the repository `AGENTS.md`.

## Related

- `docs/decision-log.md`, *The testing decisions, 2026-09-17* — the alternatives seen and not taken,
  and what the model project's record did and did not supply.
- `docs/PLAN.md` — the landing order; the implementation follows in later pull requests of this
  stack.
- ADR-005, *Local Environment Scripts* — the split this record follows between what codefall owns
  and what the project declares, the point-of-introduction rule, and the rule that a hook never
  pulls.
- ADR-006, *Install Layout* — why `check-cases.sh` is installed once under `.codefall/shared/`.
- `extensions/skills/codefall-specify/reference/specification.md`, *Numbering* — the identifier form
  a citation uses, and the append-only rule a reported gap relies on.
- `cli/schemas/settings.schema.json` — where the `test` block is published.
- `extensions/skills/codefall-test/NOTES.md` — the lineage, naming `demo-flights` ADR 0011 as the
  origin of the case-file design.
- Repository `AGENTS.md`, *Workflow* — the definition of a breaking change this record relies on.
