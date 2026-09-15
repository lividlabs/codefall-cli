# docs/designs — operative rules

- `archive/` is history. Do not read it unless the user asks about a superseded design by name.
- A design's identifier is permanent. `DESIGN-007` means the same document after it is archived.
- Local task identifiers (`T1`, `T2`) are permanent within a design and append-only. A retired one
  is never reused.
- Once a Task Plan has collapsed to an ID mapping, Beads is authoritative for the tasks. Never
  restore a task table over it.
- ADRs live in `docs/adrs/` and are never rewritten. A revision is a new, superseding ADR.
- Status describes the document, never the work. Beads holds work state.
- Status transitions are `/design`'s to make, never a hand edit.
