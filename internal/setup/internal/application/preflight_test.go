package application

import (
	"slices"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// beadsRequest is the run every preflight test makes: the tracker and the harness are settled, and
// what the test changes is the state of the machine and of the directory.
func beadsRequest() Request {
	return Request{Dir: workingDir, Tracker: domain.TrackerBeads, Harness: domain.HarnessClaudeCode}
}

// A tool that is missing is found before anything has been done, so a run that cannot finish has
// not started either: no step began and no file was written. Every missing tool is named, because
// someone clearing prerequisites wants the whole list in one go.
func TestRunChecksItsToolsBeforeTheFirstStep(t *testing.T) {
	files := newFakeFileSystem()
	observer := &recordingObserver{}

	_, err := NewInitialize(files, newFakeCommandRunner()).Run(t.Context(), beadsRequest(), observer)

	want := "claude is not on PATH (install Claude Code: https://docs.anthropic.com/en/docs/claude-code/setup); " +
		"bd is not on PATH (brew install beads); " +
		"git is not on PATH"
	if err == nil || err.Error() != want {
		t.Errorf("Run error = %v, want %q", err, want)
	}

	if len(observer.started) != 0 || len(observer.finished) != 0 {
		t.Errorf("observer saw %d started and %d finished, want none of either",
			len(observer.started), len(observer.finished))
	}

	if len(files.files) != 0 || len(files.made) != 0 {
		t.Errorf("wrote %v and created %v, want nothing touched", files.files, files.made)
	}
}

// A missing tool is one complaint, not also every question that tool would have answered: without
// git there is no way to know whether this is a repository, so nothing is claimed about it.
func TestPreflightDoesNotComplainTwiceAboutOneMissingTool(t *testing.T) {
	runner := uninitialized()
	delete(runner.paths, gitCommand)

	_, err := NewInitialize(settled(""), runner).Run(t.Context(), beadsRequest(), nil)

	want := "git is not on PATH"
	if err == nil || err.Error() != want {
		t.Errorf("Run error = %v, want exactly %q", err, want)
	}

	if len(runner.calls) != 0 {
		t.Errorf("asked %+v, want nothing asked of a tool that is not there", runner.calls)
	}
}

// bd init run outside a repository silently runs git init first, which is a decision about the
// project that init has no business making — so a directory that is not a work tree refuses the run.
func TestPreflightRefusesADirectoryThatIsNotAGitRepository(t *testing.T) {
	runner := uninitialized()
	runner.runs[gitWorkTree] = CommandResult{
		ExitCode: 128, Stderr: "fatal: not a git repository (or any of the parent directories): .git\n",
	}

	files := newFakeFileSystem()

	_, err := NewInitialize(files, runner).Run(t.Context(), beadsRequest(), nil)

	want := "not a git repository (bd init would create one; run git init first)"
	if err == nil || err.Error() != want {
		t.Errorf("Run error = %v, want exactly %q", err, want)
	}

	// Nothing is asked about the index of a repository that is not there.
	if got := commandsAsked(runner); slices.Contains(got, gitStaged) {
		t.Errorf("asked %q, want no question about the index", got)
	}

	if len(files.files) != 0 || len(files.made) != 0 {
		t.Errorf("wrote %v and created %v, want nothing touched", files.files, files.made)
	}
}

// bd init commits with no pathspec, so whatever is staged is swept into a commit that says it
// initialised Beads. A run that would do that refuses to start.
func TestPreflightRefusesAStagedIndexWhenBeadsIsNotInitializedYet(t *testing.T) {
	runner := uninitialized()
	runner.runs[gitStaged] = CommandResult{ExitCode: 1}

	files := newFakeFileSystem()

	_, err := NewInitialize(files, runner).Run(t.Context(), beadsRequest(), nil)

	want := "the git index has staged changes; bd init commits the whole index, so commit or unstage them first"
	if err == nil || err.Error() != want {
		t.Errorf("Run error = %v, want exactly %q", err, want)
	}

	if len(files.files) != 0 || len(files.made) != 0 {
		t.Errorf("wrote %v and created %v, want nothing touched", files.files, files.made)
	}
}

// Beads that is already initialised means bd init will not run, and then what is staged is nobody's
// business but the person who staged it. The question is not even asked.
func TestPreflightIgnoresTheIndexWhenBeadsIsAlreadyInitialized(t *testing.T) {
	runner := toolsInstalled()
	runner.runs[gitStaged] = CommandResult{ExitCode: 1}

	if _, err := NewInitialize(settled(""), runner).Run(t.Context(), beadsRequest(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := commandsAsked(runner); slices.Contains(got, gitStaged) {
		t.Errorf("asked %q, want no question about the index", got)
	}
}

// git failing for a reason of its own is neither a clean index nor a dirty one, so it is reported as
// what it is, with what git said.
func TestPreflightReportsGitFailingForItsOwnReasons(t *testing.T) {
	runner := uninitialized()
	runner.runs[gitStaged] = CommandResult{ExitCode: 128, Stderr: "fatal: bad object HEAD\nsee git-diff(1)\n"}

	_, err := NewInitialize(newFakeFileSystem(), runner).Run(t.Context(), beadsRequest(), nil)

	want := "git diff --cached --quiet exited 128: fatal: bad object HEAD"
	if err == nil || err.Error() != want {
		t.Errorf("Run error = %v, want exactly %q", err, want)
	}
}

// commandsAsked is every command the runner was given, probes included — which is what a test about
// preflight is looking at.
func commandsAsked(runner *fakeCommandRunner) []string {
	got := make([]string, 0, len(runner.calls))
	for _, call := range runner.calls {
		got = append(got, call.command)
	}

	return got
}
