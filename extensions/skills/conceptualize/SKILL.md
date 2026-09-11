---
name: conceptualize
description: Get an idea onto paper before anyone specifies or scaffolds it — take whatever the user arrived with, from a sentence to a folder of mockups, organize it into one or more numbered concept documents under docs/concepts/, and record what is still unknown rather than inventing answers.
argument-hint: "[the idea, or a path to a document you already have]"
disable-model-invocation: true
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Write
  - Edit
  - Bash
  - WebSearch
  - WebFetch
---

# Conceptualize

Write down what someone wants to build, and why, before anyone decides what it does or how it is
built.

The output is a **concept document** at `docs/concepts/CONCEPT-NNN-slug.md`. It is deliberately
informal. A concept carries the *why* — the problem, the reason it matters now, the rough shape of an
answer — and it stops well short of the detail a specification needs.

Conceptualizing is not specifying. The moment the document starts saying what a user will observe
when the thing works, it has become `specify`'s job. See [The specify off-ramp](#the-specify-off-ramp).

Plugin paths in this document — the ones that start with `../` — are relative to this skill's
directory, the one holding this `SKILL.md`. Resolve them from where the file lives; they are not
relative to the user's project.

## Scope — why, not what

| In scope | Out of scope |
| --- | --- |
| The problem, and who feels it | What a user will observe when it works |
| Why it matters now rather than later | Acceptance criteria of any kind |
| The rough shape of an answer | Screens, fields, endpoints, entities, schemas |
| What is deliberately not covered | Work breakdown, sequencing, estimates |
| Constraints and systems this touches | Which layer or component the work lands in |
| What nobody has decided yet | Anything a reader could build from |

The tell that this skill has failed is a document someone could implement. A concept gives a reader
enough to reason about the idea and to decide whether to specify it. Nothing more.

### How big a concept is

There is no ceiling. A whole product, a product line, a capability area, a subsystem, or a rework of
how something already works are all concepts.

The floor is about **size, not category**. A fix can absolutely be a concept: "auditing in the
accounting system is broadly broken" is a concept, because the problem is large enough that someone
has to think before anyone specifies.

One bug, one screen, one endpoint, one field is usually smaller than this document is worth, and
`specify` or the work itself is the better destination. **That is advice, not a gate.** Say what you
think and why — "this looks like one field, which `specify` handles in a couple of questions" — and
then write the concept if that is what they want. The user knows things about the work that you do
not, and a concept nobody needed costs a file.

### When the pile covers more than one problem

A large input often holds several problems rather than one. A vision document can cover three, and a
folder of mockups spanning six surfaces usually does.

**Each problem that stands on its own is a concept.** They are written as siblings — `CONCEPT-003`,
`CONCEPT-004`, `CONCEPT-005` — never as a parent and children. What connects them is the source they
came from: each one's `Related` line points at the same saved document or mockup directory, so a
reader who finds one finds the rest.

**Do the breakdown with the user, not for them.** Say what you found and where the seams look like
they are, then let them decide the cut:

> This reads like three problems to me — the audit trail, the export format, and who is allowed to
> see any of it. I would write three concepts. Does that match how you think about it?

They may say it is one, and one is then the answer. Splitting is a reading of their material and you
can be wrong about it.

**Each concept is confirmed before it is written**, the same as a single one. Several documents in a
run is several confirmations, not one approval covering the set.

## The specify off-ramp

When what the user describes is really a specification, say so once, say why, and offer to switch.
These are the tells:

| Tell | Why it means `specify` |
| --- | --- |
| One capability, described end to end | A concept frames a problem; this already answers it |
| The *why* is inseparable from the *what* | A concept exists to carry a *why* that outlives any one feature |
| The problem is already written down and agreed | A second document restates it and adds no information |

**Every one of these is about the size of the thing, or about the *why* already existing. None of
them is about how concretely the user can describe the *what*.** That distinction is the whole point
of this section. Someone can name every screen, state exactly what a user will see, and hand you
finished mockups, and still have written nothing down about the problem any of it solves — and that
person needs a concept more than most, because the reasoning exists only in their head.

Say it plainly and give them the choice:

> This is one capability described end to end, and the reason for it is already written down in
> CONCEPT-002. A concept here would restate what you have. Want to run `/specify` instead?

**If they agree, stop this skill.** Do not write a concept as well. If they disagree, write the
concept — the rule below about pushing back once and then deferring applies here too.

