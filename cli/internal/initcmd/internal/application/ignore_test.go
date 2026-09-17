package application

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

var (
	ignoreFull    = filepath.Join(workingDir, settings.IgnoreName)
	gitIgnoreFull = filepath.Join(workingDir, settings.GitIgnoreName)
)

// stampIgnored is a .gitignore that already names the stamp, for the tests about the other file.
func stampIgnored(files *fakeFileSystem) *fakeFileSystem {
	files.files[gitIgnoreFull] = []byte(settings.RefreshStamp + "\n")

	return files
}

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

func TestIgnoreStepWritesTheFilesWhenThereAreNone(t *testing.T) {
	files := settled("{}")

	result := ignoreResult(t, files)

	if result.Step != domain.IgnoreStep {
		t.Fatalf("the last step = %+v, want %+v", result.Step, domain.IgnoreStep)
	}

	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	if want := "wrote .ignore and wrote .gitignore"; result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	want := settings.IgnoreComment + "\n" + settings.IgnoreEntry + "\n"
	if got := string(files.files[ignoreFull]); got != want {
		t.Errorf("%s =\n%q\nwant\n%q", settings.IgnoreName, got, want)
	}

	want = settings.GitIgnoreComment + "\n" + settings.RefreshStamp + "\n"
	if got := string(files.files[gitIgnoreFull]); got != want {
		t.Errorf("%s =\n%q\nwant\n%q", settings.GitIgnoreName, got, want)
	}
}

// A project's .gitignore is nearly always its own already, so the stamp entry is appended to it,
// and the .ignore file that is already right is left alone and left out of the sentence.
func TestIgnoreStepAddsTheStampToAGitignoreThatIsAlreadyThere(t *testing.T) {
	files := settled("{}")
	files.files[ignoreFull] = []byte(settings.IgnoreEntry + "\n")
	files.files[gitIgnoreFull] = []byte("node_modules/\ndist/\n")

	result := ignoreResult(t, files)

	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	if want := "added " + settings.RefreshStamp + " to .gitignore"; result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	want := "node_modules/\ndist/\n\n" + settings.GitIgnoreComment + "\n" + settings.RefreshStamp + "\n"
	if got := string(files.files[gitIgnoreFull]); got != want {
		t.Errorf("%s =\n%q\nwant\n%q", settings.GitIgnoreName, got, want)
	}

	if got := string(files.files[ignoreFull]); got != settings.IgnoreEntry+"\n" {
		t.Errorf("%s =\n%q\nwant it untouched", settings.IgnoreName, got)
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
			files := stampIgnored(settled("{}"))
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
			files := stampIgnored(settled("{}"))
			files.files[ignoreFull] = []byte(tc.existing)

			result := ignoreResult(t, files)

			if tc.skipped {
				if result.Outcome != domain.OutcomeSkipped {
					t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
				}

				want := ".ignore already names " + settings.IgnoreEntry +
					" and .gitignore already names " + settings.RefreshStamp
				if result.Detail != want {
					t.Errorf("detail = %q, want %q", result.Detail, want)
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
