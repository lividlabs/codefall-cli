# ADR-006: Install Layout

## Status

Accepted — 2026-09-17

## Context

`codefall init` installs the extension for each harness a project chose. Codefall supports five of
them, and they read from two directories: Claude Code reads `.claude/`, while Antigravity, Codex,
Muse, and OpenCode all read `.agents/`. The extension step copies the embedded tree — everything
except the per-harness hook definitions, which the hook step reads straight from the binary — into
each of those directories, once per directory. A project that uses Claude Code together with any
other harness receives the whole tree twice.

Measured on 2026-09-17, the tree is 64 files under `skills/` at about 375 KB, and seven files
outside it at about 28 KB: the shared merge guard, the three files under `shared/`, and three
maintainer documents, `README.md`, `AGENTS.md`, and `docs/ROADMAP.md`.

Two facts about that tree decide what a layout can be.

**Skills are the only part a harness finds by convention.** A harness looks in `.claude/skills/` or
`.agents/skills/` and loads whatever is there. Codefall does not tell it where to look, and cannot
move that directory without the harness losing the skills.

**Everything else is reached by a path codefall writes.** Each harness's hook definition names the
merge guard in full: Claude Code's names `.claude/hooks/shared/codefall-block-merge-to-main.sh` from
the repository root, Codex's and Antigravity's name the same file under `.agents/`, and OpenCode's
plugin reaches it from its own directory through `import.meta.dir`. Every skill that reads a shared
file names it relative to the skill's own directory — `../../shared/…`, twenty-one references today.
Both ends of each of those paths are codefall's, so a file reached that way can live wherever this
record says it lives.

The duplication of the second kind can be removed outright. Whether the first kind can be removed
too is what the three layouts below disagree about.

**(a) One real copy in `.agents/`, with `.claude/` linking into it.** The smallest change: one tree
on disk, one link per project that uses Claude Code. It works only while Claude Code follows a
symlinked skills directory, and it leaves the layout asymmetric — one harness's directory is a link
and the others are real.

**(b) One real copy in `.codefall/`, with every harness directory linking into it.** Symmetric, and
it removes every duplicate byte. It makes the same bet as (a) against every harness at once, and
against every harness added later: each has to follow a link out of its own skills directory before
it finds a single skill.

**(c) `.codefall/` holds what is reached by a path codefall writes, and the skills stay copied into
each skills directory.** It removes the duplication codefall controls and leaves the duplication the
harnesses control. Nothing is a link, so no harness's behaviour can break it, and a harness added
later brings one more copy of the skills and nothing else.

Option (c) is chosen. The reason is what it does not depend on: a link followed, by software
codefall does not write, on a platform it does not test. The cost is bounded and known — two copies
of the skills at most, because every harness but Claude Code shares `.agents/`, so a project with
four harnesses installs no more than a project with two.

The layout is also the moment to stop installing what a project never reads, which is the second
decision recorded here.

## Decision

### What lives in `.codefall/`

`.codefall/` is already codefall's own directory in a project: it holds `settings.json`,
`manifest.json`, the refresh stamp, and `reviews/`. It gains the two directories that are reached by
a path rather than by convention:

- `.codefall/hooks/shared/` — the scripts every harness's hooks run.
- `.codefall/shared/` — the files skills read, and the scripts a verb runs.

Each is written once per install, whatever harnesses were chosen, and the scripts stay executable
wherever they land.

### What is copied per skills directory

`skills/` is copied into each chosen harness's skills directory — `.claude/skills/` and
`.agents/skills/` — once per directory, as it is today. A directory two harnesses share is written
once, and each of those harnesses records the same files in the manifest.

### The one path a skill uses for a shared file

A skill names a shared file `../../../.codefall/shared/<file>`. From `.claude/skills/<verb>/` and
from `.agents/skills/<verb>/` that resolves to the same file, so one string is correct for every
harness and the prose does not have to say which directory it is being read from. The path rule in
`extensions/skills/AGENTS.md` changes to match, and `extensions/scripts/skill-health.sh` resolves a
reference beginning `../../../.codefall/` under `extensions/` in the source tree, which is where
that file is before it is installed.

