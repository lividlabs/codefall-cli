---
name: codefall-conceptualize
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
when the thing works, it has become `codefall-specify`'s job.

Paths that start with `reference/`, `templates/`, or `../` are relative to this skill's directory,
not the user's project.

## Files beside this one

Read each when its step says to; none is loaded up front.

- `templates/concepts/CONCEPT.md` — the document template. Read at step 5. Every word of
  instruction in it — the HTML comment, the guidance under each heading, the `(Optional — …)`
  lines — is stripped on emit.
- `templates/concepts/AGENTS.md` — the operative rules this skill installs at
  `docs/concepts/AGENTS.md`.
- `reference/inputs.md` — where the material a user brings goes, and when a document they brought
  is adopted as the concept. Read at step 2 when they point at anything.
- `../../shared/import-mockup.md` — the shared procedure for bringing a mockup into the
  repository. Read from `reference/inputs.md` when the material is a mockup.

## Scope — why, not what

| In scope | Out of scope |
| --- | --- |
| The problem, and who feels it | What a user will observe when it works |
| Why it matters now rather than later | Acceptance criteria of any kind |
| The rough shape of an answer | Screens, fields, endpoints, entities, schemas |
| What is deliberately not covered | Work breakdown, sequencing, estimates |
| Constraints and systems this touches | Which layer or component the work lands in |
| What nobody has decided yet | Anything a reader could build from |

The tell that this skill has failed is a document someone could implement.

### How big a concept is

There is no ceiling. A whole product, a product line, a capability area, a subsystem, or a rework of
how something already works are all concepts.

The floor is about **size, not category**. "Auditing in the accounting system is broadly broken" is
a concept. One bug, one screen, one endpoint, one field is usually smaller than this document is
worth, and `codefall-specify` or the work itself is the better destination. **That is advice, not a
gate.** Say what you think and why — "this looks like one field, which `codefall-specify` handles in
a couple of questions" — and then write the concept if that is what they want.

### When the pile covers more than one problem

**Each problem that stands on its own is a concept.** They are written as siblings —
`CONCEPT-003`, `CONCEPT-004`, `CONCEPT-005` — never as a parent and children. Each one's `Related`
line points at the same saved source or mockup directory.

**Do the breakdown with the user, not for them.** Say what you found and where the seams look like
they are, then let them decide the cut:

> This reads like three problems to me — the audit trail, the export format, and who is allowed to
> see any of it. I would write three concepts. Does that match how you think about it?

They may say it is one, and one is then the answer. **Each concept is confirmed before it is
written.** Several documents in a run is several confirmations.

## The specify off-ramp

When what the user describes is really a specification, say so once, say why, and offer to switch.

| Tell | Why it means `codefall-specify` |
| --- | --- |
| One capability, described end to end | A concept frames a problem; this already answers it |
| The *why* is inseparable from the *what* | A concept exists to carry a *why* that outlives any one feature |
| The problem is already written down and agreed | A second document restates it and adds no information |

**None of these is about how concretely the user can describe the *what*.** Finished mockups and
named screens with nothing written down about the problem they solve is a concept, not a spec.

Say it plainly and give them the choice:

> This is one capability described end to end, and the reason for it is already written down in
> CONCEPT-002. A concept here would restate what you have. Want to run `/specify` instead?

**If they agree, stop this skill.** If they disagree, write the concept — push back once, then
defer.

## What a concept looks like

Two sections are mandatory: **Problem** and **Proposed shape**. Everything else appears only when
the user actually said something about it. The template is `templates/concepts/CONCEPT.md`.

- **The document's length is proportional to what the user put in.** Get their words onto paper in
  an order that reads well; never fill in what they did not say. An optional section with nothing
  to say is deleted, heading included.
- **Unknowns are correct output.** Record what is unresolved under **Open questions**, phrased as
  the question, and move on.
- **The `Related` line carries backward links** when they exist: the saved source under
  `docs/concepts/sources/`, any mockup directory under `docs/mockups/`, and any concept this one
  revises or replaces.
- **The `specs` field holds a list** — `SPEC-003, SPEC-004, SPEC-005`. `codefall-specify` appends
  each identifier as it creates one, and changes nothing else in the file.

## Status and lifecycle

`Status` is one word plus a date. Two optional header fields carry the pointers.