## What a concept looks like

Two sections are mandatory: **Problem** and **Proposed shape**. Everything else appears only when the
user actually said something about it.

**The document's length is proportional to what the user put in.** This is the rule that matters most
in this skill. Your job is to get their words onto paper in an order that reads well — not to fill in
what they did not say. A heading with invented content under it is worse than no heading, because a
later reader cannot tell which parts came from the user.

**Unknowns are correct output.** A concept with four open questions in it has done its job. Record
what is unresolved under **Open questions**, phrased as the question, and move on. Resolving
everything is not the goal and pretending to have resolved something is a defect.

### The template

```markdown
# CONCEPT-NNN: {Title}

**Status:** Ready — {date}
**Related:** specs: _(identifiers once `/specify` creates them)_ · plan: _(link once `/design` creates it)_

<!--
Problem and Proposed shape are the minimum — every concept needs
both what's wrong and some sense of what you're pointing at. Every
other section is optional: include it if it's true and useful for
this piece of work, cut it if it isn't. A concept for one broken
capability might be just those two short paragraphs. Something that
reshapes a whole product might use all of them.
-->

## Problem

What's broken, missing, or costing us something right now — and who
feels it. State it plainly. No discovery narrative, no framing for
an audience that needs convincing — just what you know to be true.

## Proposed shape

The rough shape of a solution — enough to reason about, not enough
to build from. This is not the spec. If you catch yourself writing
API signatures or data models here, stop — that's `/specify`'s job.
Even a single sentence is fine if that's genuinely all there is yet.

## Why now

*(Optional — include when it's not obvious why this matters now
rather than later.)*
Why this, and why not later. What gets worse, or what do we lose, if
this sits untouched.

## Non-goals

*(Optional — include only when the scope could plausibly be
misread or is tempting to creep.)*
What this concept deliberately does not cover. This is what keeps
`/specify` from scope-creeping into adjacent problems this concept
wasn't trying to solve.

## Environment & constraints

*(Optional — include when this touches existing systems, has
integration points, or `/scaffold` needs to know something before
setting up the workspace. Skip for self-contained work.)*
Existing systems, integration points, and technical realities that
`/design` and `/scaffold` need before touching anything — what this
work depends on, what it touches, what it must not break.

## Open questions

*(Optional — include only if something real is actually unresolved.)*
Anything unresolved that `/specify` or `/design` will need to settle
before they can commit to an approach.

## Alternatives considered

*(Optional — only worth including when the choice itself was hard.)*
Other shapes considered, and why they were set aside.
```

**Every word of instruction is stripped on emit** — the HTML comment, the guidance paragraph under
each heading, and the `(Optional — …)` lines. The written concept carries the user's content and
nothing else. An optional section with nothing to say is deleted, heading included.

**The `Related` line also carries backward links** when they exist: the saved source under
`docs/concepts/sources/`, any mockup directory under `docs/mockups/` that came in with it, and any
concept this one revises or replaces.

**The `specs` field holds a list**, because one concept can produce several specifications —
`SPEC-003, SPEC-004, SPEC-005`. A concept large enough to need more than one spec is normal, and the
concept is what groups them: `specify` writes siblings rather than a parent spec, precisely because
this line already does that job. `specify` appends each identifier as it creates one, and changes
nothing else in the file.

## Status and lifecycle

`Status` is one word plus a date. Two optional header fields carry the pointers, because a concept
can be under active work and partly revised at the same time, and a single status line cannot say
both.

| Header | Meaning | Set by | Editable in place | Lives in |
| --- | --- | --- | --- | --- |
| `Status: Draft — <date>` | The user said, or clearly implied, that they are stopping and coming back | this skill, only on that signal | yes | `docs/concepts/` |
| `Status: Ready — <date>` | Written and agreed. The normal end of a session | this skill | yes | `docs/concepts/` |
| `Status: Active — <date>` | Work has started against it | `implement` | no, except the Status line | `docs/concepts/` |
| `Status: Archived — <date>` | Wholly replaced, or dropped | this skill | no | `docs/concepts/archive/` |
| `Revised by: CONCEPT-NNN — <date>` | Part of it was replaced; this concept is still live. Optional, repeatable | this skill | — | — |
| `Replaced by: CONCEPT-NNN — <date>` | Accompanies `Archived` when something took its place. Omitted when the concept was simply dropped | this skill | — | — |

