---
name: specify
description: Turn a feature idea into a specification another session can implement — interview for what the user will observe, push back on vague answers, audit what already exists, then write requirements with EARS acceptance criteria into a spec document in the repository, mirrored to the issue tracker.
argument-hint: "[what you want to build]"
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

# Specify

Turn a feature idea into a specification precise enough that another session can implement it
without re-interviewing anyone.

The output is a **spec document in the repository** at `docs/specs/SPEC-NNN-slug.md`, holding one or
more requirements, each with a user story and numbered acceptance criteria. The issue tracker gets a
generated mirror so people can see what is ready, in progress, and done. **The document is
canonical**; the issues are regenerated from it.

Everything else this skill does — the interview, the push-back, the audit of what already exists —
is in service of making those criteria correct.

Specifying is not designing. Once the criteria are written and confirmed, stop. How the thing gets
built is `design`'s work, and `implement`'s after that.

## Scope — what, not how

Specify decides **what will be true when this is done**. It does not decide **how it gets built**.

| In scope | Out of scope |
| --- | --- |
| Who the consumer is and what they get out of it | Which layer, component, or module the work lands in |
| What they can observe when it works | File paths, libraries, services, frameworks |
| Domain nouns and what they mean | Their fields, types, relations, or storage |
| What must keep working that already works | API shapes, message contracts, schema |
| What is explicitly not included | Work breakdown, ticket sequencing, dependency edges |
| Which questions remain unresolved | Estimates, effort sizing, build order |

This is a hard line, and crossing it is the most common way this skill fails. The tell is a
specification that names a technology. If the criteria mention a table, a class, an endpoint, or a
package, they have stopped describing the outcome and started describing the solution.

