# ADR-005: Local Environment Scripts

## Status

Accepted — 2026-09-16

## Context

Pulling the default branch before starting work has consequences that the pull does not perform. A
migration the local database needs, a dependency to install, code to regenerate, a container to
rebuild — each arrives as a diff, and a teammate who does not read diffs has no way to know which
of them arrived. Codefall exists in part so that teammates who do not read diffs can still do the
work, so the gap has to close on their behalf.

Skills in other projects are custom built to fit the shape of the project, so each knows
that project's package manager, its migration tool, and its containers. Codefall serves
many project shapes, and a skill that knows one shape does not generalize. The question is what
codefall can own about this without knowing the project, and what shape the project-specific
remainder takes.

The job splits along that line. **Bringing the checkout current** — fetch, compare against the
default branch, notice a dirty tree — is the same for every project. **Bringing the environment
current** is not: it depends on the stack, the tools, and the services the project runs locally.
Codefall can own the first outright. For the second it can only own the shape of the answer and
the moments at which the answer is called.

Three shapes were considered for the project-specific part.

**(a) A skill that does the environment work itself**, re-deriving what to run from the repo on
every invocation. This is the per-project skill above, generalized. It varies between runs, costs a model
turn to decide what a lockfile already says, and can be run only from a harness with the extension
installed — not by a person at a terminal, not by CI, and not by a harness codefall does not
support.

**(b) A script codefall generates at run time**, written fresh by a model whenever it is needed.
This has every cost of (a) plus a file that changes under the project without a diff anyone
reviewed.

**(c) Scripts the project owns**, drafted once by a model from what the repo shows, confirmed by a
person, committed, and revised in the same PR as the change that makes them stale. Codefall knows
where they are because the project declares them, and knows what they promise because the promise
is fixed here. Anyone and anything can run them.

Option (c) is chosen. What it needs from this record is the contract: the declaration, the two
commands, what each promises, and how codefall knows whether one has run. Once any project
declares the block, changing any of these is a breaking change by the definition in the repository
`AGENTS.md`.

Two things the contract deliberately does not settle are named in the Consequences.

## Decision

### The declaration

`.codefall/settings.json` gains an optional `local` block with two string fields, each a shell
command run from the project root:

```json
"local": {
  "start": "scripts/local.sh start",
  "update": "scripts/local.sh update"
}
```

The values are commands, not paths: a project may point at a `Makefile` target, a package script,
or anything else it already has. Both fields are required when the block is present. `doctor`
reports the block missing, a field missing, or a command whose program is not on `PATH`.

### What `start` promises

`start` brings up whatever the project needs running locally to develop against — a database, a
queue, a container set. It exits `0` when everything it manages is up, whether it started them or
found them already running. Calling it twice in a row is safe and the second call is cheap.

### What `update` promises

`update` makes the local environment match the checkout it runs in: installs dependencies to the
lockfile, applies pending migrations, regenerates generated code, rebuilds what a changed
definition invalidates. It exits `0` when the environment matches. Calling it when nothing changed
is safe and cheap — it asks the package manager and the migration tool, both of which already
answer "nothing to do" quickly, rather than doing the work unconditionally.

`update` may assume `start` has run. It is not required to start anything itself, and codefall runs
`start` before it.

### Both are plain scripts

A person can run either from a terminal. CI can run either. Neither depends on the extension being
installed or on a harness being present. A non-zero exit and a message on stderr are how one
reports a problem; the message is written for the person who will read it next, since a harness
will quote it to them.

### The stamp

`.codefall/refresh.stamp` holds the commit hash at which `update` last exited `0`. It is written by
codefall after a successful run, never by the script, so the script stays plain. It is machine
local and git-ignored — `codefall init` adds the entry — and it lives in the working directory, so
each worktree carries its own.

