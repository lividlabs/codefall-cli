---
name: codefall-equip
description: Equip a project with the two local-environment scripts codefall-refresh runs — start, which brings the services the project develops against up, and update, which makes the local environment match the checkout — by finding what the project already has, or drafting them from what the repository shows, confirming, and declaring them under local in .codefall/settings.json. Builds and rebuilds; codefall-scaffold and codefall-implement follow it as the procedure.
argument-hint: "[path]"
disable-model-invocation: true
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Write
  - Edit
  - Bash
---

# Equip

Equip a repository with the tools it needs to run locally: a `start` command that brings its
services up, and an `update` command that makes the local environment match the checkout. Both are
the project's own scripts, declared under `local` in `.codefall/settings.json`, and
`codefall-refresh` is what runs them from then on.

Equipping is not refreshing. This skill finds, drafts, or revises the scripts and declares them; it
runs them only to prove them. Bringing an environment current on an ordinary day is `codefall-refresh`.

Paths that start with `reference/`, `templates/`, or `../` are relative to this skill's directory,
not the user's project.

## Files beside this one

Read each when its step says to; none is loaded up front.

- `reference/signals.md` — what in a repository says which tools it uses, the entry points a
  project may already have, and what each maps to in `start` and `update`. Read at step 2.
- `templates/local.sh` — the default shape of the script when the project has no task runner of
  its own: one file, two subcommands. Read at step 3.

## Scope — the scripts, not the environment

| In scope | Out of scope | Whose |
| --- | --- | --- |
| Finding the scripts a project already has | Running them on an ordinary day | `codefall-refresh` |
| Drafting them when there are none | Pulling or comparing the checkout | `codefall-refresh` |
| Revising them when the project's tools change | Deciding which tools the project uses | `codefall-design` |
| Declaring them in `.codefall/settings.json` | Per-branch databases, shared or otherwise | the project |
| Proving a candidate meets the contract | Any other settings field | `codefall init` |

**Every run does the same thing**, whether the scripts exist or not: read the project, read the
scripts if they are there, propose the draft or the revision, confirm, write, declare. There is no
first-time mode.

## The contract

What the declared commands promise. A candidate that does not keep it is reported and never
declared; a draft that would not keep it is not written.

- Both are **shell commands run from the project root**, declared as strings. A Makefile target, a
  package script, or a script of the project's own are all fine.
- **`start`** brings up what the project needs running locally to develop against. Exit `0` when
  everything it manages is up, whether it started it or found it running. Safe to call twice; the
  second call is cheap.
- **`update`** makes the local environment match the checkout: dependencies to the lockfile,
  pending migrations applied, generated code regenerated, whatever a changed definition
  invalidates rebuilt. Exit `0` when the environment matches. Safe to call when nothing changed,
  and cheap then — it asks the package manager and the migration tool, which already answer
  "nothing to do" quickly, rather than doing the work unconditionally.
- **`update` may assume `start` has run.** It does not start anything; `codefall-refresh` runs
  `start` first.
- **Plain scripts.** A person runs either from a terminal. CI can run either. Neither needs the
  extension or a harness. A non-zero exit and a sentence on stderr are how one reports a problem,
  written for the person who reads it next.
- **Never destructive.** No dropping a database, no deleting data directories, no `--force`
  recreation. A command that resets state to get to a known state is not idempotent; it is
  starting over every time.

## Search first, then ask with evidence

On an existing project the scripts usually exist under some name already. Read the repository
before asking anything, name what was found, and ask one question:

> I found `make dev-up`, which starts Postgres and Redis through compose, and `npm run db:migrate`.
> Are these the start and update commands, or should I draft new ones?

A user who knows corrects it. A user who does not know takes what was found. A repository with
nothing gets the draft offer. **Never ask whether scripts exist before looking**: the repository
answers that faster than the person, and the person may not know.

## When another verb follows this skill

`codefall-scaffold` and `codefall-implement` read this file as the procedure; nothing invokes it
through the harness.

- **`codefall-scaffold`, at runnable-skeleton depth.** Nothing exists to search for; skip step 2.
  Draft from what scaffold emitted — its manifest, its compose file if any, its migration tool —
  declare, and prove the scripts as part of scaffold's own verification. At docs-only depth write
  nothing and name this skill in scaffold's report as owed work.
