package domain

import (
	_ "embed"
	"strings"
)

// The markers that delimit codefall's testing section, the third section it writes to AGENTS.md. A
// pair of its own, for the same reason the local-environment section has one: a project set up
// before the section existed gains it beside the sections it already has, and each is brought
// current on its own.
const (
	TestingSectionBegin = "<!-- BEGIN CODEFALL TESTING -->"
	TestingSectionEnd   = "<!-- END CODEFALL TESTING -->"
	// TestingRootPlaceholder is where the declared testing root goes in the words below.
	TestingRootPlaceholder = "{{TESTING_ROOT}}"
)

//go:embed testing_section.md
var testingSection string

// TestingSection is what codefall tells an agent about the project's test cases (ADR-007): where
// they live, that they are written through implement and run through test, and that equip is what
// sets the harness up.
//
// Unlike the two sections beside it, the words name a path the project chose, so the file holds a
// placeholder and this fills it. The alternative was a section naming the `test` block instead of
// the directory, which would leave every reader of AGENTS.md a lookup to perform before it could
// act — and the reader is an agent reading one file for its rules.
func TestingSection(root string) string {
	return strings.ReplaceAll(testingSection, TestingRootPlaceholder, root)
}
