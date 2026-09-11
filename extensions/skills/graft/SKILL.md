---
name: graft
description: Bring a project's codefall documents up to date with the current templates — compare what the plugin ships now against what the project has, report every difference with its provenance, and apply only the pieces the user takes, one at a time. Works on projects scaffolded before provenance existed, and on repos adopting the stance for the first time.
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

# Graft

A project scaffolded at codefall 0.2.1 never receives anything the templates gained since — new
ADRs, revised rules, renamed files all live in the plugin, not in the project. Graft closes that
gap: it reads what the project has, compares it against what the plugin ships now, reports every
difference, and applies exactly the pieces the user takes.

The name is the contract. You graft onto **rootstock** — an existing, living project that is not
being replaced. You graft one **scion** at a time — application is per-item, never wholesale. And a
graft either **takes or is rejected** — a revised template meets a project that amended its
ancestor, and the amendment wins.

This skill is **only ever invoked explicitly**, and more firmly than the others: `scaffold` runs on
an empty directory, but graft runs on a project people depend on. Never suggest it, never fire it
from a passing remark, never chain into it from another skill. Someone types `/graft` on purpose or
it does not run.

Template paths in this document are relative to `../scaffold/templates/`, resolved from this skill's
directory — the one holding this `SKILL.md` — because graft reasons about `scaffold`'s templates and
has none of its own. Its one bundled reference is `lineage.md` beside this file, the record of what
every current template used to be called. Neither path is relative to the user's project.

## Scope — documents, not code

Graft moves **documents**. It does not restructure code.

| In scope | Out of scope |
| --- | --- |
| Inherited ADRs — new ones; revisions, landed as superseding ADRs; renames | Moving code between architectures or topologies |
| `docs/adrs/_TEMPLATE.md` | Extracting a component into a service |
| `AGENTS.md` skeleton drift, reported | Re-slicing component boundaries |
| `.codefall/scaffold.json` — written or brought current | Touching source, build config, or lint rules |
| A `docs/decision-log.md` line recording the graft | The project's own `ADR-NNN` decisions |

Moving a project from ports-and-adapters to package-by-component, or pulling a component out into
a service, are real jobs — and they are code refactors with a different blast radius. They belong
to a future `migrate` verb. If the user asks for one, say that plainly and stop; do not do a small
version of it here.

The project's own decisions — the bare-numbered `ADR-NNN` sequence — have no template behind them
and are never graft's to update. They enter the picture only when a rename leaves them citing an
old identifier, and even then graft reports the stale reference rather than silently editing a
decision the project wrote.

## Provenance states

`.codefall/scaffold.json` records, per inherited ADR, an `amended` flag and a `sha256` (see
`scaffold` step 4). Together they separate four states, and the state decides everything graft may
do:

| State | Meaning | Graft may |
| --- | --- | --- |
| **untouched** | hash matches the recorded one, `amended: false` | supersede or rename it when the user takes the item |
| **amended** | `amended: true` — changed during the original interview | report the diff, leave it alone |
| **edited** | hash no longer matches — changed since scaffolding | report the diff, leave it alone |
| **unverifiable** | no provenance and no old template reachable | treat as edited |

**Only untouched is ever safe to take mechanically.** Taking a revision never rewrites the
ratified file — it lands as a new, superseding ADR (step 5) — but building that successor from the
current template is only correct when the template fully accounts for what the project has. For an
amended or edited ADR it does not: the project changed the decision on purpose, and a
template-built successor would silently drop that change from what governs. So for those two the
diff *is* the deliverable — authoring the superseding ADR that carries their amendment forward is
the user's work, and after they write it they can run graft again.

There is one narrow exception: a user who has seen the diff may explicitly say *supersede my
version with the template's*. That is their call, and supersession keeps their version on the
record — but their amendment stops governing, so say exactly that, require them to name the file,
and only then land it like any other taken revision. Never offer this as the convenient path.

## Where the old templates come from

With `.codefall/scaffold.json` present this question does not arise: the recorded hashes classify
every file without any historical template. It arises for projects scaffolded before provenance
existed, and there the answer is a ladder — take the first rung that works:

1. **The local plugin cache** — `~/.claude/plugins/cache/<marketplace>/codefall/<version>/`. Exact
   snapshots, offline, but only of versions this machine actually installed.
2. **Git history** — if the plugin root, two levels up from this skill's directory, sits inside a
   clone that can reach the release tag, `git show <tag>:<path>` works. Installed marketplace
   clones are usually **shallow** with few or no tags, so try `git fetch --depth=1 origin tag <tag>`
   before concluding the tag is missing. Tags come in two forms and old paths differ from current
   ones — `lineage.md` records both.
3. **GitHub** — `gh api repos/lividlabs/codefall-plugin/contents/<historical-path>?ref=<tag>`,
   or the raw URL. Needs network.
4. **Nowhere** — then the file is **unverifiable**. Say so and treat it as edited.

**Never reconstruct an old template from memory of what it probably said.** A plausible-looking
reconstruction poisons every classification built on it: it marks edited files untouched and
untouched files edited. The ladder or nothing.

