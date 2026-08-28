package domain

import (
	"strings"
	"testing"
)

// The section is spliced into somebody's file by its markers, so the markers have to be exactly
// where the splice expects them: one pair, the first line and the last, with the newline that ends
// the file after the closing one. A stray extra marker in the body would make a later run replace
// the wrong span.
func TestBeadsSectionIsMarkedAtBothEnds(t *testing.T) {
	if !strings.HasPrefix(BeadsSection, BeadsSectionBegin+"\n") {
		t.Errorf("the section starts %q, want it to open with %q",
			firstLine(BeadsSection), BeadsSectionBegin)
	}

	if !strings.HasSuffix(BeadsSection, "\n"+BeadsSectionEnd+"\n") {
		t.Errorf("the section ends %q, want it to close with %q and a newline",
			lastLines(BeadsSection), BeadsSectionEnd)
	}

	if got := strings.Count(BeadsSection, BeadsSectionBegin); got != 1 {
		t.Errorf("the section has %d opening markers, want exactly 1", got)
	}

	if got := strings.Count(BeadsSection, BeadsSectionEnd); got != 1 {
		t.Errorf("the section has %d closing markers, want exactly 1", got)
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")

	return line
}

func lastLines(s string) string {
	if len(s) > 60 {
		return s[len(s)-60:]
	}

	return s
}
