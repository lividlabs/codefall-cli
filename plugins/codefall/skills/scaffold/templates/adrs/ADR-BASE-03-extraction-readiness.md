# ADR-BASE-03: Extraction Readiness

## Status

Accepted — <date>

## Context

ADR-BASE-02 claims that "component boundaries double as extraction (microservice) seams." That claim
is the main thing package-by-component buys over package-by-feature, and it is not free — it is only
true if the components stay independent in ways the layout alone cannot guarantee.

The profile's boundary-enforcement ADR covers **import** coupling: facades, layer direction,
cross-component access. That is necessary and it is not sufficient. A codebase can satisfy every one
of those rules and still take months to split, because the couplings that make extraction expensive
are not imports:

- two components' tables joined by a foreign key,
- one transaction spanning two components' writes,
- a rich domain entity travelling across a facade,
- a shared data-access module every component depends on.

None of those show up in an import graph. All of them turn "move this folder into a service" into a
rewrite.

**We build a monolith on purpose.** A single deployable is the right default: no network in the way
of a use case, one transaction where the domain wants one, one place to run and debug. The decision
here is not to distribute anything. It is to keep the *option* cheap, so that extracting a component
later is a deployment change rather than a redesign.

**This ADR is gated on whether the surface could ever be split at all**, not on which layout was
chosen. A backend or service can be: keep these rules whether or not anyone currently intends to
split it, because intentions are revisable and coupling built in the meantime is not. A CLI, desktop
app, mobile app, or library cannot — there is no service to extract to, so the rules have no referent
and their cost buys nothing.

Note what does *not* gate it. Having no current appetite for microservices is not a reason to skip
these rules; ADR-BASE-02 is explicit that a missing extraction goal is not even a reason to abandon
package-by-component. And a surface on ports-and-adapters has no component boundaries to talk about
yet, so the rules are not actionable there — but if it grows into components as ADR-BASE-02 expects,
they apply from that moment, and the decisions made in the meantime are what make that growth cheap
or expensive.

## Decision

Components stay independently extractable. Concretely, five rules.

### 1. Each component owns its data

- A component's tables are private to it. No other component reads or writes them directly, in code
  or in a query.
- **No foreign keys across components.** Reference another component's entity by its **id**, and
  resolve it through that component's facade. A database-level FK is a hard join between two things
  you intend to be able to separate.
- Give each component its own schema or table prefix from day one. It costs nothing in a monolith and
  it makes the ownership boundary legible — and later, mechanical.

### 2. A transaction never spans two components

- One transaction commits one component's writes. A workflow that must change two components does so
  as an ordered sequence, with the second step made idempotent and retriable, or via an outbox.
- This is the rule that decides whether extraction is a move or a rewrite, because a cross-component
  transaction becomes a distributed transaction the moment there is a network in the middle.
- It is also the most expensive rule to hold. If a project knowingly breaks it, that is a legitimate
  call — **record it in the decision log as accepted extraction debt**, naming the components and the
  workflow. Unrecorded, it is discovered during the extraction it makes impossible.

### 3. Facades exchange contracts, not entities

- What crosses a component boundary is a **data shape** — a DTO, a record of primitives and ids — not
  a domain entity with its methods and invariants.
- Passing `orders.Order` into `billing` couples billing to orders' *model*, so every change to that
  entity is a change to billing. Passing an `OrderSummary` couples it to a contract you chose.
- The entity stays inside the component that owns it. This is the boundary rule that most often gets
  quietly broken, because passing the object you already have is easier than defining the shape you
  mean.

### 4. Synchronous only where the domain requires an answer

- A cross-component call that needs an immediate result stays synchronous. One that only needs the
  other component to *know* something becomes an event.
- This is not "use events everywhere" — indirection has its own cost, and in a monolith a direct call
  is usually right. It is a recognition that every synchronous cross-component call is a future
  network hop, so the ones that did not need to be synchronous are the ones that hurt later.

### 5. Shared means technical, never data

- Shared modules hold **technical primitives**: clock, logger, id generation, result types, config
  access.
- A shared ORM layer, schema module, or migrations bundle that every component imports satisfies the
  "shared modules can't import components" rule while welding every component to one data model.
  Splitting one out then means splitting the shared module too, which is the hardest kind of change.

### Frontends

The same principle applies where micro-frontends are a possibility, and the failure modes rhyme:

- A feature owns its state slice. One global store that every feature reads and writes is the
  frontend's shared data layer.
- Router state is navigation, not a message bus. A feature that discovers what another feature did by
  reading the URL is coupled through it.
- The design system is a versioned dependency, not a folder features reach into.

## Consequences

- Extraction becomes a deployment decision — take the folder, give it a process and an API — instead
  of a project. That is the entire return on package-by-component.
- Cross-component reads cost a facade call rather than a join, so some queries are two round trips
  in-process where SQL could have done one. Accepted deliberately; if a specific path makes this
  unacceptable, that is a real finding and belongs in the decision log with the reason.
- Denormalised reads become normal: a component keeps the small projection of another's data it needs
  rather than joining for it.
- **Most of these rules are review concerns rather than lint rules.** Some can be checked — the types
  crossing a facade, cross-schema references in migrations, imports of a shared data module.
  Transaction spans and synchronous-versus-event judgement cannot be, and are decided by people. That
  is the normal condition of architectural guidance, not a defect in it.
- Because enforcement is partial, the decision log carries the weight: a knowingly broken rule that is
  written down is a debt, and one that is not is a trap.
- On a surface that cannot be split at all — a CLI, a desktop or mobile app, a library — the rules
  have nothing to refer to and this ADR should not have been emitted. That is a question about the
  surface, not about its layout: such a project may still be package-by-component, and should be if
  it has separable capabilities.

## Related

- ADR-BASE-02 — Package-by-component (the claim this ADR makes true)
- ADR-BASE-01 — Clean Architecture (entities stay in the component that owns them)
- The matched profile's **boundary-enforcement** ADR (the import half of the same concern)
