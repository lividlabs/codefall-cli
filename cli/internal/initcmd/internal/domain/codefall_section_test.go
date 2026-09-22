package domain

import (
	"strings"
	"testing"
)

// The same rule the Beads section is held to: one pair of markers, the first line and the last, so
// the splice replaces exactly codefall's words and nothing of the project's.
func TestCodefallSectionIsMarkedAtBothEnds(t *testing.T) {
	if !strings.HasPrefix(CodefallSection, CodefallSectionBegin+"\n") {
		t.Errorf("the section starts %q, want it to open with %q",
			firstLine(CodefallSection), CodefallSectionBegin)
	}

	if !strings.HasSuffix(CodefallSection, "\n"+CodefallSectionEnd+"\n") {
		t.Errorf("the section ends %q, want it to close with %q and a newline",
			lastLines(CodefallSection), CodefallSectionEnd)
	}

	if got := strings.Count(CodefallSection, CodefallSectionBegin); got != 1 {
		t.Errorf("the section has %d opening markers, want exactly 1", got)
	}

	if got := strings.Count(CodefallSection, CodefallSectionEnd); got != 1 {
		t.Errorf("the section has %d closing markers, want exactly 1", got)
	}
}

// The section is a frame, not the reference: it names the file that holds the detail, so an agent
// that needs the chain table knows where it is without the section carrying it.
func TestCodefallSectionNamesTheInstalledDetail(t *testing.T) {
	if !strings.Contains(CodefallSection, ".codefall/shared/workflow.md") {
		t.Errorf("the section does not name the installed workflow file:\n%s", CodefallSection)
	}
}