- **`codefall-implement`, on a task that introduces infrastructure, a dependency, a migration, or
  generated code.** The bead's acceptance criteria name the script change; the bead is the
  confirmation. Read the declared scripts, revise them per step 3 for what the task introduced,
  keep the contract, and change the declaration only when an entry point moved. The change lands
  in the task's own pull request and is named in its body.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

### 1. Read the project

The target is the path argument, or the working directory. Read `.codefall/settings.json` and note
whether a `local` block is declared; read the declared scripts when it is. Read the project's
`AGENTS.md`, root and scoped, for verify commands and conventions. If there is no
`.codefall/settings.json` at all, stop and say `codefall init` comes first.

### 2. Find what is already there

Read `reference/signals.md`. Two searches:

- **Entry points** the project already has — Makefile and justfile targets, package scripts, a
  `scripts/` directory, a Taskfile or mise tasks, a `bin/setup` — whose names or bodies say start,
  up, setup, bootstrap, migrate, sync, dev.
- **Signals** of the tools the environment needs — lockfiles, a compose file, a migrations
  directory, a codegen config, an `.env.example` — which decide what a draft has to do and what a
  candidate has to cover.

Check each candidate against [the contract](#the-contract) by reading it, not by running it. A
candidate that resets, drops, or deletes fails; say which line and why.

Then ask the one question, with what was found beside it. Three answers:

- **Declare what exists.** Go to step 4 with the candidates as they are.
- **Draft new.** Go to step 3 with the signals.
- **Revise what is declared.** The block already exists and the user says the tools changed. Go to
  step 3 with the declared scripts open.

### 3. Draft or revise

Read `templates/local.sh`. The shape follows the project:

- A project with a **task runner idiom** — a Makefile, a justfile, package scripts everyone runs —
  gets two targets in that idiom, so the commands read like the project's others.
- A project with **none** gets `scripts/local.sh` from the template, with `start` and `update` as
  its subcommands and the declaration pointing at each.

Each step in `update` is the tool's own idempotent form: `npm ci` not `npm install`, `migrate
deploy` not `migrate reset`, `go mod download`, `uv sync`. An `.env` is copied from `.env.example`
only when missing, never overwritten. `start` uses the compose file's own wait (`up -d --wait`)
where it has one. Every step carries a one-line comment saying what it brings current.

A **revision** changes only what the introduced tool needs and leaves the rest of the script as
the project wrote it. Show the diff, not the whole file.

Show the full draft, or the diff, and confirm before writing anything.

### 4. Declare

Write the `local` block into `.codefall/settings.json`, keeping every other key and the file's
formatting:

```json
"local": {
  "start": "scripts/local.sh start",
  "update": "scripts/local.sh update"
}
```

Commands are relative to the project root. A script of the project's own is made executable.

### 5. Prove them

Offer to run them; running `start` brings real services up on the user's machine, so it is a
yes-or-no rather than a step. On yes: `start`, then `update`, then `update` again. The second
`update` must exit `0` quickly, and that is the proof of the contract. A failure is fixed in the
script and the proof re-run; a failure that is the environment's — Docker not running — is
reported as such with what to do.

Then run `codefall doctor` when the CLI is on `PATH` and read its **Local environment** section:
both checks pass, or the declaration is wrong and step 4 is repeated. A warning about the refresh
stamp not being git-ignored is `codefall init`'s to fix; name it in the report.

### 6. Report

- What was declared, and whether it was found, drafted, or revised.
- Whether the scripts were proven, and by which run.
- What the user still owes the project: commit the scripts and the settings; run
  `/codefall-refresh` once so the stamp exists; `codefall init` if doctor warned about
  `.gitignore`.

## Rules

- **Search before asking.** Name what was found; never open with "do scripts exist?"
- **The contract is not negotiable.** A candidate that drops, resets, or deletes is reported and
  never declared, however convenient it is.
- **Nothing is written without confirmation.** The full draft or the diff, shown first.
- **Every run is the same procedure**, first draft and revision alike. No one-time mode.
- **Revise the smallest thing.** A revision touches what the introduced tool needs and leaves the
  project's own lines alone.
- **Declare relative to the project root.** Commands, not paths; a task runner's target is a
  command.
- **Proving runs real services, so it is offered.** Never started on the user's machine unasked.
- **Only the `local` block.** No other settings field is this skill's to touch.
- **Refresh is not this skill.** Say so and stop when the user wants the environment brought
  current rather than equipped.
