codefall-plugin
---------------

**Codefall — Opinionated skills for the software development lifecycle.**

A plugin for Claude Code (and, later, Codex and friends). Every skill is a **verb**.

## Skills

| Skill | Does | Status |
| --- | --- | --- |
| [`conceptualize`](plugins/codefall/skills/conceptualize/SKILL.md) | Get an idea onto paper before anyone specifies or scaffolds it: a numbered concept document under `docs/concepts/` that carries the problem, the rough shape of an answer, and what nobody has decided yet. | in progress |
| [`scaffold`](plugins/codefall/skills/scaffold/SKILL.md) | Start a new project on the Clean + package-by-component stance: ratified ADRs, scoped `AGENTS.md`, optionally project files and boundary lint. | in progress |
| [`graft`](plugins/codefall/skills/graft/SKILL.md) | Bring a scaffolded project's docs up to date with the current templates: report what changed since its version, with per-file provenance, and apply only what the user takes. Also handles first-time adoption of the stance. | in progress |
| [`specify`](plugins/codefall/skills/specify/SKILL.md) | Turn a feature idea into a specification another session can implement: a spec document under `docs/specs/` holding requirements with EARS acceptance criteria, mirrored to the issue tracker. | in progress |
| [`mock-up`](plugins/codefall/skills/mock-up/SKILL.md) | Get the visual surface of a feature into the repository under `docs/mockups/`: import what a design tool exported, or make the mockup here, matching the app's own design system so it looks like it belongs. | in progress |
| [`design`](plugins/codefall/skills/design/SKILL.md) | Decide how a feature gets built and put the work into the graph: a design document under `docs/designs/` scaled to the size of the change, ADRs for the choices that are hard to reverse, and the tasks in Beads with their dependency edges. | in progress |
| [`implement`](plugins/codefall/skills/implement/SKILL.md) | Execute the graph: claim ready beads, build each in an isolated worker worktree with tests as part of done, verify against acceptance criteria, open PRs, and walk the waves until the frontier is empty. Never merges to `main`. | in progress |

### Concepts

`conceptualize` writes the *why* down first — the problem, who feels it, the rough shape of an answer,
and the open questions — as `docs/concepts/CONCEPT-001-slug.md`. It is deliberately informal, and its
length is proportional to what you put in: two paragraphs is a valid concept, and so is a page.
Unknowns stay in the document as unknowns rather than being invented away.

**It takes whatever you arrive with.** A sentence, ten minutes of thinking out loud, a vision
document, a whiteboard photo, or a folder of design-tool exports. Mockups are routed to
`docs/mockups/` where `design` and `implement` look for them; everything else is saved verbatim under
`docs/concepts/sources/`, and the concept cites both. Arriving with finished screens is not a reason
to be sent elsewhere — you can have every screen drawn and still have written nothing down about the
problem they solve, which is the case this skill is most useful for.

**A big pile usually holds more than one problem.** A vision document can cover three, and a mockup
set spanning six surfaces usually does. Each problem that stands on its own becomes its own concept,
written as siblings rather than a parent and children, connected by the source they all came from. The
breakdown is proposed and you decide the cut.

**`scaffold` requires a concept**, and offers to run this skill when there isn't one. That requirement
exists because scaffolding without any idea of what is being built is where scaffolds go wrong — the
architecture questions get answered by defaults picked from a one-sentence description. With a concept
in hand most of those questions are already answered, so the scaffold session is shorter *and* the
answers are better. You can decline, and the scaffold records that it ran without one.

`specify` may draw on a concept and never requires one, because a concept carries the *why* and a
specification carries the *what* — they are different documents, and plenty of features need only the
second.

A concept is `Draft` while you are still adding to it, `Ready` once it is written and agreed, and
`Active` once work starts against it. Replacing part of one adds a `Revised by` line; replacing it
whole archives it to `docs/concepts/archive/`, where the identifier stays valid and the citations still
resolve.