The stamp answers one question: has `update` run since `HEAD` last moved? A stamp that matches
`HEAD` means yes and the run is skipped. Anything else — missing, or a different hash — means run
it. Because `update` is idempotent, a stale stamp is never wrong to act on, only wasteful, and a
hand run of the script that leaves the stamp behind costs one cheap re-run at the next session.

### Who writes and maintains them

The scripts are drafted by `codefall-equip`, which reads the repo, reads the scripts if they already
exist, proposes the draft or the revision, and writes on confirmation. A project that already has
scripts doing this work does not get new ones: equip names what it found, and on confirmation
declares those commands as they are — provided they meet the promises above. A candidate that does
not, one that drops and recreates a database on every run, say, is reported and never declared. `codefall-scaffold` follows
the same procedure when it emits project files. `codefall-implement` follows it when a task
introduces infrastructure, a dependency, a migration, or generated code, in the same pull request
as that change. The scripts are project documents: once written, codefall reports drift and never
overwrites them.

### When they run

`codefall-refresh` runs `start` when the services are down, runs `update` when the stamp does not
match `HEAD`, records the stamp, and turns a failure into a plain sentence. It is the routine
action. The check that precedes it — fetch, commits behind the default branch, dirty tree, stamp
against `HEAD` — is read-only and runs in the shared preflight every verb already calls. No hook
ever pulls or runs `update`: a hook cannot ask, and a pull moves the working tree under whatever
the session is doing.

## Consequences

- **The environment is always one command from current, for anyone.** A teammate who does not
  read diffs types `/codefall-refresh` and gets either a current environment or a sentence saying
  what to fix. A person at a terminal runs the same script.
- **Idempotency is a requirement, not a nicety.** It is what makes "bring the project up to date?"
  always answerable with yes, what makes the stamp safe to distrust, and what makes running `update`
  after every fetch acceptable. A drafted script that does unconditional work — drops and recreates
  a database, say — violates the contract and `codefall-equip` must not write one.
- **Staleness has one detector.** The stamp compares against `HEAD` and nothing finer. A commit
  that touched only documentation still triggers a run, which the idempotency requirement makes
  cheap. A finer detector — hashing the lockfile and the migrations directory — is a later
  refinement that changes nothing declared here.
- **The scripts change when the project changes, or they rot.** The contract holds only if the
  point-of-introduction rule is followed: the task that adds a queue also changes `start`, and the
  task that adds a migration tool also changes `update`. `codefall-design` names it in the task,
  `codefall-implement` counts it toward done, and `codefall-review` carries a lens for it. A project
  whose scripts fall behind gets a `refresh` that exits `0` and an environment that is not current,
  with nothing to say so.
- **Shared databases across worktrees are the project's problem.** `codefall-implement` builds in
  worktrees, but a local database is usually one per machine. A migration applied from a feature
  branch leaves the primary checkout's environment ahead of its code, and `update` on that checkout
  cannot undo it. The contract says only that `update` brings the environment level with the
  checkout it runs in; whether a project uses one database per branch is its own call.
- **Nothing is inferred at run time.** Codefall reads a declared command and runs it. A project
  that has not declared one gets a report naming `codefall-equip`, never a guess. The cost is one
  explicit step on adoption, which `codefall init`'s report and `doctor` both name so the adopter
  takes it in the adoption PR rather than the next teammate taking it by surprise.
- **The block is a breaking-change surface.** Renaming a field, changing what either command
  promises, or moving the stamp is a change every equipped project has to answer by hand, and
  carries `!` and a `BREAKING CHANGE:` footer per the repository `AGENTS.md`.

## Related

- `docs/decision-log.md`, *Local environment: equip and refresh, 2026-09-16* — the alternatives
  seen and not taken, including the shapes the drafting verb went through, and the names rejected.
- `docs/PLAN.md` — the landing order.
- `cli/schemas/settings.schema.json` — where the `local` block is published.
- `extensions/shared/preflight.sh` — where the read-only check lands.
- Repository `AGENTS.md`, *Workflow* — the definition of a breaking change this record relies on.