**`Ready` is the normal end of a session, not `Draft`.** A concept the user worked through and agreed
with is finished, whatever they called it when they started. Set `Draft` only when they said they are
stopping and will come back — "let's leave it there for now", "I want to think about this more". They
will rarely use the word.

**`Active` is set by `implement`**, automatically, at the first claim of work that traces back to
the concept — it records the observable fact that work has started. Do not set it yourself; a
concept still at `Ready` means `implement` has not run against it.

**A concept is never renumbered and its identifier is never reused.** A spec citing `CONCEPT-003`
must always mean the same document, including after it is archived — the file moves, the identifier
does not change.

### The archive

Archiving moves the file to `docs/concepts/archive/` under the same name. The identifier stays valid
and citations still resolve, to the new path.

`conceptualize` maintains `docs/concepts/AGENTS.md`. It is written when the directory is created and
added on a later run if it is missing. That is the whole retrofit story — every repo gets the file the
first time the skill runs, and a repo that has never run this skill has no `docs/concepts/` for it to
scope. It says one thing, that `archive/` is history and is not read unless asked, which is the closest
thing to an ignore file that actually works: Claude Code has no `.agentignore`.

**Never overwrite a file that has drifted.** When one exists and its content differs from what this
skill ships, show the user the difference and ask. The difference is a hand edit until they say
otherwise, and a repair that silently discards someone's rule is worse than a stale file. Replace it
only on a yes; on a no, leave it and say nothing further about it.

The file's content:

```markdown
# docs/concepts — operative rules

- `archive/` is history. Do not read it unless the user asks about a superseded concept by name.
- `sources/` holds the material concepts were written from, verbatim. Never edit one.
- Mockups that arrived with a concept live in `docs/mockups/`, cited from its `Related` line.
- A concept's identifier is permanent. `CONCEPT-003` means the same document after it is archived.
- Status transitions are `/conceptualize`'s to make, never a hand edit.
```

## What the user already has

A concept may arrive as more than a conversation: notes, a transcript, a vision document, a whiteboard
photo, sketches, screenshots, a folder of design-tool exports, a link to a design tool, or several of
those at once. **All of it is fine, and none of it is discarded.**

Having a lot of material says nothing about whether this is a concept. Someone can arrive with
finished screens and have written nothing down about the problem they solve, and that is the case this
skill is most useful for.

### Where it goes

Two destinations, decided by what the material is rather than by how it arrived.

| Material | Where |
| --- | --- |
| A picture of a surface someone will build — exports, screenshots of a design tool, a mockup set | `docs/mockups/<slug>/`, following `../../shared/import-mockup.md` |
| Everything else — notes, transcripts, vision documents, sketches of the problem, a diagram of how something works today | `docs/concepts/sources/CONCEPT-NNN-<original-filename>` |

Mockups go to `docs/mockups/` because that is where `design` and `implement` look for them, and
because a surface outlives the concept that prompted it. `../../shared/import-mockup.md` is the same
procedure `specify` and `mock-up` use, so the three never drift — follow it as written, including its
rules about never editing or interpreting what the user brought.

**When it is unclear which a thing is, ask.** A whiteboard photo of boxes and arrows is thinking; a
photo of a drawn screen is a mockup; plenty of images are genuinely either. One question settles it.
A single run can use both destinations, and often does.

**The source is always saved**, without exception — including when a document fits so well that it is
only being renamed. It keeps its original name and extension, and the concept's `Related` line points
at wherever it landed. The source records what the user actually brought; the concept records what we
made of it, and someone may need to check one against the other.

**Never edit the user's original in place, and never delete it.**

### Documents that could be the concept

When what they brought is a written document, one of two things happens:

- **The document is a strict superset of the template** — it has Problem and Proposed shape and
  everything else it needs, plus whatever more its author wanted. Adopt it: copy it to
  `docs/concepts/CONCEPT-NNN-slug.md` with its body unchanged, add the header, and leave the
  structure alone. Reformatting a coherent document into our headings destroys work and adds nothing.
- **Anything less** — and this is the common case, including a document that is nearly right — write a
  new concept from it. Take the user's sentences over your own wherever they work, and put what the
  source does not answer under **Open questions**.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

### 1. Set up and look around

No precondition check. This skill needs a place to write and nothing else.

Create `docs/concepts/` in the working directory if it does not exist, along with `docs/concepts/AGENTS.md`.
**Do not ask where the documents go.** The working directory is the project; that is not a question.