| Header | Meaning | Set by | Editable in place | Lives in |
| --- | --- | --- | --- | --- |
| `Status: Draft — <date>` | The user said, or clearly implied, that they are stopping and coming back | this skill, only on that signal | yes | `docs/concepts/` |
| `Status: Ready — <date>` | Written and agreed. The normal end of a session | this skill | yes | `docs/concepts/` |
| `Status: Active — <date>` | Work has started against it | `codefall-implement` | no, except the Status line | `docs/concepts/` |
| `Status: Archived — <date>` | Wholly replaced, or dropped | this skill | no | `docs/concepts/archive/` |
| `Revised by: CONCEPT-NNN — <date>` | Part of it was replaced; this concept is still live. Optional, repeatable | this skill | — | — |
| `Replaced by: CONCEPT-NNN — <date>` | Accompanies `Archived` when something took its place. Omitted when the concept was simply dropped | this skill | — | — |

- **`Ready` is the normal end of a session, not `Draft`.** Set `Draft` only when they said they are
  stopping and will come back — "let's leave it there for now", "I want to think about this more".
  They will rarely use the word.
- **`Active` is set by `codefall-implement`**, at the first claim of work that traces back to the
  concept. Do not set it yourself.
- **A concept is never renumbered and its identifier is never reused.** Archiving moves the file to
  `docs/concepts/archive/` under the same name; citations still resolve.

## The concepts directory

`codefall-conceptualize` maintains `docs/concepts/AGENTS.md` from `templates/concepts/AGENTS.md`:
written when the directory is created, added on a later run if it is missing.

**Never overwrite a file that has drifted.** When one exists and differs from the template, show the
difference and ask. Replace it only on a yes; on a no, leave it and say nothing further about it.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

### 1. Set up and look around

No precondition check. Create `docs/concepts/` in the working directory if it does not exist, along
with `docs/concepts/AGENTS.md`. **Do not ask where the documents go.** The working directory is the
project.

Read what is already there — the concept documents, not the archive — and pick the next identifier
as the highest existing number plus one, zero-padded to three digits. If an existing concept covers
what the user is describing, say so and link it before going further; they may want to revise that
one rather than write another.

### 2. Take the idea

One open question:

> "What's the idea? However it comes out is fine — a sentence, or ten minutes of thinking out loud.
> And if you already have anything — notes, a document, sketches, mockups — point me at it."

**Ask once, at the start, and then drop it.** Do not chase material the user did not mention.

**A stream of consciousness is the best input this skill gets.** Take all of it, ask nothing while
it is arriving, and organize it into a draft before asking anything.

If they pointed at anything, read it and follow `reference/inputs.md`.

### 3. Check the altitude

Before interviewing, judge whether this is a concept at all. Apply
[the specify off-ramp](#the-specify-off-ramp) and [the floor](#how-big-a-concept-is). Doing this
now costs one exchange; doing it after the interview wastes the whole session.

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

**Two rounds is a cap on your insistence, not on the conversation.** Stop pressing an unanswered
point after the second try and put it under Open questions. If the user keeps going, keep going
with them.

**Do not bikeshed.** Naming, wording, and which of two rough shapes is better are not this
document's problems.

**"Like $COMPANY does it."** When the user references another product, offer once to look it up. On
yes, research it and summarize only what matters to the concept; confirm the summary with the user
before it reaches the document. On no, move on without searching.

### 5. Draft and confirm

Compose the full document from `templates/concepts/CONCEPT.md` and show it before anything is
saved. Say which sections you left out and why — "no Non-goals section, because nothing you said
was at risk of being misread as in scope" — so the user can catch an omission that was a gap.

Then set the status: `Ready`, unless they said they are stopping and coming back, which is `Draft`.

### 6. Write and report

Write the concept, the source document if there was one, and `docs/concepts/AGENTS.md` if it was
missing. Report the path, the identifier, the status, and every open question the document carries.

Do not commit. Do not create issues. Do not start a specification.

## Other modes

Invoking this skill on an existing concept does one of four things. Ask which if it is not obvious.

- **Promote** `Draft` to `Ready`, or **reopen** `Ready` to `Draft` when nothing cites it yet.
- **Edit** a `Draft` or `Ready` concept in place.
- **Revise** an `Active` concept — which is frozen, because work is underway. The revision is a new
  concept, and the old one gains a `Revised by` line.
- **Archive** a concept: set `Status: Archived`, add `Replaced by` if something took its place, and
  move the file to `docs/concepts/archive/`.

Every one of these is a user's decision. Report the state and offer; never transition a concept on
your own initiative.

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
- **`Active` is not yours to set.** `codefall-implement` owns that transition.
- **Do not draw, specify, design, or estimate.** Every one of those is another verb's work.
