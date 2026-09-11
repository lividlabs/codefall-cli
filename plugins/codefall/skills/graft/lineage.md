# Template lineage

What every current template used to be called, where it lived, and when it first shipped. `graft`
reads this to recognise a rename as a rename — a project on 0.2.x holding
`ADR-003-dependency-injection.md` has `ADR-TS-01`, not a stray file — and to know which historical
paths and tags to ask git or GitHub for.

**Maintenance rule:** any PR that renames, moves, or retires a template adds a row here, in the
same PR. Without the row, graft sees a deletion plus an addition and reports nonsense.

## Repo layout epochs

The templates root moved when the repo split into marketplace and plugin. Historical fetches must
use the path that was current at the target release.

| Releases | Templates root (repo-relative) |
| --- | --- |
| 0.1.0 – 0.2.1 | `skills/scaffold/templates/` |
| 0.3.0 – | `plugins/codefall/skills/scaffold/templates/` |

## Release tags

Tag naming has been inconsistent, so try both forms before concluding a tag is missing:

| Release | Tag(s) |
| --- | --- |
| 0.1.0 | none — never tagged; reachable only via the 0.1.1 tree's history |
| 0.1.1 | `0.1.1` and `v0.1.1` (duplicate) |
| 0.2.0 | `0.2.0` |
| 0.2.1 | `v0.2.1` |
| 0.3.0 | `v0.3.0` |
| 0.3.1 | `v0.3.1` |

`CHANGELOG.md` declares bare version numbers as the convention going forward; history disagrees
with itself, which is why this table exists.

## Templates

Paths are relative to the templates root for the epoch in question. A project's emitted copy lives
in `docs/adrs/` under the same basename.

| Current template | First shipped | Former identity (releases) |
| --- | --- | --- |
| `adrs/_TEMPLATE.md` | 0.1.0 | same name throughout |
| `adrs/ADR-BASE-01-clean-architecture.md` | 0.1.0 | `adrs/ADR-001-clean-architecture.md` (0.1.0 – 0.2.1) |
| `adrs/ADR-BASE-02-package-by-component.md` | 0.1.0 | `adrs/ADR-002-package-by-component.md` (0.1.0 – 0.2.1) |
| `adrs/ADR-BASE-03-extraction-readiness.md` | 0.4.0 | — |
| `surfaces/typescript-react/PROFILE.md` | 0.1.0 | same name throughout |
| `surfaces/typescript-react/AGENTS.md.skeleton` | 0.1.0 | same name throughout |
| `surfaces/typescript-react/adrs/ADR-TS-01-dependency-injection.md` | 0.1.0 | `surfaces/typescript-react/adrs/ADR-003-dependency-injection.md` (0.1.0 – 0.2.1) |
| `surfaces/typescript-react/adrs/ADR-TS-02-frontend-state.md` | 0.1.0 | `surfaces/typescript-react/adrs/ADR-004-frontend-state.md` (0.1.0 – 0.2.1) |
| `surfaces/typescript-react/adrs/ADR-TS-03-boundary-enforcement.md` | 0.1.0 | `surfaces/typescript-react/adrs/ADR-005-boundary-enforcement.md` (0.1.0 – 0.2.1) |
| `surfaces/go/PROFILE.md` | 0.3.0 | — |
| `surfaces/go/AGENTS.md.skeleton` | 0.3.0 | — |
| `surfaces/go/adrs/ADR-GO-01-dependency-injection.md` | 0.3.0 | — |
| `surfaces/go/adrs/ADR-GO-02-boundary-enforcement.md` | 0.3.0 | — |
| `surfaces/go/adrs/ADR-GO-03-optional-values.md` | 0.3.0 | — |

The 0.3.0 rename changed identifiers inside the files as well as the filenames — `ADR-001` became
`ADR-BASE-01` in every cross-reference. A renamed project file therefore cites old identifiers
throughout its body, and other docs cite the old identifier of the renamed file; both are part of
what graft reports.

Nothing has been retired yet. When something is, it gets a row here with its full former identity
and the release that dropped it.