When comparing a project file against the template that emitted it, normalize the one thing
`scaffold` changes on emission: the date on the `## Status` line (`Accepted — <date>` became
`Accepted — 2026-08-12`). Everything else compares verbatim. Equal after that means untouched;
different means edited — without provenance you cannot distinguish *amended at the interview* from
*edited since*, and edited is the safe reading, so collapse to it.

## Process

### 1. Read the project — the gate

The target is the path argument, or the current directory if none was given. Name what you found —
the project, not just the path — so the user can redirect before anything else happens. If the
target shows no sign of being a project root (no VCS, no manifest, no docs), stop and ask rather
than scanning onward.

Inventory the rootstock: `.codefall/scaffold.json`, `docs/adrs/`, `docs/decision-log.md`, and every
`AGENTS.md`. That yields one of three starting points, and the rest of the process is the same for
all three — they differ only in how much step 2 must reconstruct:

- **Provenanced** — `scaffold.json` exists. The normal case from codefall 0.4.0 on.
- **Scaffolded, pre-provenance** — no `scaffold.json`, but `docs/adrs/` holds codefall-lineage
  files (current identifiers, or historical ones per `lineage.md`). The decision log's *scaffolded
  with codefall `<version>`* line, when present, names the baseline version.
- **Never scaffolded** — no codefall docs at all. This is **first-time adoption**: every applicable
  template is simply missing, and the same report-then-take flow installs the stance. Detect the
  surfaces from the repo (`package.json`/`tsconfig` suggests `typescript-react`, `go.mod` suggests
  `go`) and confirm rather than assume — the profile catalog and its refusal rule are `scaffold`'s.
  If a surface has no supported profile, refuse the same way `scaffold` does: say plainly that the
  stack isn't supported yet, name what is, offer to record the request, and stop. Do not improvise.

Adoption installs documents describing a stance the existing code does not yet follow, and does not
check the code against them — say that in the report rather than letting the docs imply otherwise.
Aligning the code is `migrate`'s job, and the boundary-enforcement obligation lands on the user
exactly as it does after a docs-only scaffold.

Steps 1–4 only read, so a dirty working tree is fine for the report. Note it, and require a clean
tree — or an explicit go-ahead — before step 5 writes anything: every application should be
reviewable as a git diff on its own.

### 2. Establish provenance

**With `scaffold.json`:** verify, don't trust. Recompute each listed file's sha256 and compare to
the recorded one: match with `amended: false` is untouched; `amended: true` is amended; mismatch is
edited. A listed file missing from disk was deleted by the project — report it as that, and leave
the entry alone. Never "correct" the file to make drift disappear; it records history, not
configuration.

A hand-backfilled `scaffold.json` — one written after the fact rather than by `scaffold` — is only
as safe as its `amended` flags. The hashes are re-verified right here, but `amended` is taken as
recorded, and a flag wrongly set to `false` marks a customized file as safe to supersede. Nothing
downstream can catch that lie; a backfill unsure about a file should say `true`, whose worst case
is a diff shown instead of an update offered.

**Without `scaffold.json`:** identify each codefall-lineage doc via `lineage.md`, find the template
that emitted it using the ladder above, and classify by normalized comparison. When the decision
log doesn't name the baseline version, there are only a handful of releases a given filename can
come from — `lineage.md` names them; compare against each. What can't be classified is
unverifiable, and unverifiable is edited. Do not guess a version, and never invent a hash.

### 3. Compare against the current templates

First fix the **applicable set** — which templates this project should have at all. Profiles and
decisions come from `scaffold.json` when present, from which profile's ADRs are on disk otherwise,
and from step 1's detection for adoption. Apply `scaffold`'s gates, not your own: a surface that
can never be split skips ADR-BASE-03, a backend-only project has no use for ADR-TS-02. If a gate
genuinely can't be answered from the project, ask — once, batched with anything else step 4 needs.

Then classify every difference:

- **Missing** — an applicable template with no counterpart in the project. New since the project's
  baseline, or all of them under adoption.
- **Revised** — the counterpart exists and the current template's content differs from what the
  project has.
- **Renamed** — the counterpart exists under a historical identifier. `lineage.md` is the record:
  a project on 0.2.x has `ADR-003-dependency-injection.md`, and that *is* `ADR-TS-01` — **recognise
  it as a rename, never as a deletion plus an addition.** Renames compose with revisions; report
  both facts about the one file. A rename also implies stale references wherever other docs cite
  the old identifier — find them with Grep, list them with the item.
- **Retired** — a project doc whose template no longer ships. The decision didn't evaporate because
  the plugin moved on; report it, leave the file alone.
- **Skeleton drift** — the profile's `AGENTS.md.skeleton` changed. An `AGENTS.md` is filled in per
  project at scaffold time, so it is *always* effectively amended: skeleton changes are reported as
  advisory, never applied wholesale.
- **Provenance** — `scaffold.json` itself is missing or stale. Offer to write it; that is how a
  pre-provenance project stops needing the ladder next time.

### 4. Report — and stop

