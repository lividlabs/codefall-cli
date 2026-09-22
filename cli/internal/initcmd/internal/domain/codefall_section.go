package domain

import _ "embed"

// The markers that delimit codefall's process section, the fourth section it writes to AGENTS.md
// and the first in a file that has none. A pair of its own, for the same reason the other three
// have one: a project set up before the section existed gains it beside the sections it already
// has, and each is brought current on its own.
const (
	CodefallSectionBegin = "<!-- BEGIN CODEFALL PROCESS -->"
	CodefallSectionEnd   = "<!-- END CODEFALL PROCESS -->"
)

// CodefallSection is what codefall tells an agent about the process the other three sections sit
// inside: the chain of verbs, the verbs beside it, that every verb is invoked deliberately and
// applies only what the user takes, that a human performs every merge, and who is authoritative
// for what. The detail is `.codefall/shared/workflow.md`, which the extension step installs; the
// section names it rather than restating it. It is a fact about codefall — the words it says —
// which is why it lives in the domain, beside the file that holds them.
//
//go:embed codefall_section.md
var CodefallSection string
