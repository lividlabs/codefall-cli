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

- Named as **verbs** (`codefall-scaffold`, `codefall-graft`), one directory each:
  `skills/<verb>/SKILL.md`.
- Every skill carries `disable-model-invocation: true`. Running one is a deliberate act.
- A skill reports and offers; it applies only what the user takes. Nothing lands unrequested.
  Recording an observable fact is the exception: a skill that owns a status transition sets it when
  the fact occurs and reports that it did — `codefall-implement` flipping a concept to `Active` at first
  claim is this shape. Judgment transitions — promote, archive, revise — stay offer-only.
- **Refuse only what you cannot do.** A missing surface profile, a missing tool, an unsupported
  tracker — those are exits. Disagreeing about size, altitude, or fit is not: say what you think and
  why, then do what the user asks. `codefall-conceptualize`'s floor and `codefall-specify`'s cohesion check are both
  this shape.
- Never present an option that would be refused — unsupported stacks and planned profiles are
  exits, not menu choices. See the stack question in `codefall-scaffold`'s SKILL.md.
- Every verb reads `.codefall/skills/<verb>/CUSTOMIZE.md` from the user's project when it exists —
  project procedure the extension cannot know. The procedure lives in
  `extensions/shared/customizations.md`; a skill points at it and never restates it.
- Extension paths inside a skill are relative to the skill's own directory — `../../shared/…`,
  `../scaffold/templates/…` — never `${CLAUDE_PLUGIN_ROOT}`.
  Harnesses that mirror the tree under `.agents/skills/` do not define that variable, and the
  relative form resolves under them and under Claude Code alike. Hooks are the exception, and the
  rules for them live in `hooks/`: `hooks/shared/` holds the script every harness's guard runs,
  and `hooks/<harness>/` holds the definition `codefall init` registers — never copied as files.
- **Whose document is it** decides who repairs it. A file the extension ships that nobody amends — the
  operative rules a verb installs alongside a directory it owns, like `docs/concepts/AGENTS.md` — is
  repaired by the verb that owns it, on run. A template that becomes the project's own document, one
  that gets stamped, amended, and cited, belongs to `codefall-graft`, and ships a row in its scope table and in
  `lineage.md`. Do not route a fixture through `codefall-graft`: it buys consent machinery for a decision with
  no stakes, and costs a registration step whose failure is silent. Either way, **never overwrite a
  file that has drifted** — show the difference and ask.
- Renaming, moving, or retiring a template ships a row in
  `skills/codefall-graft/lineage.md`, in the same PR. `codefall-graft` can only tell a rename from
  a deletion plus an addition because that record exists.
- New skill prose matches the established register — declarative, reasons attached, refusals
  stated plainly. Read `codefall-scaffold`'s SKILL.md end to end before writing one.

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
