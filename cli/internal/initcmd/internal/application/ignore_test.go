package application

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

var ignoreFull = filepath.Join(workingDir, settings.IgnoreName)

// ignoreResult is the last result of a run, which is the ignore step's: it runs last so that
// .ignore is the author's file to commit rather than one bd init sweeps up.
func ignoreResult(t *testing.T, files *fakeFileSystem) domain.StepResult {
	t.Helper()

	report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(
		t.Context(), beadsRequest(), nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	results := report.Results()

	return results[len(results)-1]
}

func TestIgnoreStepWritesTheFileWhenThereIsNone(t *testing.T) {
	files := settled("{}")

	result := ignoreResult(t, files)

	if result.Step != domain.IgnoreStep {
		t.Fatalf("the last step = %+v, want %+v", result.Step, domain.IgnoreStep)
	}

	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	want := settings.IgnoreComment + "\n" + settings.IgnoreEntry + "\n"
	if got := string(files.files[ignoreFull]); got != want {
		t.Errorf("%s =\n%q\nwant\n%q", settings.IgnoreName, got, want)
	}
}

// The file may be the project's own, so the step appends: overwriting a file that has drifted is the
// one thing the extension's rules never allow.
func TestIgnoreStepAppendsToAFileThatAlreadySaysSomething(t *testing.T) {
	for _, tc := range []struct {
		name     string
		existing string
	}{
		{name: "the file ends with a newline", existing: "vendor/\nnode_modules/\n"},
		{name: "the file does not", existing: "vendor/\nnode_modules/"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := settled("{}")
			files.files[ignoreFull] = []byte(tc.existing)

			result := ignoreResult(t, files)

			if result.Outcome != domain.OutcomeDone {
				t.Errorf("outcome = %v, want DONE", result.Outcome)
			}

			got := string(files.files[ignoreFull])

			if !strings.HasPrefix(got, "vendor/\nnode_modules/\n") {
				t.Errorf("%s =\n%q\nwant it to keep what was there", settings.IgnoreName, got)
			}

			if !strings.HasSuffix(got, settings.IgnoreEntry+"\n") {
				t.Errorf("%s =\n%q\nwant it to end with %q", settings.IgnoreName, got, settings.IgnoreEntry)
			}

			// The entry a file did not end with a newline before must not be joined to its last line.
			if strings.Contains(got, "node_modules/"+settings.IgnoreEntry) {
				t.Errorf("%s =\n%q\nwant the entry on a line of its own", settings.IgnoreName, got)
			}
		})
	}
}

// A rerun must not grow the file. The step compares whole lines, so an indented entry counts as
// present and a commented-out one does not — ripgrep does not read the comment either.
func TestIgnoreStepLeavesAFileThatAlreadyNamesTheDirectory(t *testing.T) {
	for _, tc := range []struct {
		name     string
		existing string
		skipped  bool
	}{
		{name: "written plainly", existing: settings.IgnoreEntry + "\n", skipped: true},
		{name: "indented", existing: "  " + settings.IgnoreEntry + "  \n", skipped: true},
		{name: "commented out", existing: "# " + settings.IgnoreEntry + "\n", skipped: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := settled("{}")
			files.files[ignoreFull] = []byte(tc.existing)

			result := ignoreResult(t, files)

			if tc.skipped {
				if result.Outcome != domain.OutcomeSkipped {
					t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
				}

				if got := string(files.files[ignoreFull]); got != tc.existing {
					t.Errorf("%s =\n%q\nwant it untouched", settings.IgnoreName, got)
				}

				return
			}

			if result.Outcome != domain.OutcomeDone {
				t.Errorf("outcome = %v, want DONE", result.Outcome)
			}

			if !strings.HasSuffix(string(files.files[ignoreFull]), settings.IgnoreEntry+"\n") {
				t.Errorf("%s =\n%q\nwant the entry added", settings.IgnoreName, files.files[ignoreFull])
			}
		})
	}
}
