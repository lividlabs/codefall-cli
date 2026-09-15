# Bundled templates and the surface catalog

What ships with this skill and which surfaces it can scaffold. Read at step 1, before the stack
question. Paths are relative to this skill's directory.

## Stack-agnostic templates

Every project gets these:

| Path | What it is |
| --- | --- |
| `../templates/adrs/ADR-BASE-01-clean-architecture.md` | The dependency rule and the four layers |
| `../templates/adrs/ADR-BASE-02-package-by-component.md` | Top-level organization; when to prefer p&a |
| `../templates/adrs/ADR-BASE-03-extraction-readiness.md` | Keeping components cheap to split out later — **surfaces that can never be split skip this one** |
| `../templates/adrs/_TEMPLATE.md` | Thin ADR template for new decisions |

## Surface profiles

Per surface, under `../templates/surfaces/<name>/` — a `PROFILE.md`, an `AGENTS.md.skeleton`, and
the profile's own ADRs.

- `typescript-react` supplies DI (Inversify), frontend state (TanStack Query / Zustand /
  `useState`), and boundary enforcement (`eslint-extension-boundaries`).
- `go` supplies DI (`samber/do`), boundary enforcement (`internal/` packages plus `depguard`), and
  optional values (`samber/mo`'s `Option`).

## ADR identifiers

Three separate namespaces that never interact:

| Namespace | Who owns it | Examples |
| --- | --- | --- |
| `ADR-BASE-NN` | the stack-agnostic core in `templates/adrs/` | `ADR-BASE-01`, `ADR-BASE-02` |
| `ADR-<PREFIX>-NN` | a surface profile, prefix declared in its `PROFILE.md` | `ADR-TS-01`, `ADR-GO-02` |
| `ADR-NNN` | **this project's own** decisions, starting at `ADR-001` | `ADR-001` seam ADR |

Numbering restarts at 01 inside each profile, so a project matching two profiles gets `ADR-TS-01`
and `ADR-GO-01` side by side. A profile that has no use for a concern does not ship that ADR; there
is no gap to explain. Project ADRs are a separate sequence starting at `ADR-001`, never a
continuation of the inherited ones.

The inherited ADRs ship **Accepted** with a real date. Amend one only when the interview requires
it; if you amend, edit the Context and Decision so the file reads as a decision made for this
project, and say what you changed in the final report.

## Surface catalog

Profiles are scoped to a **surface**, not to a kind of product. A project composes as many profiles
as it has surfaces. "Desktop app" is not a profile — depending on where its domain logic lives, it
is either one web surface or a web surface plus a native one.

| Surface profile | Covers | Status |
| --- | --- | --- |
| `typescript-react` | TypeScript/Node backends; React frontends, web and Native — including Tauri, Electron, and RN apps whose native side is only wiring | **supported** |
| `go` | Go services, APIs, workers, daemons, and CLIs that hold domain logic | **supported** |
| `rust-native` | Rust services, and Tauri shells that hold domain logic | planned |
| `kotlin-native` · `swift-native` | Android, iOS | planned |
| `dart-flutter` | Flutter, mobile and desktop | planned |
| `python` · `java` | backends and services | planned |

A profile is **supported** only when `../templates/surfaces/<name>/PROFILE.md` is complete. Nothing
else counts — not a language this skill mentions, not one you know well, not one that is "basically
the same as" a supported profile.

A Tauri desktop app or a React Native mobile app is scaffoldable today when its native side is a
thin shell; neither is when that side carries domain logic. That is a question for the user, not a
guess from the directory listing — React Native ships `android/` and `ios/` empty.
