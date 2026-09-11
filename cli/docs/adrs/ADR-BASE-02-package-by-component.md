# ADR-BASE-02: Package-by-Component

## Status

Accepted — 2026-08-16

## Context

Clean Architecture's rings (ADR-BASE-01) are conceptual, not a directory layout — so we still need to
choose how code is organized at the top level.

Simon Brown's "The Missing Chapter" (in *Clean Architecture*) evaluates four schemes: **package by
layer**, **package by feature**, **ports and adapters**, and **package by component**. Package-by-
layer and package-by-feature are the weaker two (they don't enforce boundaries — code can bypass a
layer or reach into a feature's internals). The two sound ones are **ports-and-adapters** and
**package-by-component**. Uncle Bob's *Screaming Architecture* adds the guiding principle: the top
level should reflect the **domain**, not the framework or the layers.

This baseline chooses **package-by-component**, and names when **ports-and-adapters** is the better
fit instead.

## Decision

### Organize by component (capability), not by layer

- **Top-level = capabilities** (components), each a folder that bundles its own Clean layers
  (`domain/` `application/` `infrastructure/` `presentation/`) *inside* it.
- Each component exposes **one public API via a facade**; everything else is internal. The facade is
  whatever the language gives you — an `index.ts` barrel in TypeScript, a module with `internal` /
  `pub(crate)` / lowercase-unexported members elsewhere. The surface profile's **boundary-enforcement**
  ADR names it.
- Cross-cutting technical capabilities live in a **library of small, focused shared modules** (each
  with its own facade) — not duplicated per component, and **not** dumped into one god "kernel."
- The top level *screams* the domain (`orders/`, `billing/`, …), not the framework.

### Per surface

- Applies to backend and frontend alike. On a **frontend**, a "component" is a **feature module** —
  not a UI/React component (name collision). The shared UI / design system is its own package.
- On a frontend, the Clean inner rings still hold (framework-agnostic domain + use cases); the UI
  framework and gateways are the outer edge (see the profile's frontend-state ADR, where it has one).

### When to prefer ports-and-adapters instead

Choose **ports-and-adapters** (organize inside = framework-free core vs outside = adapters) when:
- the domain is one cohesive thing rather than several separable capabilities, or
- your capability boundaries are genuinely unknown and you want zero commitment yet.

These are alternatives, so **either one alone is enough** — which is why the list must stay short and
must not include proxies for the real question.

Package-by-component's advantage — an enforced boundary around each capability, which also doubles as
a service seam if you ever want one — only pays off once boundaries are known.

**Not having a service-extraction goal is not a reason to choose ports-and-adapters.** The payoff here
is encapsulation, not deployment: without an enforced boundary, package-by-component degrades into
package-by-feature with everything public, and that degradation is just as bad in a CLI that will
never be split as in a platform that will. A tool with three separable capabilities and no intention
of ever deploying them apart still gets a screaming domain, per-capability facades, local reasoning,
and a contained blast radius. Extraction readiness is what becomes moot when nothing
will ever be extracted — the layout does not.

### For this project

`codefall-cli`'s capabilities are **not named yet**, which is the second of the two conditions above.
We choose **package-by-component anyway**, deliberately: in Go the component boundary is the
compiler-enforced `internal/` layout (ADR-GO-02), and retrofitting it later rewrites every import path
in the repo — a cost ports-and-adapters would defer rather than avoid.

What carries the weight instead is **start coarse and split**. The tree ships with **no components at
all** — `internal/` is empty, and `cmd/codefall/main.go` prints a banner without wiring anything. The
layout, the `internal/` nesting convention, and the `depguard` rules are in place and proven; the
first real capability fills them in. Start with one coarse component, not three speculative ones.

## Consequences

- The structure screams the domain. Component boundaries double as extraction (microservice) seams
  only on a surface that can be split — this CLI cannot be, so what the boundary buys here is
  encapsulation and local reasoning, not deployment options.
- It asks you to commit to capability boundaries — **start coarse and split** as the domain
  clarifies; re-slicing an existing boundary is the expensive case.
- Cross-component workflows go through facades / app-layer coordination rather than reaching across.
- **Enforcement is mandatory** (the profile's boundary-enforcement ADR): without it,
  package-by-component silently degrades into
  package-by-feature with everything public.

## Related

- ADR-BASE-01 — Clean Architecture (the layers that nest inside each component)
- The matched profile's **boundary-enforcement** ADR (what makes the facades real)
- Simon Brown, "The Missing Chapter"; R. C. Martin, "Screaming Architecture"
