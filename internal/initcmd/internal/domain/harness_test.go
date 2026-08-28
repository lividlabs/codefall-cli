package domain

import (
	"slices"
	"testing"
)

func TestHarnessesAreSortedAndCopied(t *testing.T) {
	if got, want := Harnesses(), []string{HarnessClaudeCode}; !slices.Equal(got, want) {
		t.Errorf("Harnesses() = %q, want %q", got, want)
	}

	// The caller gets a copy: mutating it must not change what the next caller sees.
	Harnesses()[0] = "mutated"

	if got := Harnesses()[0]; got != HarnessClaudeCode {
		t.Errorf("Harnesses()[0] after a caller mutated its copy = %q, want %q", got, HarnessClaudeCode)
	}
}

func TestParseHarness(t *testing.T) {
	if got, err := ParseHarness(HarnessClaudeCode); err != nil || got != HarnessClaudeCode {
		t.Errorf("ParseHarness(%q) = %q, %v, want %q, nil", HarnessClaudeCode, got, err, HarnessClaudeCode)
	}

	err := errorFrom(ParseHarness("codex"))

	want := `harness "codex" is not supported yet (supported: claude-code)`
	if err == nil || err.Error() != want {
		t.Errorf("ParseHarness(%q) error = %v, want %q", "codex", err, want)
	}
}

// errorFrom drops the value of a (T, error) pair so a table can assert on the error alone.
func errorFrom(_ string, err error) error {
	return err
}
