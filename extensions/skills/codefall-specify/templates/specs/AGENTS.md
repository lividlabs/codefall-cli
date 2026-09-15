# docs/specs — operative rules

- `archive/` is history. Do not read it unless the user asks about a superseded spec by name.
- Spec, requirement, and criterion identifiers are permanent and append-only. A retired number is
  never reused, and `SPEC-003` means the same document after it is archived.
- The spec document is canonical. Tracker issues are generated from it and are regenerated, not
  hand-edited.
- Mockups are keyed by surface under `docs/mockups/<slug>/`, never filed under a spec. One mockup
  can serve several specs.
- Status transitions are `/specify`'s to make, never a hand edit.