### Specifications

`specify` writes the *what* as `docs/specs/SPEC-003-slug.md`. A spec holds one cohesive feature, cut
into **requirements** — each with a user story and its own acceptance criteria — because a requirement
is the unit somebody picks up and builds.

Acceptance criteria are written in [EARS](https://alistairmavin.com/ears/), the Easy Approach to
Requirements Syntax, published at Rolls-Royce in 2009 and used here unchanged. It constrains a
requirement to six sentence shapes with the clauses always in the same order:

```
The system SHALL order exported segments by departure time
WHEN a traveler selects export, the system SHALL produce a file containing the itinerary
IF the trip is missing a departure date, THEN the system SHALL name the missing field
WHILE an export is in progress, the system SHALL show progress and allow cancellation
```

The point of the notation is that a criterion reads as an obligation rather than an observation, so
the criteria *are* the requirements and there is no second list to keep in step with them. Failure
behavior gets its own keyword, which is what makes the error paths visible as a group instead of
scattered among the happy ones.

Identifiers nest and share a prefix, so `grep SPEC-003` finds the document, its requirements, and
every test and ticket that cites them:

```
SPEC-003                        the spec
SPEC-003-REQ-01                 a requirement
SPEC-003-REQ-01-AC-01           a criterion
```

Numbering is append-only at every level. A retired number is never reused, so a test citing
`SPEC-003-REQ-01-AC-04` never silently comes to mean something else.

**The document is canonical, and the tracker is a mirror of it.** GitHub gets a parent issue for the
spec and a child issue per requirement, carrying that requirement's story and criteria in full so
nobody has to click through to work the ticket. Re-running `specify` regenerates those bodies. The
spec's own `Status` is `Draft`, `Ready`, or `Archived` and describes the document only — whether the
work is queued, underway, or done is the tracker's to say.

A feature too large for one cohesive spec becomes sibling specs rather than a parent and children.
The concept above them is what groups them, which is why a concept's `Related` line holds a list.

Skills are **explicitly invoked** — `/scaffold`, `/specify`, and so on. Each carries
`disable-model-invocation: true`, so none of them fire on their own; scaffolding a project or filing
an issue is a deliberate act, not something inferred from a passing remark.

### Mockups

`mock-up` puts the picture in the repository, at `docs/mockups/<slug>/`. It runs before or after
`specify` — a mockup can be what makes the requirements obvious, or it can be drawn once they are
settled.

Two ways in. When you already work in a design tool, it **imports** what you exported and never
touches the files again: an imported asset is the record of what someone decided, and redrawing it
loses that. When you don't, it **makes** one.

**A mockup, not a wireframe.** Before drawing anything it reads the repository for your design
system, your tokens, your existing screens, and the fonts actually in use, then inlines what it found
so the file stays self-contained. The default is as close to what would ship as the repository lets it
get — someone opening it should see your product with a new screen in it, not a grey diagram of one.
A box drawing is still available when there is nothing to match yet or structure is the only open
question; it just isn't the starting point.

Everything else is a default with the reason attached and a stated case for going the other way:
static HTML, one file per state, realistic content, one viewport. When the open question is how
something *behaves* — a picker, a multi-step flow, a filter that has to feel right — it builds the
thing working, with a pinned library if that is what makes it faithful. Working or not, it stays a
reference: it proves the interaction and the implementer rebuilds it in the app's stack.

**Where the answer is genuinely open, you get options.** Two or three versions of the screen, named
by what differs — `list-table.html` and `list-cards.html` — with a line each on what they are better
at and which one it would pick. The one you take keeps the plain name and the README records the
decision.

The states are where the value is — populated, empty, and the primary failure at minimum, because the
empty and error screens are the ones nobody describes in an interview and where features come back
from review.

It opens by asking whether this is for a spec, for a concept, or a fresh start, and lists what is
there so you can pick one. A concept is a wider frame that usually spans several surfaces, so it says
how big the run would be before starting rather than refusing it.

The slug names the **surface**, not the spec. One screen gets touched by several specs over its life
and outlives all of them, so a mockup filed under whichever spec arrived first makes the second one
either duplicate it or reach into another spec's directory. Specs reference mockups by path under
their **Design notes**.

When `specify` records a requirement whose surface has no picture yet, its tracker issue is labelled
`requires-mockup` and `design` refuses to act on it. Landing the mockup clears the label.

### Designs

`design` decides the *how* and puts the work into the graph. Two outputs, and the second is the one
that always exists: a design document at `docs/designs/DESIGN-007-slug.md`, and the tasks in Beads
with their dependency edges.

**Not every change earns a document.** A fix contained to one component, changing nothing public and
coming to one or two tasks, gets beads and nothing else — a design document for a null check is the
ceremony this avoids. Anything crossing a component boundary, or fanning out past roughly three
dependent tasks, gets one. The document then has three required sections — Overview, Architecture,
Task Plan — and seven more that appear only when their trigger fires. A conditional section with
nothing behind it is deleted, heading and all.

Research findings go inline, next to the decision they bear on. There is no sibling `research.md`,
no `data-model.md`, and no `contracts/` directory: a finding filed away from its decision is a note
nobody reads.

**The task plan is staged before it is real.** It starts as a table with local identifiers, so the
dependency edges can be reviewed while they are still cheap to change — a flat list of tasks does not
catch the thing review is for, which is a wrong ordering or a missing prerequisite:

```
| ID | Task                              | Depends on | Design ref |
|----|-----------------------------------|------------|------------|
| T1 | Add `StageContext` type + serde   | —          | Components |
| T2 | Wire context load into `/scaffold`| T1         | Architecture |
```

Once you approve it, those rows become beads and the table is replaced by the line that records what
became what — `Epic: bd-a2g · T1→bd-unz · T2→bd-s58`. Beads is authoritative from that moment, and a
duplicate task list left behind in a git-tracked file would drift from it. The mapping is what lets a
later run update the graph instead of duplicating it.

**Revising a design reconciles the graph rather than rebuilding it.** A bead nobody has touched is
edited, whatever changed. A bead someone has claimed, commented on, or closed is replaced only when
the work already done against the old wording would no longer count — a ticket should not change
under the person holding it. A task that leaves the design is reported to you, never closed on its
own, because someone may still be working it.

Choices that are hard to reverse — a new dependency, a schema other components will build on, a
rejected alternative that cost real analysis — become an ADR in the project's own `ADR-NNN` sequence.
Most designs need none. A ratified ADR is never rewritten: a revision lands as a new, superseding
one.

The design's `Status` is `Draft`, `Ready`, or `Archived` and describes the document only. Whether the
work is queued, underway, or done is Beads' to say, the same division `specify` makes with its
tracker.

### Implementation

`implement` executes what `design` put into the graph. Point it at a bead, an epic, or a design — or
at nothing, and it shows the ready work and asks. An epic means walking the whole graph: each close
unblocks the next tasks, waves of background workers build them in isolated worktrees, and the run
continues until the frontier is empty.

**One approval starts it.** The go gate shows the landing strategy with its reason, a branch diagram,
the waves and the models proposed per task, what will be claimed in Beads, and — when a concept sits
behind the work — that go flips it to `Active`. After go, only a failure stops the run, and a failed
worker gets one automatic retry at higher effort before anything reaches you. That absence of
mid-run gates is what makes an overnight run possible.

**How work lands is read from the graph's shape.** Independent chains stack toward `main` in
parallel when their predicted file scopes are disjoint; overlap or fan-in serializes them into one
topological stack; an epic branch appears only when fan-in meets a real need for parallelism, or
when increments must not land on `main`. There is no depth cap — a deep stack costs only a muddy
three-dot diff until it drains bottom-up.

**A bead closes at done — acceptance criteria verified, checks green, PR open — not at merge.** That
is Beads' own semantics, and it is what lets a stacked dependent start the moment its parent's
branch is pushed. The merge seam is carried by gates: every PR gates a "landed" bead inside the
epic, so the epic cannot close until you have merged everything, and the next session's
`bd gate check` turns your merges into bead state.

**A human performs every merge to `main`.** The run ends at open PRs and a reported bottom-up merge
order, and the plugin ships a hook that mechanically denies the alternative. Tests are part of done
— the ones the design planned and the ones the work turned out to need — while regression and
fresh-context retesting stay with the future `test` verb.

### The stance

**Pure Clean Architecture organized package-by-component**, with boundaries **mechanically
enforced** and components that stay **cheap to extract**. Dependencies point inward only; the
interfaces a use case needs live with the use case in `application/`, not in `domain/`; the top level
is capabilities, each behind a facade with the Clean layers nested inside; a composition root per app
binds implementations.

The output is a **monolith on purpose** — one deployable, no network between use cases. What it is
not is a monolith you are stuck with: each component owns its own data, no transaction spans two of
them, and what crosses a facade is a contract rather than an entity. Pulling a component out later is
a deployment change, not a redesign.

That core is stack-agnostic ([ADR-BASE-01 through ADR-BASE-03](plugins/codefall/skills/scaffold/templates/adrs/)) — Clean
Architecture, package-by-component, and keeping the resulting monolith cheap to split. Each surface adds
a profile supplying its own ADRs under its own prefix — `ADR-TS-01` and up for `typescript-react`,
which is Inversify, the TanStack Query / Zustand / `useState` split, and `eslint-plugin-boundaries`;
`ADR-GO-01` and up for `go`. Numbering restarts per profile, so two profiles never collide, and a
project's own ADRs are a separate sequence starting at `ADR-001`.

### Surfaces

Profiles are scoped to a **surface**, not to a kind of product — a project composes as many profiles
as it has surfaces. And a surface is defined by **where domain logic lives**, not by which languages
appear in the repo.

| Surface profile | Covers | Status |
| --- | --- | --- |
| `typescript-react` | TypeScript/Node backends; React frontends, web and Native — including Tauri, Electron, and RN apps whose native side is only wiring | **supported** |
| `go` | Go services, APIs, workers, daemons, and CLIs that hold domain logic | **supported** |
| `rust-native` | Rust services, and Tauri shells that hold domain logic | planned |
| `kotlin-native` · `swift-native` | Android, iOS | planned |
| `dart-flutter` | Flutter, mobile and desktop | planned |
| `python` · `java` | backends and services | planned |

So a **Tauri desktop app and a React Native mobile app are both scaffoldable today**, whole, when
their native side is boilerplate plus a few commands wrapping OS APIs. Those commands are gateway
implementations living in the `infrastructure/` ring — per ADR-BASE-01, a Tauri `invoke`, a bridge call,
an HTTP request, and a platform channel are all just gateways, and the use case never learns which
one it got. Rust or Kotlin in the repo doesn't make it a native surface, any more than a Postgres
driver makes SQL one.

It's two surfaces only when the native side holds real domain logic. Then it needs its own profile
plus a **seam ADR** deciding which side owns the domain.

React is the component model; the renderer is an outer-ring detail, so React DOM and React Native
share one profile. What differs is toolchain — and the profile records the trap: Metro transpiles
with Babel, not `tsc`, so ADR-TS-01's `emitDecoratorMetadata` is inert and Inversify fails at runtime
unless `babel-plugin-transform-typescript-metadata` is added.

`scaffold` reads your concept, or asks you to describe the project when you declined one, decomposes
it into surfaces, and matches each against this table. **If any surface has no profile, it stops** — it won't improvise ADRs for an unsupported
language or scaffold only the half that fits. A profile counts as supported once
`templates/surfaces/<name>/PROFILE.md` is complete.

## Roadmap

The verbs chain: `conceptualize` frames the idea, `scaffold` makes the project, `specify` states the
problem, `mock-up` shows what it looks like, `design` decides the shape, `implement` writes it,
`review` checks it.

| Verb | Does |
| --- | --- |
| `review` | Review specs or code. Eventually multi-harness. |
| `test` | Re-prove the system after work lands: regression passes, fresh-context retesting, coverage campaigns, agentic runs. |
| `migrate` | Restructure code to a changed stance — architecture moves, component extraction. The code-refactor counterpart to `graft`, which moves only documents. |

## Layout

This repo is a **marketplace** at the root and a **plugin** in `plugins/codefall/`. The two are
versioned independently on purpose: the catalog stays current on `main` while the entry pins the
plugin to a released commit.

```
.claude-plugin/
  marketplace.json    # the catalog — pins plugins/codefall to a released commit
plugins/
  codefall/           # the plugin; this subtree is what gets installed
    .claude-plugin/
      plugin.json     # plugin manifest
    hooks/
      hooks.json                  # PreToolUse wiring for the merge guard
      block-merge-to-main.sh      # denies merges and pushes to the default branch
    shared/
      customizations.md           # how every verb reads .codefall/skills/<verb>/CUSTOMIZE.md
      import-mockup.md            # one import procedure, used by specify and mock-up
      preflight.sh                # the beads precondition check; reports, never repairs
    skills/
      conceptualize/
        SKILL.md                    # the concept template and the status lifecycle
      design/
        SKILL.md                    # the tiers, the document, the staged task plan, the Beads graph
      graft/
        SKILL.md
        lineage.md                  # what every template used to be called; graft's rename record
      implement/
        SKILL.md                    # scope, landing strategies, the go gate, workers, the bead loop
        worker-prompt.md            # the strategy-blind prompt each background worker is rendered
      mock-up/
        SKILL.md                    # the two modes, the design survey, defaults, options, states
      scaffold/
        SKILL.md
        templates/
          adrs/                     # stack-agnostic: ADR-BASE-01..03, _TEMPLATE
          surfaces/
            typescript-react/
              PROFILE.md            # prefix, fit, visibility model, toolchain, depth notes
              adrs/                 # ADR-TS-01..03
              AGENTS.md.skeleton    # scoped per-surface rules doc
            go/
              PROFILE.md
              adrs/                 # ADR-GO-01..03
              AGENTS.md.skeleton
      specify/
        SKILL.md                    # the spec template, EARS criteria, the status lifecycle
        trackers/
          github/
            PROFILE.md            # capabilities, field mapping, issue generation and refresh
```

Within the plugin, `skills/`, `commands/`, `agents/`, and `hooks/hooks.json` are auto-discovered by
Claude Code — no manifest entries needed when adding new ones.

## Install

```
/plugin marketplace add lividlabs/codefall-plugin
/plugin install codefall@codefall
```

The catalog is read from `main`, so it is always current, but the `codefall` entry is pinned to the
commit of the most recent release. Merges to `main` do not reach installs; merging a release PR does.

### Local development

Point the marketplace at your checkout and install the `codefall-dev` entry, which resolves to
`plugins/codefall/` in your working tree rather than to a published commit:

```
/plugin marketplace add /path/to/codefall-plugin
/plugin install codefall-dev@codefall
```

Installing `codefall` from a local checkout would still fetch the pinned commit from GitHub, so use
`codefall-dev` when you want to see your edits.

## License

[MIT](LICENSE) — Copyright (c) 2026 Livid Labs, LLC, authored by Dave Jensen.

The templates under `plugins/codefall/skills/scaffold/templates/`, and everything `scaffold` copies from them into
your project, are additionally available under [0BSD](LICENSE): no attribution, no notice, no
obligation. Your architecture documents are yours.