The report is graft's default output, and for many runs its only one. For each item: what it is,
its provenance state, what changed and why — one line each, with the diff itself for anything
amended or edited, because there the diff is the deliverable — and whether it can be taken
automatically or is the user's to merge. Say what taking means: a revision lands as a superseding
ADR, never as a rewrite of the one the project ratified.

When the plugin's changelog is reachable (a repo checkout, or GitHub — the installed plugin subtree
does not carry it), use it to say *why* things changed, grouped by release since the project's
baseline. When it isn't, say the report comes from template comparison alone. Either way the
comparison, not the changelog, is the source of truth for *what* changed.

Then **stop**. Nothing is applied unrequested — a report that quietly rewrote files defeats the
verb. The user takes items by naming them, "everything safe" included; present the choice with one
question, items grouped by whether they can be taken automatically.

### 5. Apply what the user takes — one scion at a time

Only requested items, and only missing or untouched ones are ever taken. **Graft never rewrites a
ratified ADR.** A taken revision lands as a new ADR that supersedes the old one, which stays on
the record; the only in-place edit graft ever makes to an existing ADR is flipping its Status line
to `Superseded by <id> — <date>`, which is the edit the discipline prescribes. Git history is not
the record here — the documents are.

- **Missing** — instantiate from the current template, exactly as `scaffold` step 4 emits it:
  stamp today's real date (`date +%F`, not memory) on the Status line, keep the flat `docs/adrs/`
  layout.
- **Untouched, revised** — write the current template as a **new, superseding ADR** stamped with
  today's date, and flip the old one's Status. When the revision arrived together with a rename,
  the new file simply takes the current template's name — `ADR-TS-01-dependency-injection.md`
  superseding `ADR-003-dependency-injection.md`, which keeps its old name and its place. When the
  name hasn't changed, suffix an edition so both can exist:
  `ADR-BASE-02.2-package-by-component.md`, identifier `ADR-BASE-02.2`, editions counted within
  this project. Cross-link the pair in Related — the successor names what it supersedes.
- **Renamed, unrevised** — when only the identity moved and the decision didn't, a move is
  addressing, not substance: `git mv` where the tree is git-managed, update the identifiers the
  file itself carries, and say so in the report. In practice renames almost always arrive with a
  revision, and then the supersession above already produces the new name.
- Either way a rename strands references. Fix ones in *untouched* inherited docs as part of the
  item. Stale references in amended or edited docs, in `AGENTS.md`, or in the project's own ADRs
  are theirs: list each one, offer the targeted edit, and make it only on a yes.
- **Amended / edited** — never superseded by taking an update, even one that renames the file
  around them. The narrow, user-named exception is in [Provenance states](#provenance-states).
- **`AGENTS.md`** — targeted section edits on request only, never a wholesale replacement.
- **`.codefall/scaffold.json`** — after every applied item, update its entry: `id`, `file`, and a
  `sha256` computed with `shasum -a 256` on the file as written — never invented. A supersession
  touches two files: add an entry for the successor, which is what future runs compare against,
  and recompute the superseded file's hash after its Status flip. Add
  `"lastGraft": { "pluginVersion": "<version>", "date": "<date>" }` at the top level, reading the
  version from `../../.claude-plugin/plugin.json` — do not guess it. Where the file didn't exist,
  write it fresh: this plugin version, today's date, the profiles and decisions as established, and
  an `amended` per ADR — `false` only for files that now hash-match a current template, `true` for
  anything kept that differs.
- **`docs/decision-log.md`** — one line under `Locked`: *grafted codefall `<version>` on `<date>` —
  took ADR-BASE-03, ADR-TS-01; left ADR-TS-02 (amended)*. Humans read this; `scaffold.json` stays
  authoritative.

A partial graft is a normal outcome, not a failure state. Per-file hashes carry the truth about
what is current, so the next run of graft needs no memory of this one.

### 6. Verify

Whatever was written must actually hold together. Every link in every touched doc resolves.
`scaffold.json` parses, and every hash in it matches the file on disk — recompute, don't assume.
No doc still cites a renamed identifier except the ones the user chose to leave — Grep for the old
IDs and check the hits against that list. Every Status line that says Superseded names an ADR that
exists, and every successor's Related names what it superseded. Then show the git diff summary:
the graft should be reviewable as one coherent change.

### 7. Report what took

- What was **taken**, file by file.
- What was **left**, and why — amended, edited, retired, or declined.
- Stale references the user chose to keep, so they aren't rediscovered as a surprise.
- What the user still owes the project: the hand-merges they said they'd do, and — after an
  adoption — the boundary-enforcement obligation, named exactly as `scaffold` names it after a
  docs-only run.

## Conventions

Repo-wide rules — verb naming, template lineage upkeep, ADR immutability — live in the plugin
repo's root `AGENTS.md`. Specific to this skill:

- `lineage.md` is required: any change that renames, moves, or retires a template ships a row
  in it, in the same PR. Graft can only tell a rename from a deletion because that record exists.
- Graft writes nothing outside the target project. The plugin's own files — templates, lineage,
  this skill — are maintained through PRs, not by a run of graft.
