package application

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

var (
	agentsFull       = filepath.Join(workingDir, "AGENTS.md")
	claudeMemoryFull = filepath.Join(workingDir, "CLAUDE.md")
)

// section is what a file that has been through the step says, without the newline that ends it.
var section = strings.TrimSuffix(domain.BeadsSection, "\n")

// agentsResult is what the fifth step did in a run that got that far.
func agentsResult(t *testing.T, report domain.Report) domain.StepResult {
	t.Helper()

	results := report.Results()
	if len(results) != 5 {
		t.Fatalf("Results() = %+v, want a result for each of the five steps", results)
	}

	return results[4]
}

// run is the whole init run, in a project whose settings and harness file are already settled, with
// whatever markdown the test wants beside them.
func run(t *testing.T, markdown map[string][]byte) (*fakeFileSystem, domain.StepResult) {
	t.Helper()

	files := settled("{}")
	for path, body := range markdown {
		files.files[path] = body
	}

	report, err := NewInitialize(files, toolsInstalled(), newFakePluginFetcher()).Run(t.Context(), beadsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	return files, agentsResult(t, report)
}

// A project with no AGENTS.md gets one that is the section and nothing else, and the CLAUDE.md that
// points at it.
func TestAgentsStepCreatesTheFilesThatAreNotThere(t *testing.T) {
	files, result := run(t, nil)

	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	want := "created AGENTS.md with the Beads section and created CLAUDE.md"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if got := string(files.files[agentsFull]); got != section+"\n" {
		t.Errorf("AGENTS.md =\n%s\nwant\n%s", got, section+"\n")
	}

	want = "See [AGENTS.md](AGENTS.md) — the rules for this repo live there, and there only.\n"
	if got := string(files.files[claudeMemoryFull]); got != want {
		t.Errorf("CLAUDE.md = %q, want %q", got, want)
	}
}

// A file that already says something keeps every word of it and gains the section at the end, with
// one blank line in between however the file happened to end.
func TestAgentsStepAppendsToAFileThatAlreadySaysSomething(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before string
	}{
		{name: "the file ends with a newline", before: "# Project\n\nWhat it is.\n"},
		{name: "the file does not", before: "# Project\n\nWhat it is."},
		{name: "the file ends with several newlines", before: "# Project\n\nWhat it is.\n\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, result := run(t, map[string][]byte{agentsFull: []byte(tc.before)})

			if result.Outcome != domain.OutcomeDone {
				t.Errorf("outcome = %v, want DONE", result.Outcome)
			}

			if want := "added the Beads section to AGENTS.md and created CLAUDE.md"; result.Detail != want {
				t.Errorf("detail = %q, want %q", result.Detail, want)
			}

			want := "# Project\n\nWhat it is.\n\n" + section + "\n"
			if got := string(files.files[agentsFull]); got != want {
				t.Errorf("AGENTS.md =\n%q\nwant\n%q", got, want)
			}
		})
	}
}

// A section that is already there is rewritten between its markers, and what the file says on either
// side comes back byte for byte.
func TestAgentsStepReplacesTheSectionInPlace(t *testing.T) {
	const before = "# Project\n\nWhat it is.\n\n" +
		domain.BeadsSectionBegin + "\n## Beads\n\nSomething older.\n" + domain.BeadsSectionEnd +
		"\n\n## Afterwards\n\nMore of the project's own words.\n"

	files, result := run(t, map[string][]byte{agentsFull: []byte(before)})

	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	if want := "updated the Beads section in AGENTS.md and created CLAUDE.md"; result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	want := "# Project\n\nWhat it is.\n\n" + section +
		"\n\n## Afterwards\n\nMore of the project's own words.\n"

	if got := string(files.files[agentsFull]); got != want {
		t.Errorf("AGENTS.md =\n%q\nwant\n%q", got, want)
	}
}

// A second run has nothing to do: the section is what codefall would have written, and the file is
// not touched. This is what makes init safe to run again.
func TestAgentsStepSkipsASectionThatIsCurrent(t *testing.T) {
	before := "# Project\n\n" + section + "\n"

	files, result := run(t, map[string][]byte{
		agentsFull:       []byte(before),
		claudeMemoryFull: []byte("See [AGENTS.md](AGENTS.md).\n"),
	})

	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	if want := "the Beads section in AGENTS.md is current"; result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if got := string(files.files[agentsFull]); got != before {
		t.Errorf("AGENTS.md = %q, want it untouched", got)
	}
}

