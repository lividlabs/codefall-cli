package harness

import (
	"slices"
	"testing"
)

func TestAllIsSortedAndCopied(t *testing.T) {
	want := []string{Antigravity, ClaudeCode, Codex, Muse, OpenCode}

	if got := All(); !slices.Equal(got, want) {
		t.Errorf("All() = %q, want %q", got, want)
	}

	// The caller gets a copy: mutating it must not change what the next caller sees.
	All()[0] = "mutated"

	if got := All()[0]; got != Antigravity {
		t.Errorf("All()[0] after a caller mutated its copy = %q, want %q", got, Antigravity)
	}
}

func TestParse(t *testing.T) {
	if got, err := Parse(Codex); err != nil || got != Codex {
		t.Errorf("Parse(%q) = %q, %v, want %q, nil", Codex, got, err, Codex)
	}

	err := errorFrom(Parse("cursor"))

	want := `harness "cursor" is not supported yet (supported: antigravity, claude-code, codex, muse, opencode)`
	if err == nil || err.Error() != want {
		t.Errorf("Parse(%q) error = %v, want %q", "cursor", err, want)
	}
}

// errorFrom drops the value of a (T, error) pair so a table can assert on the error alone.
func errorFrom(_ string, err error) error {
	return err
}

// Every harness codefall can set up has somewhere to install, which is what makes the table the
// roster: a row that went missing would take the name out of All and Parse as well, rather than
// leaving a name the steps could accept and then have nowhere to write.
func TestEveryHarnessHasSomewhereToInstall(t *testing.T) {
	for _, name := range All() {
		dir, known := SkillsDir(name).Get()

		switch {
		case !known:
			t.Errorf("SkillsDir(%q) = None, want a directory", name)
		case dir == "":
			t.Errorf("SkillsDir(%q) = %q, want a directory", name, dir)
		}
	}
}

func TestSkillsDirNamesTheConventionEachHarnessFollows(t *testing.T) {
	for _, tc := range []struct {
		harness string
		want    string
	}{
		{ClaudeCode, ".claude"},
		{Antigravity, ".agents"},
		{Codex, ".agents"},
		{Muse, ".agents"},
		{OpenCode, ".agents"},
	} {
		t.Run(tc.harness, func(t *testing.T) {
			if got := SkillsDir(tc.harness).OrEmpty(); got != tc.want {
				t.Errorf("SkillsDir(%q) = %q, want %q", tc.harness, got, tc.want)
			}
		})
	}
}

// A harness codefall cannot set up has no directory to report, and the absence is the answer rather
// than an empty string that reads like one.
func TestSkillsDirIsAbsentForAHarnessCodefallCannotSetUp(t *testing.T) {
	if got := SkillsDir("aider"); got.IsPresent() {
		t.Errorf("SkillsDir(%q) = %v, want None", "aider", got)
	}
}
