# codefall-design — where it came from

What this took from elsewhere, and what it deliberately did not, so nobody re-adds it. None of this
is instruction — the skill is the instruction.

## Taken

**From Kiro**, the primary model: the design document's section set, and the principle that the
design document is architecture only, with tasks living somewhere else. Kiro requires six sections —
Overview, Architecture, Components and Interfaces, Data Models, Error Handling, Testing Strategy —
and three of those are conditional here, because a required section with nothing behind it teaches
readers to skim. Also from Kiro: research inline rather than in sibling files, and bug fixes routed
away from the full artifact, which is tier 0.

**From spec-kit**: the Technical Context block, which forces concreteness where prose hand-waves;
the shape of the Alternatives table; and numbered-identifier traceability, which here is the local
task IDs and the Design ref column.

**From `codefall-envision`**: the `Related` row as a labeled list rather than a set of rows. A
new kind of upstream artifact costs a label rather than an edit to the template and the skill; when
a bug-report verb lands, its designs carry `bug: BUG-012` and nothing else moves.

**From `codefall-specify`**: status describes the document, never the work. It diverges from
`codefall-envision`, which carries an `Active` state because a vision has no tracker
representation to carry it. A design has one — the beads it created — so there is no `Active` here.

## Dropped

**From spec-kit**: the constitution check, the phase gate and its re-check ceremony, the Project
Structure section, and the `research.md` / `data-model.md` / `quickstart.md` / `contracts/` file
sprawl. Codefall's stance is that what gets proposed gets built, so the gating and negotiation loops
have nothing to gate — what is kept is the structure that keeps documents scannable and
machine-parseable.

**A per-row hash in the collapsed Task Plan.** Beads holds the content that was created, so a later
run compares the document against the graph itself rather than against a record of what the
document used to say.

**Recording both the spec and its vision on `Related`.** A spec already names its vision in its
own `**Vision:**` row; a design that carries both keeps a second copy of a link nothing keeps in
sync.

## Why the Task Plan collapses

Beads is the source of truth the moment the issues exist. A duplicate task list left behind in a
git-tracked document diverges from the graph, and nobody updates it. The mapping line survives
because a later run on a changed document needs to know which beads this design already produced,
so it can update them rather than duplicate the graph.