**The third row is the one that needs watching.** Naming a domain noun is in scope, because a
specification cannot describe an itinerary without the word. Saying what the noun is made of is not.
See [Key entities](#key-entities).

**You may read the codebase, but only to answer two questions**: does this already exist, and what
would this change silently break? Reading code to decide *how* to build the thing is `design`'s job
and is out of bounds here, even when the answer seems obvious.

Plugin paths in this document — `trackers/…` and the ones that start with `../` — are relative to
this skill's directory, the one holding this `SKILL.md`. Resolve them from where the file lives;
they are not relative to the user's project.

## What a specification is

One **spec document** per feature. Inside it, one or more **requirements**, each with a user story
and its own acceptance criteria.

A requirement is the unit that becomes a ticket. That is why a spec holds several rather than one:
a feature worth specifying almost always has more than one thing a consumer can do with it, and each
of those is separately buildable and separately demonstrable.

### The user story

Each requirement opens with one:

```
As a <consumer>, I want <capability>, so that <benefit>.
```

**The "so that" clause is mandatory.** A story without it names an action with no reason, and a
criterion written against it has nothing to be judged against. If the user cannot supply the benefit,
that is a signal the requirement is not understood yet — ask, do not fill it in.

**The consumer is not always a person, and pretending otherwise makes the story worse.** A system can
be the consumer:

> As the BFF, I want to cache search results, so that a repeated search does not pay the upstream
> round trip again.

Reach for a human consumer first, because a human benefit is harder to fake. But do not manufacture
one. `As an engineer` is usually a punt — legitimate when building developer tooling, evasive
otherwise.

**When no real persona exists, say so in the story rather than inventing one:**

> A clean persona was hard to identify here; the closest consumer is the on-call engineer,
> because they are the only party who observes the outcome.

That sentence is more useful to the implementer than a plausible fiction.

**Some obligations have no story, and that is expected.** Audit logging, data retention, and
authorization rules are real requirements that nobody wants as a user. Write them as a requirement
with a system consumer, or as criteria under the requirement they constrain. Do not invent a
traveler who wants their security events logged.

### Acceptance criteria are written in EARS

EARS — the Easy Approach to Requirements Syntax — constrains a requirement to a small set of
sentence shapes with the clauses always in the same order. It was published by Alistair Mavin and
colleagues at Rolls-Royce in 2009 and is used unchanged here.

The generic form:

```
While <optional pre-condition>, when <optional trigger>,
the <system> shall <system response>
```

The ruleset: zero or many preconditions, zero or one trigger, one system, one or many responses.
That yields six patterns.

| Pattern | Shape | Use for |
| --- | --- | --- |
| Ubiquitous | `The system shall <response>` | facts that are always true |
| State driven | `While <precondition>, the system shall <response>` | obligations that hold throughout a state |
| Event driven | `When <trigger>, the system shall <response>` | responses to an expected event |
| Optional feature | `Where <feature is included>, the system shall <response>` | behavior present only in some configurations |
| Unwanted behaviour | `If <trigger>, then the system shall <response>` | responses to an undesired situation |
| Complex | `While <precondition>, when <trigger>, the system shall <response>` | combinations of the above |

Worked:

```
SPEC-003-REQ-01-AC-01
  WHEN a traveler selects export on a trip that has a destination and both dates,
  the system SHALL produce a file containing the destination, departure date,
  return date, and traveler count

SPEC-003-REQ-01-AC-02
  IF the trip is missing a departure or return date, THEN the system SHALL make
  export unavailable and SHALL name the missing field

SPEC-003-REQ-01-AC-03
  WHILE an export is in progress, the system SHALL show progress and SHALL allow
  the traveler to cancel it

SPEC-003-REQ-01-AC-04
  The system SHALL order exported segments by departure time
```

**Pick the pattern that fits, not the one that is familiar.** The most common error in EARS is
writing everything as `WHEN … THEN`, including facts that are always true. `WHEN a traveler views a
trip THEN the system SHALL display its tags` is a ubiquitous requirement wearing an event: the
display obligation does not depend on a trigger. Write `The system SHALL display a trip's tags` and
drop the ceremony. The ruleset allows zero triggers for exactly this reason.

**Failure behavior uses `IF … THEN`, not `WHEN`.** The distinction between an expected event and an
undesired situation is worth keeping, because it makes the failure paths visible as a group — which
is what the interview asks for and what an implementer most often finds missing.

**Say "the system" or name the actual system.** Either is fine and consistency within a spec matters
more than which. What is never fine is dropping the subject, because a criterion without a subject
states an observation rather than an obligation and stops being a requirement.

**EARS governs criteria and nothing else.** The story, Context, Key entities, Edge cases,
Assumptions, Out of scope, and Open questions are plain prose. Do not write `SHALL` in them.

#### Criteria must be observable in a running system

Not observable in a test — observable in the product. This matters most for technical work, and it
has a consequence you must surface rather than absorb quietly.

Take the BFF cache above. "A repeated search does not call upstream" is only observable if something
emits that fact — a metric, a log line, a response header, an admin view. **That instrumentation is
now part of the requirement**, and it was arrived at by writing the criterion.

Say so out loud when it happens:

> Making this observable means the service has to expose cache hits somewhere. That is a real
> addition to the work. In scope, or do you want to leave it unverifiable for now?

Never quietly inflate the requirement, and never write a criterion nobody can check.

#### Implementation vocabulary is prohibited

`WHEN I POST to /api/trips` passes a syntax check and defeats the entire purpose. So does naming a
table, a component, a queue, or a library. A criterion describes what someone observes, in words they
would use.

The test: could this criterion still be true after a complete rewrite of the implementation? If not,
it is describing the solution.

#### Numbering

Three identifiers, nested, all hyphenated to match `CONCEPT-002` and `ADR-BASE-01`:

```
SPEC-003                        the spec
SPEC-003-REQ-01                 a requirement within it
SPEC-003-REQ-01-AC-01           a criterion within that requirement
```

Spec numbers are three digits, requirement and criterion numbers two, all zero-padded.

**Criterion numbers scope to their requirement**, so `AC-01` restarts under each one.
`SPEC-003-REQ-01-AC-01` and `SPEC-003-REQ-02-AC-01` are both valid and are different criteria.

**Write identifiers in full inside the document.** A short local form saves characters and costs the
reader the ability to copy a criterion reference straight into a test name or a ticket.

**Numbering is append-only, at every level.** A deleted criterion retires its number within its
requirement; a deleted requirement retires its number within its spec; a spec identifier is never
reused. A test citing `SPEC-003-REQ-01-AC-07` must never silently come to mean a different criterion.
This is the same discipline this repo applies to ADR history and concept identifiers.

Because everything shares one prefix, `grep SPEC-003` finds the document, its requirements, and every
citation of its criteria in one pass. Protect that: do not invent an alternate short form.

#### Criteria exist for coverage, never for symmetry

**Never write a criterion to fill a section.** If a requirement's behavior is fully covered by three
criteria, write three. Requirements within one spec will have different numbers of criteria, and a
document that looks lopsided is not a defect.

The only question is whether the behavior is covered. Balance across requirements is not a goal.

### Edge cases

Optional, per requirement. Boundary conditions that were considered and **deliberately left
unhandled**.

The section has a narrow job, because most edge cases do not belong in it. An edge case with a
decided answer is an `IF … THEN` criterion and belongs above with the other criteria. An edge case
nobody has an answer for is an **Open question**. What stays here is the third kind: something the
user considered, decided not to specify behavior for, and wants on the record.

> - A trip edited while its export is running — the exported file may be stale; not addressed in
>   this spec

Recording that is worth more than it looks. Without it, the next reader cannot tell a gap that was
weighed from one nobody noticed.

Do not use this section as a worksheet. Edge cases raised during the interview get resolved into
criteria, open questions, or an out-of-scope line, and only what survives that gets written.

### Key entities

Optional, per spec. The domain nouns the specification uses, and what each one means.

```
- **Trip** — a planned journey a traveler has saved
- **Itinerary** — the traveler-readable rendering of a trip
```

**Names and meanings only.** No fields, no types, no relations, no identifiers, no storage. `Trip`
has a meaning; `Trip has a departureDate: Date and belongs to a User` is a data model, and a data
model is `design`'s output, not this skill's input.

Include the section when the feature involves data and the vocabulary needs settling. Omit it when
the nouns are ordinary English that nobody could misread.

### Assumptions

Optional, per spec. Calls made without asking, written down so the user can overrule them.

This is a different thing from the two sections next to it, and keeping them separate is what makes
each useful:

| Section | Means |
| --- | --- |
| **Assumptions** | I decided this without asking. Correct me if it is wrong. |
| **Out of scope** | We decided together that this is not included. |
| **Open questions** | Nobody knows yet, and someone has to answer before this is built. |

An assumption is legitimate because the interview cannot ask forty questions. It is not a licence to
skip push-back: anything the interview should have surfaced belongs in the interview. Assumptions are
for the defensible defaults underneath a question nobody would think to ask.

> - Recipients read the itinerary outside the app; no recipient account is required
> - The existing share sheet is reused rather than replaced

## The template

```markdown
# SPEC-NNN: {Title}

**Status:** Ready — {date}
**Concept:** CONCEPT-002-trip-sharing
**Issue:** #142

[The Concept row is omitted when no concept framed this work. The Issue row is filled in once
the tracker mirror exists.]

## Context

[Two or three paragraphs in plain language: why this matters, what the consumer is doing when
they need it, how it fits what already exists. No technology. When a concept framed this work,
the header already points at it — do not restate what it says.]

## Key Entities

[Optional. Domain nouns and what they mean. No fields, no types, no relations.]

- **<Noun>** — <what it means>

## SPEC-NNN-REQ-01: {Requirement title}

As a <consumer>, I want <capability>, so that <benefit>.

### Acceptance criteria

- **SPEC-NNN-REQ-01-AC-01** — <EARS criterion>
- **SPEC-NNN-REQ-01-AC-02** — <EARS criterion>

### Edge cases

[Optional. Boundary conditions considered and deliberately left unhandled. Anything with a
decided answer belongs above as a criterion.]

- <boundary condition> — <what happens, and that it is not addressed here>

## SPEC-NNN-REQ-02: {Requirement title}

As a <consumer>, I want <capability>, so that <benefit>.

### Acceptance criteria

- **SPEC-NNN-REQ-02-AC-01** — <EARS criterion>

## Existing behavior preserved

[Only when the audit found something. Otherwise omit the section entirely.]

- <element> — <what it does today>

## Assumptions

[Optional. Calls made without asking, stated so they can be overruled.]

- <assumption>

## Out of scope

- <explicitly not included>

## Design notes

[Visual direction, mockup path, distilled research the user confirmed. Omit if empty.]

## Open questions

[Only when push-back did not resolve something. Otherwise omit.]

- <unresolved question>
```

**Omit empty sections.** A document full of headings with nothing under them trains readers to skim.
That includes the header rows: a spec with no concept has no `**Concept:**` line, not an empty one.

**Every word of instruction is stripped on emit.** The bracketed guidance under each heading does not
reach the written spec.

## Status and lifecycle

Three states, one word plus a date.

| Status | Meaning | Lives in |
| --- | --- | --- |
| `Draft — <date>` | Being written. The user stopped and is coming back | `docs/specs/` |
| `Ready — <date>` | Written and agreed. The normal end of a session | `docs/specs/` |
| `Archived — <date>` | Superseded or dropped | `docs/specs/archive/` |

**The spec's status describes the document, never the work.** Whether the work is queued, underway,
or done is the tracker's to say, and it says it better than a status line can. This is the one place
`specify` deliberately diverges from `conceptualize`, which carries an `Active` state precisely
because a concept has no tracker representation to carry it.

**`Ready` is the normal end of a session, not `Draft`.** A spec the user worked through and agreed
with is finished. Set `Draft` only when they said they are stopping and will come back.

**A spec is never renumbered and its identifier is never reused.** A test citing
`SPEC-003-REQ-01-AC-04` must always mean the same criterion, including after the spec is archived —
the file moves to `docs/specs/archive/` under the same name, and the citation still resolves.

`Archived` gains a `**Replaced by:** SPEC-NNN — <date>` line when something took its place, and is
omitted when the spec was simply dropped.

Status transitions are this skill's to make. Report the state and offer; never transition a spec on
your own initiative.

## The specs directory

`specify` maintains `docs/specs/AGENTS.md`. It is written when the directory is created and added on
a later run if it is missing. A repo that has never run this skill has no `docs/specs/` to scope, so
that is the whole retrofit story.

**Never overwrite a file that has drifted.** When one exists and its content differs from what this
skill ships, show the user the difference and ask. The difference is a hand edit until they say
otherwise, and a repair that silently discards someone's rule is worse than a stale file. Replace it
only on a yes; on a no, leave it and say nothing further about it.

The file's content:

```markdown
# docs/specs — operative rules

- `archive/` is history. Do not read it unless the user asks about a superseded spec by name.
- Spec, requirement, and criterion identifiers are permanent and append-only. A retired number is
  never reused, and `SPEC-003` means the same document after it is archived.
- The spec document is canonical. Tracker issues are generated from it and are regenerated, not
  hand-edited.
- Mockups are keyed by surface under `docs/mockups/<slug>/`, never filed under a spec. One mockup
  can serve several specs.
- Status transitions are `/specify`'s to make, never a hand edit.
```

## Mockups are keyed by surface, not by spec

Mockups live at `docs/mockups/<slug>/`, where the slug names the surface — `booking-history`,
`trip-share`. They are **never** filed under a spec or a concept.

The reason is that a mockup is a view of a surface, not of a specification. One screen gets touched
by several specs over its life, sometimes across several concepts, and it outlives any one of them.
Filing it under the spec that happened to arrive first makes the second spec either duplicate it or
reach into another spec's directory.

Specs reference mockups by path, under **Design notes**. Do not move existing mockups into a spec
directory, and do not create one.

## Cohesion and splitting

**A spec holds one cohesive feature.** Its requirements share a consumer and a purpose, and someone
reading them together should see one thing rather than a list.

The check is whether the requirements belong to each other. Requirements that share nothing but the
session they were written in are two specs.

**Push back once when a spec looks incohesive**, and carry a specific alternative — not "this is
large" but "requirements one through three are about exporting and four and five are about sharing
permissions; those look like two specs to me."

**Then defer.** If the user disagrees, write what they asked for. They know the domain, and what
looks like two features from outside is often one from inside. You may simply be wrong.

**When a split happens, the results are siblings, not a parent and children.** `SPEC-003`,
`SPEC-004`, and `SPEC-005` sit alongside each other, and the concept above them is what groups them:
a concept's `Related` line holds every spec that came out of it. Do not invent a parent spec — the
grouping already exists, and a second one has to be kept current.

When a sibling is large enough to deserve its own interview, say so and suggest a separate `specify`
session for it rather than writing a thin document now.

Record the concern in the spec **only** when the user did not engage with it — glossed over it, or
moved on without answering. Phrase it as an observation for `design` to weigh, not a warning. If they
considered it and disagreed, the matter is closed and nothing goes in the document.

## Project customizations

Follow `../../shared/customizations.md` for this verb.

## Process

### 1. Check preconditions

Run the shared check against the user's project — the script's path resolves from this skill's
directory, and the argument is the project. It reports what is set up and repairs nothing.

```bash
"../../shared/preflight.sh" .
```

`beads=ok` advances to step 2. Otherwise read `beads_reason`, tell the user what is missing, hand
over the command that fixes it, and **stop**:

| `beads_reason` | What is wrong | Give them |
| --- | --- | --- |
| `not_installed` | `bd` is not on PATH | `brew install beads` |
| `not_initialized` | this repository has no beads database | `bd init` |
| `unreadable` | bd found a database and could not read it | quote `beads_detail` — bd's own words are more use than a paraphrase |

**Never run the remedy.** `bd init` writes `.beads/`, git hooks, `.claude/settings.json`, and a
block in `AGENTS.md` and `CLAUDE.md`, then commits all of it. That is the user's decision, not a
repair a skill makes on its way to somewhere else.

### 2. Ask what they want to build

One open question:

> "What would you like to build? A sentence or two is enough to start."

**Then look for a concept.** If `docs/concepts/` exists, read the live concepts there — not
`archive/` — and offer the relevant one as context:

> CONCEPT-002 covers the auditing rework and looks like the frame for this. Want me to work from it?

A concept is **never required**. It carries the *why* this skill does not interview for, so taking one
usually shortens the interview and improves the Context section. What it does not carry is acceptance
criteria: a concept's **Proposed shape** is a rough direction, not a specification, and step 5 still
interviews for everything below. Do not lift criteria out of a concept and do not treat its **Open
questions** as settled.

Do not change the concept's `Status`. Work starting is `implement`'s transition to record, not this
skill's.

### 3. Check whether it already exists

Three searches, all cheap, all worth doing before spending the user's time on an interview.

**The specs** — read `docs/specs/` for a spec that already covers this. Not `archive/`.

**The codebase** — Glob and Grep for what they described. Filenames, exported identifiers, route
segments, domain nouns.

**The tracker** — the profile supplies the search command.

### 4. Report what you found

- **A spec already covers it.** Link it. Ask whether it is the same thing, and whether they want to
  revise that one rather than write another.
- **It exists in code.** Describe what is there, with file paths. Ask whether that is what they meant,
  and if not, what the distinction is.
- **An issue already covers it.** Link it. Ask whether it is the same thing.
- **Nothing found.** Say so and move on.

If the user confirms existing work covers their need, **stop the skill**. Nothing to specify.

### 5. Interview

This is the step that decides whether the specification is worth anything.

**Style.** One or two questions at a time, never a wall. Start broad and narrow. Use the user's own
vocabulary rather than translating it into jargon.

**Cover what is still unclear** — skip anything the description already settled:

1. Who the consumer is.
2. What they can do that they could not do before.
3. What starts it — an action, an event, a schedule.
4. What they observe when it works.
5. What happens when it does not — empty, unauthorized, upstream failure, bad input.
6. Which boundary conditions matter, and which are being left alone.
7. What this depends on that does not exist yet.
8. What is explicitly out of scope.

Question five feeds the `IF … THEN` criteria and question six feeds **Edge cases**. They are
different questions: five asks what the system does when something goes wrong, six asks which
situations nobody intends to handle.

**Push back on vague answers.** Name the vague word and ask for a concrete replacement:

| They said | Ask |
| --- | --- |
| "fast" | Faster than what? What latency is acceptable, and at which percentile? |
| "good UX" | What does good look like here — a reference product, a specific interaction? |
| "manage" | Which actions: create, edit, delete, reorder, archive? |
| "integrate with X" | Which part of X — their search, their booking flow, their SSO? |
| "like before" | Like which screen, which flow? Walk me through it. |
| "just works" | What is the success path, and what is the failure path? |
| "real-time" | Under a second, or under a minute? |

Before advancing, judge the answers against consumer, trigger, observable outcome, and failure
behavior. **If fewer than roughly three of the applicable ones are concrete, do not advance.** Say
which ones are thin and ask.

**Insist for at most two rounds.** If the user overrules — "just write it with what we have" — write
it, and record every unresolved item under **Open questions**. Recorded, not resolved, and never
quietly dropped.

**Raise design concerns as flags, not rulings.** When the described experience has a problem the user
may not have considered, name it once and let them decide:

> "One thing I want to flag — a destructive action behind a hover has no reachable equivalent on
> touch. How are you thinking about that?"

Cap at two rounds per concern. If it stays unresolved, it goes under **Open questions** and you move
on. You are raising a concern, not overriding a decision, and the user knows their product.

**"Like $COMPANY does it."** When the user references another product, offer once to look it up. On
yes, research it and summarize only the patterns that matter — never dump page content. Confirm the
distillation with the user before it reaches the document. On no, move on without searching.

### 6. Audit what already exists

A specification describes what the user wants **added**. If the surface already has functionality the
specification does not mention, an implementer may take the specification literally and remove it.

**Silent omission from a specification is not a deletion.** This is the rule that matters, and
authoring time is the only cheap moment to enforce it — by implementation time, nobody remembers the
specification was written against a stale picture.

For each surface the feature touches — a screen, a route, a component, an entity, a table:

1. **Read the current artifact.** Not its name, its contents. If it does not exist yet, skip; there is
   nothing to audit.
2. **Enumerate what is there.** Tabs, fields, states, columns, branches — whatever the unit of
   functionality is.
3. **Intersect with what the user described.** What did they account for? What did they leave out?
4. **If there is a real gap, ask.**

   > "The profile screen has three tabs today: Account Info, Travel Preferences, and Saved Passengers.
   > You mentioned the first two. Is Saved Passengers preserved as-is, folded into something new, or
   > intentionally going away?"

   Three outcomes:

   - **Preserved** — record it under **Existing behavior preserved**.
   - **Merged** — record how it composes, and cover it with a criterion.
   - **Removed** — this becomes its own acceptance criterion. **A removal is never an implicit
     consequence of the specification.**

Audit the surface the feature touches, not the whole application.

### 7. Shape the requirements

Cut what the user described into requirements. Each one is a capability a consumer can use and a
ticket someone can pick up.

**Requirements decompose by what a consumer can observe.** `design` decomposes by what can be built,
against the build graph, and one requirement may well become several of its tickets. Do not do that
cut here.

**The sizing question**: could one person hold this requirement in their head well enough to design
it in a single pass? If not, it is more than one requirement.

Then apply the [cohesion check](#cohesion-and-splitting) to the set. If they do not belong to each
other, say so once and offer the split.

### 8. Mockups

When the feature has a visual surface, ask whether a mockup exists.

**If the user has one**, import it — follow `../../shared/import-mockup.md`. The imported files land
under `docs/mockups/<slug>/`, keyed by surface, and the spec references that path under **Design
notes**.

**If the user wants one but does not have it**, or has no design tool, the specification proceeds
without it and the tracker issue is marked `requires-mockup`. `design` refuses to act on an issue
carrying that label, which is what keeps the gap from being forgotten.

**Do not draw a mockup inside this skill.** Producing one is separate work with its own concerns.

### 9. Recap before writing

The recap is the user's last chance to catch a misread, so it has to be substantive.

> Here is what I have. Tell me what is wrong.
>
> **Consumer**: …
> **Requirements I would write**: …
> **What starts each one**: …
> **What they observe**: …
> **Failure behavior**: …
> **Boundary conditions being left alone**: …
> **Already exists and must keep working**: …
> **Assuming without asking**: …
> **Out of scope**: …
> **Unresolved**: …
>
> Does that match what you have in mind?

**Every line is at least a sentence.** A line that collapses to a fragment or a dash means that topic
was not interviewed — go back and ask. The recap checks your work as much as the user's.

Advance when every line is substantive, the user has confirmed it, and you could write the criteria
without guessing at any of them.

### 10. Write, then confirm

Pick the identifier: read `docs/specs/`, take the highest existing number plus one, zero-padded to
three digits. Read `archive/` for this and this only — a retired identifier is never reused.

Compose the full document and **show it to the user before anything is written**. Nothing lands
unapproved. Say which optional sections you left out and why — "no Key Entities section, because the
nouns here are ordinary English" — so the user can catch an omission that was actually a gap.

Then set the status: `Ready`, unless they said they are stopping and coming back, which is `Draft`.

### 11. Write the spec

Write `docs/specs/SPEC-NNN-slug.md`, and `docs/specs/AGENTS.md` if it was missing.

Do not commit.

### 12. Mirror to the tracker

Follow the creation sequence in the tracker profile. The document is canonical and the issues are
generated from it, so this step never asks the user to re-approve content they already approved.

### 13. Wrap up

Report the spec path, its identifier, its status, every open question it carries, and the issues that
were created, with links.

If a concept framed this work, add the spec identifier to its `Related` line. That line is written
expecting a specification to complete it, and it holds a list because one concept can produce several
specs. Add the identifier and change nothing else in the file.

## Other modes

Invoking this skill on an existing spec does one of four things. Ask which if it is not obvious.

- **Promote** `Draft` to `Ready`, or **reopen** `Ready` to `Draft`.
- **Edit** a `Draft` or `Ready` spec. Adding a requirement appends the next `REQ` number; adding a
  criterion appends the next `AC` number within its requirement. Retired numbers stay retired.
  Re-mirror to the tracker afterwards.
- **Archive** a spec: set `Status: Archived`, add `Replaced by` if something took its place, move the
  file to `docs/specs/archive/`, and close or relabel its issues per the tracker profile.
- **Split** a spec into siblings, per [cohesion and splitting](#cohesion-and-splitting).

Every one of these is a user's decision. Report the state and offer; never transition a spec on your
own initiative.

## Tracker profiles

The specification is tracker-neutral. Where the mirror lands, and in what shape, is a **tracker
profile** — one directory per tracker, exactly as `scaffold` handles surfaces.

| Tracker profile | Covers | Status |
| --- | --- | --- |
| `github` | GitHub Issues, optionally with a GitHub Project | **supported** |
| `jira` | Jira Cloud and Data Center | planned |
| `linear` | Linear | planned |

A tracker is **supported** only when `trackers/<name>/PROFILE.md` is complete. Nothing else
counts. A planned profile is an exit, not a menu choice — if the user's tracker is Jira, say plainly
that `specify` does not mirror to it yet and stop.

**Resolution.** GitHub is currently the only supported profile, so there is no question to ask. State
that the mirror will land in GitHub Issues and confirm the repository. When more profiles exist, this
step reads the project's recorded choice and asks only when there is none.

**Capabilities.** Each profile declares what its tracker can do — hierarchy, relations, custom fields,
markup, labels — and states a fallback for every capability it lacks. Read the profile's capability
table before writing, and follow its fallbacks rather than improvising around a missing feature.

**The spec is written even when the mirror fails.** If tracker authentication is wrong, the document
still lands on disk — it is the deliverable. Give the user the exact command to fix the tracker and
say the mirror is pending. Do not work around it and do not discard the spec.

## Rules

- **Nothing is written without the user confirming the full document first.**
- **The document is canonical.** Tracker issues are generated from it and regenerated on later runs.
  Never treat a hand-edited issue body as the source of truth.
- **The specification says what, never how.** No file paths, no libraries, no services, no schema. A
  domain noun may be named and defined; its fields, types, and relations may not.
- **Acceptance criteria are EARS, and nothing else is.** Pick the pattern that fits rather than
  writing everything as `WHEN … THEN`.
- **"so that" is mandatory** in every user story.
- **Criteria are observable in a running system**, and the instrumentation to make them so is part of
  the requirement — surfaced out loud, never absorbed silently.
- **Criteria exist for coverage, not symmetry.** Requirements with different criteria counts are
  expected.
- **Numbering is append-only at every level.** Retired numbers are never reused.
- **Status describes the document, never the work.** Work state belongs to the tracker.
- **Silent omission is never a deletion.** A removal is always its own criterion.
- **Mockups are keyed by surface**, never filed under a spec.
- **Push back once, then defer.** On vagueness, on design concerns, on cohesion. The user knows the
  domain and you may be wrong.
- **Unresolved is recorded, not dropped.** Open questions go in the document.
- **Never overwrite a file that has drifted.** Show the difference and ask.
