# ADR-GO-02: Boundary Enforcement

## Status

Accepted — 2026-08-16

## Context

Package-by-component (ADR-BASE-02) is only real if the boundaries are enforced; unenforced, it degrades
into package-by-feature with everything public. What enforcement costs depends on the language's
visibility model, which is why this ADR is per-profile.

Go is the favorable case. **`internal/` directories are enforced by the compiler**: a package under
`.../x/internal/...` is importable only from the tree rooted at `x`, and nesting works, so
`internal/orders/internal/domain` is reachable from `internal/orders/...` and nowhere else. A
reach-around is a build failure that no comment can suppress.

But the compiler rejects import *cycles*, not *outward* imports. A `domain/` package importing its
sibling `infrastructure/` is acyclic, compiles clean, and passes `go vet`. So the facade half of the
problem is free and the layer half is not.

## Decision

### Component unit

- A component is a **directory under `internal/`**, and its root package **is** the facade. Exported
  identifiers there are the public API; everything else lives in the component's nested `internal/`,
  split into `domain/` `application/` `infrastructure/` `presentation/`.
- Use separate **modules** only for genuinely cross-repo sharing. Within a repo, directories plus
  `internal/` already give real enforcement, and multi-module repos cost bookkeeping for nothing.
- A library module intended for import keeps the per-component nested `internal/` but places the
  component directories at the module root rather than under a root `internal/`.

### Tooling

- **The compiler is primary.** It covers the facade rules with no configuration and no opt-out.
- **`depguard`, via `golangci-lint`**, covers what the compiler will not: import direction inside a
  component, and the shared-module rule. Checked in CI and in the editor through the language server.
- **`go-arch-lint`** optional, as a graph-level gate if the `depguard` rules stop expressing the
  architecture cleanly.

**Use strict allow-lists, not deny-lists.** `depguard`'s `deny.pkg` is prefix matching with no glob
support, so a deny-list has to name every forbidden package path literally and cannot be written once
for all components. `list-mode: strict` inverts it — anything not allowed is forbidden — which is
both shorter and safer, because a layer gains no new permissions when someone adds a package.

```yaml
version: "2"
linters:
  default: none
  enable: [depguard]
  settings:
    depguard:
      rules:
        domain-layer:
          list-mode: strict
          files: ["**/internal/domain/**"]
          allow: [$gostd]
        application-layer:
          list-mode: strict
          files: ["**/internal/application/**"]
          allow:
            - $gostd
            - <module>/internal/<component>/internal/domain
```

`domain/` allowing only `$gostd` needs no module path at all and states the rule positively: the
domain imports the standard library and nothing else. `application/` must name its own component's
`domain` package literally, which is why **this file is generated at scaffold time rather than copied
from a template** — the module path and component names are not known until then.

### When

- Wire it in **at the project scaffold, day one**, same as every profile. The `internal/` layout is
  the harder half to retrofit: moving packages after the fact rewrites every import path in the repo.

### Rules to enforce, and by what

1. **Facade-only imports** — no reaching past a component's root package into its internals.
   *Compiler.*
2. **Clean layer rule** — inward only: `domain` ← `application` ← `infrastructure` / `presentation`;
   no outward or sideways imports. *Linter.*
3. **Cross-component only via public facades.** *Compiler* — the same nested `internal/` that gives
   rule 1.
4. **Shared modules can't import components** — one way. *Linter*; both are ordinary importable
   packages and the compiler has no opinion.
5. **Inner rings can't import delivery or gateway packages** — no `net/http`, `database/sql`, driver
   packages, or CLI framework in `domain/` or `application/`. *Linter.* This is the Go reading of the
   frontend rule: the inner rings name no transport and no storage.

## Consequences

- Rules 1 and 3 are free, absolute, and cannot be disabled — stronger than any lint-based scheme, and
  the reason this ADR is a third the size of the `typescript-react` one.
- The cost is verbosity in paths. `internal/orders/internal/application` reads awkwardly, and this is
  the price of compiler-backed facades. Do not flatten it to make imports prettier.
- The `depguard` block in `.golangci.yml` *is* the encoded layer rule; keep it in review scope. It is
  generated per project, so it is not a file to diff against a template.
- A `depguard` rule whose `files` patterns match nothing reports `0 issues` and exits 0 —
  indistinguishable from a rule that works. Prove it by committing a deliberate violation, watching
  it fail, and reverting.
- Verification has two halves here, unlike other profiles: a cross-component reach-around must fail
  `go build`, and an outward layer import must fail `golangci-lint run` **while still compiling**.
  Checking only the first tests the compiler, not your configuration.

## Related

- ADR-BASE-02 — Package-by-component (the boundaries being enforced)
- ADR-BASE-01 — Clean Architecture (the layer rule at #2)
- ADR-GO-01 — Dependency injection (why `application/` names no implementations)
