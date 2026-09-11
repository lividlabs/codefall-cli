codefall
--------

**Codefall — Opinionated skills for the software development lifecycle, and the CLI that installs them.**

The repo holds two components:

- [`cli/`](cli/) — the Go installer (`codefall init`, `codefall doctor`). Rules: [cli/AGENTS.md](cli/AGENTS.md)
- [`extensions/`](extensions/) — the installable plugin: skills, hooks, and shared scripts. Rules:
  [extensions/AGENTS.md](extensions/AGENTS.md); the skills table lives in
  [extensions/README.md](extensions/README.md)

Every skill is a **verb**. The verbs chain: `conceptualize` frames the idea, `scaffold` makes the
project, `specify` states the problem, `mock-up` shows what it looks like, `design` decides the
shape, `implement` writes it, `review` checks it.

### Concepts

`conceptualize` writes the *why* down first — the problem, who feels it, the rough shape of an
answer, and the open questions — as `docs/concepts/CONCEPT-001-slug.md`. It is deliberately
informal, and its length is proportional to what you put in: two paragraphs is a valid concept, and
so is a page. Unknowns stay in the document as unknowns rather than being invented away.

**It takes whatever you arrive with.** A sentence, ten minutes of thinking out loud, a vision
document, a whiteboard photo, or a folder of design-tool exports. Mockups are routed to
`docs/mockups/` where `design` and `implement` look for them; everything else is saved verbatim
under `docs/concepts/sources/`, and the concept cites both.

A big pile usually holds more than one problem. Each problem that stands on its own becomes its own
concept, written as siblings rather than a parent and children. The breakdown is proposed and you
decide the cut.

**`scaffold` requires a concept**, and offers to run this skill when there isn't one — architecture
questions answered from a one-sentence description is where scaffolds go wrong. You can decline, and
the scaffold records that it ran without one.

A concept is `Draft` while you are adding to it, `Ready` once agreed, `Active` once work starts.
Replacing part of one adds a `Revised by` line; replacing it whole archives it to
`docs/concepts/archive/`.

### Specifications

`specify` writes the *what* as `docs/specs/SPEC-003-slug.md` — one cohesive feature, cut into
**requirements** with user stories and [EARS](https://alistairmavin.com/ears/) acceptance criteria.
Identifiers nest (`SPEC-003-REQ-01-AC-01`), numbering at every level is append-only, and the
document is canonical while the tracker (GitHub Issues today) mirrors it.

### Mockups

`mock-up` puts the picture in the repository at `docs/mockups/<slug>/` — imported from a design
tool untouched, or made here inline after a design-system survey. The slug names the **surface**,
not the spec, because one screen outlives several specs. When `specify` records a requirement whose
surface has no picture, its tracker issue is labelled `requires-mockup` and `design` refuses until
the mockup lands.

### Designs

`design` decides the *how* and puts the work into the graph. Two outputs: a document at
`docs/designs/DESIGN-007-slug.md` when the change earns one (contained fixes get beads only), and
Beads tasks with dependency edges. The task plan is staged as a reviewable table before it becomes
Beads; revising a design reconciles the graph. Hard-to-reverse choices become ADRs.

### Implementation

`implement` executes the graph: point it at a bead, an epic, or a design. One go gate starts the
run; after that only a failure stops it. Landing strategy is read from the graph's shape — parallel
chains, stacked topological merge, an epic branch only when fan-in needs it. A bead closes at done
(criteria verified, checks green, PR open), **never** at merge — a human performs every merge to
`main`, and the plugin ships a hook that mechanically denies the alternative.

### The stance

**Pure Clean Architecture organized package-by-component**, boundaries mechanically enforced,
components cheap to extract. The output is a **monolith on purpose** — one deployable, no network
between use cases — that stays cheap to split later ([ADR-BASE-01 through
ADR-BASE-03](extensions/skills/scaffold/templates/adrs/)). Profiles are scoped per **surface**
(`typescript-react`, `go` supported today), numbering per profile so they never collide.

### Skills detail

The full skills table, the status lifecycle of each document, and the hook inventory live in
[extensions/README.md](extensions/README.md); the verb-level reasoning in each
[`extensions/skills/<verb>/SKILL.md`](extensions/skills/).

## Layout

```
cli/            # the Go installer; make check is the gate
extensions/     # the installable plugin — skills/, hooks/, shared/
```

Skill detail: `extensions/skills/<verb>/SKILL.md`, `disable-model-invocation: true` on every one —
running a skill is a deliberate act. The extension ships hooks per harness
(definitions under `extensions/hooks/`): a `PreToolUse` guard that denies merges and pushes to the
default branch, and a `SessionStart` prime on what Beads knows where the harness has the event.

## Install

```
codefall init
```

`codefall init` detects (or takes `--harness`) and installs for the harness: Claude Code, Codex,
Muse, opencode, Antigravity. Harnesses on the `.agents/skills` convention get the extensions tree
mirrored under their project; Claude Code gets it through its own mechanism. `codefall doctor`
verifies the settings afterward.

## Roadmap

See [extensions/docs/ROADMAP.md](extensions/docs/ROADMAP.md).

## License

[MIT](LICENSE) — Copyright (c) 2026 Livid Labs, LLC, authored by Dave Jensen.

The templates under `extensions/skills/scaffold/templates/`, and everything `scaffold` copies from
them into your project, are additionally available under [0BSD](LICENSE): no attribution, no notice,
no obligation. Your architecture documents are yours.
