package domain

import (
	"slices"
	"testing"
)

func TestHarnessesAreSortedAndCopied(t *testing.T) {
	want := []string{
		HarnessAntigravity, HarnessClaudeCode, HarnessCodex, HarnessMuse, HarnessOpenCode,
	}

	if got := Harnesses(); !slices.Equal(got, want) {
		t.Errorf("Harnesses() = %q, want %q", got, want)
	}

	// The caller gets a copy: mutating it must not change what the next caller sees.
	Harnesses()[0] = "mutated"

	if got := Harnesses()[0]; got != HarnessAntigravity {
		t.Errorf("Harnesses()[0] after a caller mutated its copy = %q, want %q", got, HarnessAntigravity)
	}
}

func TestParseHarness(t *testing.T) {
	if got, err := ParseHarness(HarnessCodex); err != nil || got != HarnessCodex {
		t.Errorf("ParseHarness(%q) = %q, %v, want %q, nil", HarnessCodex, got, err, HarnessCodex)
	}

	err := errorFrom(ParseHarness("cursor"))

	want := `harness "cursor" is not supported yet (supported: antigravity, claude-code, codex, muse, opencode)`
	if err == nil || err.Error() != want {
		t.Errorf("ParseHarness(%q) error = %v, want %q", "cursor", err, want)
	}
}

// errorFrom drops the value of a (T, error) pair so a table can assert on the error alone.
func errorFrom(_ string, err error) error {
	return err
}