### What the hook definitions name

Every harness's definition names `.codefall/hooks/shared/codefall-block-merge-to-main.sh`. The
prefix logic that writes an install's path below the repository root into a definition that names
the root still applies, and the OpenCode plugin's reach from `import.meta.dir` changes with the
destination.

### Nothing in the layout is a link

Every installed file is a real file. Codefall creates no symlinks and depends on none, in any
harness directory or in `.codefall/`.

### The maintainer documents are installed nowhere

`AGENTS.md`, `skills/AGENTS.md`, `README.md`, `docs/ROADMAP.md`, and every `NOTES.md` beside a skill
stay in the codefall repository. They are written for someone working on codefall — the rules for
changing a skill, the roadmap, the lineage of what a skill took from elsewhere — and a project that
uses codefall has no reader for them. No skill, hook, or script names any of them, so nothing loses
a file it reads.

### The old layout is not cleaned up

Init writes the new layout and re-registers its hooks. Nothing deletes what an earlier version wrote
into `.claude/` or `.agents/`: codefall removes only what it can prove it owns, and those
directories hold the project's own files beside codefall's. The pull request that lands the change
lists what to delete by hand.

## Consequences

- **This is a breaking change.** The file layout changes and so does every hook command, which is
  two of the three examples the repository `AGENTS.md` gives, so the commit carries `!` and a
  `BREAKING CHANGE:` footer. Every project that ran an earlier codefall reruns `codefall init` and
  then deletes the old files by hand. Two or three projects use codefall today and all of them are
  the maintainer's, which is why this lands now rather than being carried.
- **The leftovers are inert but they do not say so.** After a rerun nothing points at
  `.claude/shared/`, `.claude/hooks/`, or the maintainer documents an earlier install wrote, but a
  reader who opens `.claude/shared/preflight.sh` cannot tell from the file that it is dead. The
  clean-up list in the pull request description is the only thing that says it.
- **Doctor's install check needs different evidence.** It stats `<skills dir>/hooks/shared` today,
  which is evidence because nothing but codefall writes that directory. Under this layout it is not
  written per harness at all, and `.codefall/hooks/shared/` exists once for a project however many
  harnesses were installed for, so it cannot answer a per-harness question. The check has to read
  something that is still per harness — the files the manifest records for that harness, or the
  skills directory itself — and which of those it reads is the implementing pull request's to
  settle.
- **Duplication is bounded, and only the skills duplicate.** Two copies at most, about 375 KB each,
  and a sixth or seventh harness adds nothing unless it reads a directory neither existing one does.
  Everything else is written once no matter how many harnesses a project uses.
- **Symlinks were available and were still not used.** The release targets are macOS and Linux
  (`.goreleaser.yml` builds `darwin` and `linux` only), so no platform ruled links out. What ruled
  them out is that a link works only while every harness follows it, including harnesses codefall
  does not support yet, and a harness that does not follow one finds no skills at all — which reads
  to its user as codefall not being installed.
- **A shared file has one path in prose, and it is longer.** `../../../.codefall/shared/<file>`
  climbs three levels where the old path climbed two, and it assumes a skill is installed at
  `<skills dir>/skills/<verb>/`. That is the assumption the old path already made, one level
  shallower; `extensions/scripts/skill-health.sh` is what checks that every reference still
  resolves.

## Related

- `docs/decision-log.md`, *The install layout, 2026-09-17* — the alternatives seen and not taken,
  and what was found about harness support for links on the way.
- `docs/PLAN.md` — the landing order.
- Repository `AGENTS.md`, *Workflow* — the definition of a breaking change this record relies on.
- `cli/internal/shared/harness/harness.go` — where the two skills directories are defined, and where
  `SharedHooksDir` stops describing reality.
- `extensions/skills/AGENTS.md` — the path rule a skill follows for a shared file.
