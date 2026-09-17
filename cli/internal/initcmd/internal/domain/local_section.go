package domain

import _ "embed"

// The markers that delimit codefall's local-environment section, the second section it writes to
// AGENTS.md. A pair of its own, so a project set up before the section existed gains it beside the
// Beads section it already has, and each can be brought current on its own.
const (
	LocalSectionBegin = "<!-- BEGIN CODEFALL LOCAL -->"
	LocalSectionEnd   = "<!-- END CODEFALL LOCAL -->"
)

// LocalSection is what codefall tells an agent about keeping the local environment current
// (ADR-005): run refresh before starting work, keep the start and update scripts current at the
// point a change makes them stale, and run equip when nothing is declared. It is a fact about
// codefall — the words it says — which is why it lives in the domain, beside the file that holds
// them.
//
//go:embed local_section.md
var LocalSection string