// The section is current and CLAUDE.md is not there: the step has one thing to report, and it is
// that one thing rather than a sentence about AGENTS.md that would not be true.
func TestAgentsStepReportsTheOneThingItDid(t *testing.T) {
	_, result := run(t, map[string][]byte{agentsFull: []byte(section + "\n")})

	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	if want := "created CLAUDE.md"; result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}
}

// A CLAUDE.md the project already has is left alone, whatever it says: it is somebody's file, and a
// pointer written over it would throw the rules away rather than point at them.
func TestAgentsStepLeavesAClaudeFileThatIsAlreadyThere(t *testing.T) {
	for _, before := range []string{"# House rules\n\nDo the thing.\n", ""} {
		files, _ := run(t, map[string][]byte{claudeMemoryFull: []byte(before)})

		if got := string(files.files[claudeMemoryFull]); got != before {
			t.Errorf("CLAUDE.md = %q, want it untouched", got)
		}
	}
}

// Only Claude Code reads CLAUDE.md, so only Claude Code gets one. The harness the section itself is
// written for is every harness: AGENTS.md is the file they all read.
func TestAgentsStepWritesClaudeMdOnlyForClaudeCode(t *testing.T) {
	files := settled("{}")

	request := beadsRequest()
	request.Harness = "aider"

	result, err := NewInitialize(files, toolsInstalled(), newFakePluginFetcher()).agents(t.Context(), request)
	if err != nil {
		t.Fatalf("agents: %v", err)
	}

	if want := "created AGENTS.md with the Beads section"; result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if _, written := files.files[claudeMemoryFull]; written {
		t.Errorf("CLAUDE.md = %q, want no file for a harness that does not read one",
			files.files[claudeMemoryFull])
	}
}

// An opening marker with no closing one is a file the step cannot edit without guessing where
// codefall's words stop, so it stops the run instead and says which marker it wanted.
func TestAgentsStepRefusesAnUnclosedSection(t *testing.T) {
	const before = "# Project\n\n" + domain.BeadsSectionBegin + "\n## Beads\n\nHalf a section.\n"

	files := settled("{}")
	files.files[agentsFull] = []byte(before)

	_, err := NewInitialize(files, toolsInstalled(), newFakePluginFetcher()).agents(t.Context(), beadsRequest())
	if err == nil || !strings.Contains(err.Error(), "AGENTS.md has "+domain.BeadsSectionBegin+
		" with no "+domain.BeadsSectionEnd+" after it") {
		t.Errorf("agents error = %v, want it to name the file and the missing marker", err)
	}

	if got := string(files.files[agentsFull]); got != before {
		t.Errorf("AGENTS.md = %q, want it untouched", got)
	}
}

// A file that cannot be read or written is the step's error, named so the reader knows which file.
func TestAgentsStepReportsAFileItCannotUse(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*fakeFileSystem)
		want  string
	}{
		{
			name:  "AGENTS.md cannot be read",
			setup: func(f *fakeFileSystem) { f.errs[agentsFull] = errors.New("permission denied") },
			want:  "read AGENTS.md",
		},
		{
			name:  "AGENTS.md cannot be written",
			setup: func(f *fakeFileSystem) { f.errs["write "+agentsFull] = errors.New("read-only file system") },
			want:  "write AGENTS.md",
		},
		{
			name:  "CLAUDE.md cannot be read",
			setup: func(f *fakeFileSystem) { f.errs[claudeMemoryFull] = errors.New("permission denied") },
			want:  "read CLAUDE.md",
		},
		{
			name: "CLAUDE.md cannot be written",
			setup: func(f *fakeFileSystem) {
				f.errs["write "+claudeMemoryFull] = errors.New("read-only file system")
			},
			want: "write CLAUDE.md",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := settled("{}")
			tc.setup(files)

			_, err := NewInitialize(files, toolsInstalled(), newFakePluginFetcher()).agents(t.Context(), beadsRequest())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("agents error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

// The run reaches the section last, after bd init has made whatever commit it was going to make:
// codefall's edit to AGENTS.md is the author's to commit, not bd's.
func TestAgentsStepRunsAfterBeads(t *testing.T) {
	observer := &recordingObserver{}

	if _, err := NewInitialize(settled("{}"), toolsInstalled(), newFakePluginFetcher()).Run(
		t.Context(), beadsRequest(), observer,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	last := observer.started[len(observer.started)-1]
	if last != domain.AgentsStep {
		t.Errorf("the last step started = %+v, want %+v", last, domain.AgentsStep)
	}
}
