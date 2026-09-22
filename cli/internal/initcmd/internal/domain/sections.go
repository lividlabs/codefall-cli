package domain

// The markers that delimit each of codefall's sections of AGENTS.md wherever it has been written.
// They are what makes a section replaceable: a later run finds the pair and rewrites what is
// between them, leaving everything the file says on either side alone. HTML comments, because a
// markdown reader shows nothing for them.
//
// Each section has a pair of its own rather than one pair around all four: a project set up before
// a section existed has the pairs that existed then and nothing else, and a widened pair would
// never be found in that file, so init would append a second copy of the older words for ever.
// With its own pair, a later run adds the section the project is missing beside the ones it has,
// and each is brought current on its own.
//
// The markers are a fact about codefall, which is why they live here. The words between them are
// read from `agents/sections/` in the embedded extension tree, beside the skills and the hook
// definitions, so that everything init puts into a project is under one directory; how a section
// is spliced into somebody's file is the application layer's business.
const (
	// The Codefall section: the process the other three sit inside. First in a file that has none.
	CodefallSectionBegin = "<!-- BEGIN CODEFALL PROCESS -->"
	CodefallSectionEnd   = "<!-- END CODEFALL PROCESS -->"

	// The Beads section: the project's issue tracker. It exists because init runs
	// `bd init --skip-agents`, and bd's own section describes a setup codefall does not create.
	BeadsSectionBegin = "<!-- BEGIN CODEFALL BEADS -->"
	BeadsSectionEnd   = "<!-- END CODEFALL BEADS -->"

	// The Local environment section: keeping the checkout and the local environment current (ADR-005).
	LocalSectionBegin = "<!-- BEGIN CODEFALL LOCAL -->"
	LocalSectionEnd   = "<!-- END CODEFALL LOCAL -->"

	// The Testing section: where the project's test cases live and which verbs touch them (ADR-007).
	TestingSectionBegin = "<!-- BEGIN CODEFALL TESTING -->"
	TestingSectionEnd   = "<!-- END CODEFALL TESTING -->"

	// TestingRootPlaceholder is where the declared testing root goes in the Testing section's words.
	// Unlike the three sections beside it, they name a path the project chose, so the file holds a
	// placeholder and the step fills it. The alternative was a section naming the `test` block
	// instead of the directory, which would leave every reader of AGENTS.md a lookup to perform
	// before it could act — and the reader is an agent reading one file for its rules.
	TestingRootPlaceholder = "{{TESTING_ROOT}}"
)
