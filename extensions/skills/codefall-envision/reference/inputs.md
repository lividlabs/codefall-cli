# What the user already has

A vision may arrive as more than a conversation: notes, a transcript, a pitch document, a
whiteboard photo, sketches, screenshots, a folder of design-tool exports, a link to a design tool,
or several of those at once. **All of it is fine, and none of it is discarded.** Read at step 2 when
the user points at material.

## Where it goes

Two destinations, decided by what the material is rather than by how it arrived.

| Material | Where |
| --- | --- |
| A picture of a surface someone will build — exports, screenshots of a design tool, a mockup set | `docs/mockups/<slug>/`, following `../../../../.codefall/shared/import-mockup.md` |
| Everything else — notes, transcripts, pitch documents, sketches of the problem, a diagram of how something works today | `docs/visions/sources/VISION-NNN-<original-filename>` |

Follow the import procedure as written, including its rules about never editing or interpreting
what the user brought.

**When it is unclear which a thing is, ask.** A whiteboard photo of boxes and arrows is thinking; a
photo of a drawn screen is a mockup; plenty of images are either. One question settles it. A single
run can use both destinations.

**The source is always saved**, without exception — including when a document fits so well that it
is only being renamed. It keeps its original name and extension, and the vision's `Related` line
points at wherever it landed.

**Never edit the user's original in place, and never delete it.**

## Documents that could be the vision

When what they brought is a written document, one of two things happens:

- **The document is a strict superset of the template** — it has Problem and Proposed shape and
  everything else it needs, plus whatever more its author wanted. Adopt it: copy it to
  `docs/visions/VISION-NNN-slug.md` with its body unchanged, add the header, and leave the
  structure alone.
- **Anything less** — the common case, including a document that is nearly right — write a new
  vision from it. Take the user's sentences over your own wherever they work, and put what the
  source does not answer under **Open questions**.
