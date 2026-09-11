# ADR-003: Pure Shared Modules in the Inner Layers

## Status

Accepted — 2026-08-27

## Context

The second component copied the first. `internal/initcmd/` writes `.codefall/settings.json` and
`internal/doctor/` reads it, so both need the same definition of that file's format: the version
constant, the schema `$id`, the tracker names, the repository pattern, the tables of which fields
each tracker's block carries, and the validator built from those tables. Both also grew the same
three-line `firstLine` helper for turning an external tool's chatty output into one line of a
report. Each of those now exists twice, once in each component's `domain/` or `application/`.

Nobody chose the duplication; the layer rules produced it. ADR-GO-02 encodes the Clean Architecture
direction rule as `depguard` strict allow-lists, and those allow-lists are deliberately narrow:
`domain/` may import the standard library and `github.com/samber/mo` (ADR-GO-03), and `application/`
adds only its own component's `domain/`. Anything not named is denied. A second component that needs
the same format as the first therefore has exactly two options inside the current rules — copy it, or
reach for something the allow-list denies — and the first copy is the one that gets written.

`internal/shared/` already exists for this class of problem. ADR-GO-02 rule 4 admits shared modules
as one-way dependencies — components import them, never the reverse — and the `ui` and `process`
modules realise that rule today. But the strict allow-lists keep the inner layers off `internal/shared/`
entirely, and for those two modules that is correct rather than incidental: `ui` imports
`charm.land/lipgloss/v2` and `charm.land/bubbletea/v2`, and `process` imports `os/exec`. Letting
`domain/` import either would put a CLI framework and process execution inside the innermost ring,
which is what ADR-GO-02 rule 5 exists to prevent. The blanket denial is a blunt instrument aimed at a
real target.

**The distinction that matters, and that is easy to conflate: the layer rules are about what a package
may *depend on*, not about where a package *sits*.** `internal/shared/ui` is denied to `domain/`
because of what it imports, not because of the directory it lives in. A shared module whose own
imports are exactly the ones `domain/` is already allowed adds nothing to the domain's dependency
surface — importing it is transitively identical to importing the standard library and `samber/mo`.
The current configuration cannot express that difference, so it treats every shared module as if it
were `ui`.

The duplication is not currently drifting. Each component's schema test holds its copy of the
constants equal to `schemas/settings.schema.json`, so the two copies are pinned to a third thing
rather than to each other, and a change to one that was not made to the other fails. That is what
made the duplication safe enough to accept when the shared modules landed. It does not make it right:
there are two definitions of one file format, in two components, each with its own schema test, and
adding a tracker means editing both sides and the schema.

Three options were considered.

**(a) Keep duplicating, pinned by the schema tests.** Costs nothing to decide and has already been
tried. The pinning works, so the risk is not silent divergence — it is that every change to the
format is three edits instead of one, that the second component's copy reads as if `initcmd` owns
the format when `doctor` equally does, and that a third component repeats the pattern. The cost is
paid per change, forever.

**(b) Make settings a component with a facade.** The format would move to `internal/settings/`, and
`doctor` and `initcmd` would reach it the way clean layers reach anything outside themselves: a
gateway interface declared in each `application/`, implemented in each `infrastructure/` over the
settings facade, and injected. That is architecturally unimpeachable and completely out of
proportion. It buys inversion nobody needs — the settings format has no IO, no policy, and nothing
to fake in a test — and pays for it with two interfaces, two adapters, two registrations, and a
mapping layer, all wrapped around a table of constants. It also puts a `domain/` package's constants
behind a runtime indirection, which makes `settings.Version` a value that is injected rather than a
fact that is true.

**(c) Admit shared modules that are themselves pure into the inner layers.** Name the modules whose
imports are exactly the inner layers' own, allow those modules in `domain/` and `application/`, and
enforce their purity with a lint rule of their own so that the property the permission rests on
cannot quietly stop holding.

Option (c) is chosen. It is the option that says what is actually true: the constraint on `domain/`
is a constraint on dependencies, and a package with no dependencies beyond the domain's own violates
nothing by being imported there.

## Decision

### What a pure shared module is

A **pure shared module** is a package under `internal/shared/<module>/` that imports only:

- the Go standard library, and
- `github.com/samber/mo` (ADR-GO-03),

and nothing else, transitively within this repository. That is the same import set `domain/` is
allowed, which is the whole basis for the permission below: a pure module is not a new dependency
for the inner rings, it is code that already satisfies their rules and happens to live outside a
component because more than one component needs it.

Transitivity is part of the definition: what matters is what a pure module's imports drag in, not
just what it names. It may not import a shared module that is not pure, and it may not import a
component — ADR-GO-02 rule 4 already forbids the second, in either direction of purity.

One pure module importing another satisfies the definition: a pure module's imports are already the
standard library and `samber/mo`, so another pure module adds nothing beyond what the importing
module is already allowed. The purity rule's allow-list says so directly — each pure module is named
there too, alongside the standard library and `samber/mo` — which is what lets `internal/shared/settings`
import `internal/shared/text` while the rule still denies everything impure.

### Who may import one

**`domain/` and `application/` of any component may import a pure shared module.** This refines the
allow-lists in ADR-GO-02 rather than superseding them: rules 2 and 5 hold unchanged — dependencies
still point inward only, and the inner rings still name no transport, no storage, and no CLI
framework — because a pure module cannot name any of those either. Rule 4 holds unchanged as well:
the import stays one-way.

