# docs/designs — operative rules

- `archive/` is history. Do not read it unless the user asks about a superseded design by name.
- A design's identifier is permanent. `DESIGN-007` means the same document after it is archived.
- Local task identifiers (`T1`, `T2`) are permanent within a design and append-only. A retired one
  is never reused.
- Once a Task Plan says `Created in Beads`, Beads is authoritative for work state. The table keeps
  the tasks, edges, and design refs; it never grows a status column.
- A design bead's title, edges, and design ref change only through `/design` (Revise), which edits
  the row and the bead together — never with `bd` directly.
- A bead labelled `design-revision` whose `spec_id` is a design's path is a request to revise that
  design. `/design` lists them at its start and closes each one it settles.
- ADRs live in `docs/adrs/` and are never rewritten. A revision is a new, superseding ADR.
- Status describes the document, never the work. Beads holds work state.
- Status transitions are `/design`'s to make, never a hand edit.
