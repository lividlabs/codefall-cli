package domain

import (
	"strings"
	"testing"
)

// The same rule the Beads section is held to: one pair of markers, the first line and the last, so
// the splice replaces exactly codefall's words and nothing of the project's.
func TestTestingSectionIsMarkedAtBothEnds(t *testing.T) {
	section := TestingSection("testing")

	if !strings.HasPrefix(section, TestingSectionBegin+"\n") {
		t.Errorf("the section starts %q, want it to open with %q", firstLine(section), TestingSectionBegin)
	}

	if !strings.HasSuffix(section, "\n"+TestingSectionEnd+"\n") {
		t.Errorf("the section ends %q, want it to close with %q and a newline",
			lastLines(section), TestingSectionEnd)
	}

	if got := strings.Count(section, TestingSectionBegin); got != 1 {
		t.Errorf("the section has %d opening markers, want exactly 1", got)
	}

	if got := strings.Count(section, TestingSectionEnd); got != 1 {
		t.Errorf("the section has %d closing markers, want exactly 1", got)
	}
}

// The root is the project's to choose, so every mention of it in the words is the declared one and
// no placeholder survives into somebody's file.
func TestTestingSectionNamesTheDeclaredRoot(t *testing.T) {
	section := TestingSection("packages/web/e2e")

	if strings.Contains(section, TestingRootPlaceholder) {
		t.Errorf("the section still holds %q:\n%s", TestingRootPlaceholder, section)
	}

	if !strings.Contains(section, "packages/web/e2e/") {
		t.Errorf("the section does not name the declared root:\n%s", section)
	}

	// The section names the three verbs a reader needs from it, and nothing else names them for it.
	for _, verb := range []string{"/codefall-implement", "/codefall-test", "/codefall-equip"} {
		if !strings.Contains(section, verb) {
			t.Errorf("the section does not name %s:\n%s", verb, section)
		}
	}
}

// Both skeletons are whole documents: a heading to open, and for the README the empty case index a
// project fills in as it writes cases.
func TestTheSkeletonsAreCompleteDocuments(t *testing.T) {
	if !strings.HasPrefix(TestingAgents, "# ") || !strings.HasSuffix(TestingAgents, "\n") {
		t.Errorf("the AGENTS.md skeleton is not a whole document:\n%s", TestingAgents)
	}

	for _, heading := range []string{
		"## Runners", "## Setup and state-forcing commands", "## Real side effects",
		"## Environment notes",
	} {
		if !strings.Contains(TestingAgents, heading) {
			t.Errorf("the AGENTS.md skeleton has no %q section:\n%s", heading, TestingAgents)
		}
	}

	if !strings.HasPrefix(TestingReadme, "# ") || !strings.HasSuffix(TestingReadme, "\n") {
		t.Errorf("the README.md skeleton is not a whole document:\n%s", TestingReadme)
	}

	if !strings.Contains(TestingReadme, "| Case ID | Modalities | Variants | Notes |") {
		t.Errorf("the README.md skeleton has no case index:\n%s", TestingReadme)
	}
}