A shared module that is **not** named as pure is impure by default and keeps its current audience:
`presentation/`, `infrastructure/`, the component's facade file, and `main`. `ui` and `process` are
impure, deliberately, and stay there.

### How it is enforced

Each pure module is named in `.golangci.yml` **four times**:

1. In the `domain-layer` rule's `allow` list, so `domain/` may import it.
2. In the `application-layer` rule's `allow` list, so `application/` may import it.
3. In the `pure-shared-modules` rule's `files` list, holding that module's path under
   `list-mode: strict`.
4. In the `pure-shared-modules` rule's own `allow` list, alongside `$gostd` and `github.com/samber/mo`.

Entries 3 and 4 together are what make the first two safe. Without the third, "pure" is a claim in a
comment and the day someone adds an import to the module is the day two `domain/` packages silently
gain a dependency on it. With it, an impure import inside a pure module fails `golangci-lint run`
while still compiling — the same shape of failure ADR-GO-02 asks for everywhere else, and the reason
the lint half of the verification exists at all. The fourth is what lets a pure module import
another: naming a module in its own rule's allow-list is the only way one pure module can name
another without also opening the rule to `ui`, `process`, or anything else outside this list.

The direction rule needs no new configuration: the existing `shared-modules` rule already denies
every component's facade from anything under `internal/shared/`, and a pure module is under
`internal/shared/`.

### What belongs in one, and what does not

A pure module holds **format definitions and small helpers**: the definition of a file both
components read and write, the parsing of a name both components accept, a string utility both
components had written twice. The test is whether every component would define it the same way. If
the answer is no — if the thing encodes what one component does rather than what the project's data
is — it is that component's business rule and stays in that component's `domain/`.

Concretely, for the settings format: the version, the schema id, the tracker names, the repository
pattern, the field tables, and the validator built from them are the format, and they are shared.
`initcmd`'s `Settings` value object and its `NewSettings` constructor are how `initcmd` builds a
settings file from a survey's answers, and they stay in `initcmd`'s `domain/`, calling the shared
rules. `doctor`'s check identities and its report are how `doctor` reports on one, and they stay in
`doctor`'s.

### What this does not change

**ADR-001 is untouched.** A type from a pure shared module is not a component's domain entity, but
that does not make it a contract either. What a *facade* accepts and returns is still declared in
the component's own root package. A facade that returned `settings.Document` because "it is not a
domain type" would be exporting a shape the component does not own, and the next change to the
format would be a change to that facade's signature — which is the coupling ADR-001 exists to
prevent, arriving by a different route.

## Consequences

- **The settings format gets one definition and one schema test.** Adding a tracker becomes one row
  in the shared field table plus one `oneOf` branch in the published schema, with one test holding
  them equal, instead of the same edit in two components and two tests.
- **Four configuration entries per pure module, and nothing enforces that all four were written.**
  This is the same class of gotcha as the per-component `application-layer` entry: `depguard`'s
  allow-lists are literal prefix matches with no globs, so a pure module that is only half declared
  either cannot be imported where it should be, is not held to purity, or cannot be imported by
  another pure module. It belongs in AGENTS.md's gotchas beside the existing one.
- **The purity rule matches no files until the module exists**, and a `depguard` rule that matches
  nothing reports `0 issues` and exits 0 — indistinguishable from one that works. The existing rule
  applies: prove it against a deliberate violation once the module is there.
- **A pure module is importable from everywhere, which makes it a place things drift to.** `ui` and
  `process` are protected from becoming junk drawers by their audience — a `domain/` package cannot
  reach them, so nothing gets put there for a `domain/` package's convenience. A pure module has no
  such protection. The question to ask before adding to one is "would every component define this the
  same way?", and the answer for anything that encodes one component's behaviour is no.
- **Purity is a property that can be lost, and losing it is informative.** If a pure module comes to
  need a dependency — a schema library, a file system, a clock it calls rather than receives — the
  `pure-shared-modules` rule fails, and the fix is not to widen the allow-list. It is to recognise
  that the module has grown behaviour rather than definitions, and that behaviour with dependencies
  is what option (b) above is for: a component with a facade, reached from the inner layers through a
  gateway. The lint failure is the signal that the earlier judgement has expired.
- **The blanket rule "inner layers never import `internal/shared/`" is gone**, and reading
  `.golangci.yml` is now the only way to know which shared modules an inner layer may name. That is a
  real loss of a rule that could be stated in one sentence and checked by eye. It is accepted because
  the replacement is checked by a tool rather than by eye, and because the one-sentence rule was
  producing duplication.

## Related

- ADR-GO-02 — Boundary enforcement. Rule 4 gives shared modules their direction; rules 2 and 5 give
  the inner rings their allow-lists. This ADR **refines** those allow-lists and supersedes none of
  the rules: the direction is unchanged and the inner rings gain no dependency they did not already
  have.
- ADR-BASE-02 — Package-by-component. Shared modules as a library of small, focused packages, each
  with its own facade, rather than one kernel — a pure module is one of those, held to a stricter
  import rule.
- ADR-001 — Facade contracts. What crosses a facade is still declared in the component's root
  package; a pure module's types are not contracts.
- ADR-GO-03 — Optional values. `github.com/samber/mo` is the one non-standard-library import a pure
  module may carry, because it is the one the domain already carries.