Read what is already there — the concept documents, not the archive — and pick the next identifier as
the highest existing number plus one, zero-padded to three digits. If an existing concept covers what
the user is describing, say so and link it before going further. They may want to revise that one
rather than write another.

### 2. Take the idea

One open question:

> "What's the idea? However it comes out is fine — a sentence, or ten minutes of thinking out loud.
> And if you already have anything — notes, a document, sketches, mockups — point me at it."

**Ask once, at the start, and then drop it.** Someone who has a vision document will say so. Chasing
material the user did not mention wastes the part of this process that produces original thinking.

**A stream of consciousness is the best input this skill gets.** Take all of it, ask nothing while it
is arriving, and organize it into a draft before asking anything. Questions asked mid-dump interrupt
the only part of this process that produces original material.

If they pointed at anything, read it and follow [What the user already has](#what-the-user-already-has).

### 3. Check the altitude

Before interviewing, judge whether this is a concept at all. Apply [the specify off-ramp](#the-specify-off-ramp)
and [the floor](#how-big-a-concept-is). Doing this now costs one exchange; doing it after the
interview wastes the whole session.

### 4. Interview

Only for what is genuinely missing. A user who arrived with a clear problem and a rough answer has
given you a concept already — write it.

Ask about, in this order, and skip anything already answered:

1. What is broken, missing, or costing something — and who feels it.
2. The rough shape of an answer, if they have one. "I don't know yet" is a complete answer and goes
   under Open questions.
3. Why now rather than later.
4. What this deliberately does not cover.
5. What it touches that already exists, and what it must not break.

**Two rounds is a cap on your insistence, not on the conversation.** Stop pressing an unanswered point
after the second try and put it under Open questions. If the user keeps going, keep going with them —
never cut short a user who is still thinking out loud.

**Do not bikeshed.** Naming, wording, and which of two rough shapes is better are not this document's
problems. When a question is interesting but does not change what gets written, do not ask it.

**"Like $COMPANY does it."** When the user references another product, offer once to look it up. On
yes, research it and summarize only what matters to the concept. Confirm the summary with the user
before it reaches the document. On no, move on without searching.

### 5. Draft and confirm

Write the full document and show it before anything is saved. Say which sections you left out and why —
"no Non-goals section, because nothing you said was at risk of being misread as in scope" — so the user
can catch an omission that was actually a gap.

Then set the status: `Ready`, unless they said they are stopping and coming back, which is `Draft`.

### 6. Write and report

Write the concept, the source document if there was one, and `docs/concepts/AGENTS.md` if it was
missing. Report the path, the identifier, the status, and every open question the document carries.

Do not commit. Do not create issues. Do not start a specification.

## Other modes

Invoking this skill on an existing concept does one of four things. Ask which if it is not obvious.

- **Promote** `Draft` to `Ready`, or **reopen** `Ready` to `Draft` when nothing cites it yet.
- **Edit** a `Draft` or `Ready` concept in place. Both are editable; that is what makes them the steady
  states.
- **Revise** an `Active` concept — which is frozen, because work is underway and a silent edit changes
  what people are building. The revision is a new concept, and the old one gains a `Revised by` line.
- **Archive** a concept: set `Status: Archived`, add `Replaced by` if something took its place, and
  move the file to `docs/concepts/archive/`.

Every one of these is a user's decision. Report the state and offer; never transition a concept on your
own initiative.

## Rules

- **Nothing is written without the user confirming the full document first.**
- **The concept document says why, never what will be observed.** That is about the document. Mockups
  and other material can arrive with it and be cited from it; they are inputs, not the concept.
- **Length is proportional to input.** Never fill a heading. An optional section with nothing behind it
  is deleted, heading and all.
- **Unknowns are recorded, not resolved.** Open questions are correct output.
- **Push back once, then defer** — on altitude, on scope, on size, on a shape you think is wrong. Say
  what you think and why, then do what they ask. The user knows the domain and you may be wrong.
- **Everything the user brought is saved and never edited**, whether it lands in `sources/` or in
  `docs/mockups/`.
- **Identifiers are append-only.** A retired number is never reused, and archiving moves a file without
  renumbering it.
- **`Active` is not yours to set.** `implement` owns that transition.
- **Do not draw, specify, design, or estimate.** Every one of those is another verb's work.
