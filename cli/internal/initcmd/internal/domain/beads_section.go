package domain

import _ "embed"

// The markers that delimit codefall's section wherever it has been written. They are what makes the
// section replaceable: a later run finds the pair and rewrites what is between them, leaving
// everything the file says on either side alone. HTML comments, because a markdown reader shows
// nothing for them.
const (
	BeadsSectionBegin = "<!-- BEGIN CODEFALL BEADS -->"
	BeadsSectionEnd   = "<!-- END CODEFALL BEADS -->"
)

// BeadsSection is what codefall tells an agent about the project's issue tracker, marker lines
// included. It is a fact about codefall — the words it says — which is why it lives in the domain,
// beside the file that holds them; where it is written, and how it is spliced into a file that
// already exists, is the application layer's business.
//
// It exists because init runs `bd init --skip-agents`: bd would otherwise append a section of its
// own to AGENTS.md and CLAUDE.md, and that text does not match the setup codefall creates. bd's
// version is the reference this one started from:
// https://github.com/gastownhall/beads/blob/6c124203e771433a3550c348771a5b5e27fd3c21/internal/templates/agents/defaults/beads-section-minimal.md
//
//go:embed beads_section.md
var BeadsSection string
