package application

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

const workingDir = "/work"

var (
	codefallDir  = filepath.Join(workingDir, ".codefall")
	settingsFull = filepath.Join(codefallDir, "settings.json")
	claudeFull   = filepath.Join(workingDir, ".claude", "settings.json")
)

// The two commands the plugin step runs, keyed the way the fake runner keys them.
const (
	marketplaceAdd = "claude plugin marketplace add lividlabs/codefall-plugin --scope project"
	pluginInstall  = "claude plugin install codefall@codefall --scope project -y"
)

// --- fakes -------------------------------------------------------------------------------------

type fakeFileSystem struct {
	dirs  map[string]bool
	files map[string][]byte
	errs  map[string]error
	made  []string
}

func newFakeFileSystem() *fakeFileSystem {
	return &fakeFileSystem{
		dirs:  map[string]bool{},
		files: map[string][]byte{},
		errs:  map[string]error{},
	}
}

func (f *fakeFileSystem) DirExists(path string) (bool, error) {
	if err, ok := f.errs[path]; ok {
		return false, err
	}

	return f.dirs[path], nil
}

func (f *fakeFileSystem) ReadFile(path string) ([]byte, error) {
	if err, ok := f.errs[path]; ok {
		return nil, err
	}

	data, ok := f.files[path]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}

	return data, nil
}

func (f *fakeFileSystem) MkdirAll(path string) error {
	if err, ok := f.errs["mkdir "+path]; ok {
		return err
	}

	f.made = append(f.made, path)
	f.dirs[path] = true

	return nil
}

func (f *fakeFileSystem) WriteFile(path string, data []byte) error {
	if err, ok := f.errs["write "+path]; ok {
		return err
	}

	f.files[path] = data

	return nil
}

type runCall struct {
	dir     string
	command string
}

type fakeCommandRunner struct {
	paths map[string]string
	runs  map[string]CommandResult
	errs  map[string]error
	calls []runCall
}

func newFakeCommandRunner() *fakeCommandRunner {
	return &fakeCommandRunner{
		paths: map[string]string{},
		runs:  map[string]CommandResult{},
		errs:  map[string]error{},
	}
}

func (r *fakeCommandRunner) LookPath(name string) mo.Option[string] {
	if path, ok := r.paths[name]; ok {
		return mo.Some(path)
	}

	return mo.None[string]()
}

func (r *fakeCommandRunner) Run(
	_ context.Context, dir, name string, args ...string,
) (CommandResult, error) {
	command := strings.TrimSpace(name + " " + strings.Join(args, " "))
	r.calls = append(r.calls, runCall{dir: dir, command: command})

	if err, ok := r.errs[command]; ok {
		return CommandResult{}, err
	}

	return r.runs[command], nil
}

// claudeInstalled is the runner a run with the claude-code harness needs: the harness CLI is on
// PATH, and every command it is given succeeds, because the zero CommandResult exited 0.
func claudeInstalled() *fakeCommandRunner {
	runner := newFakeCommandRunner()
	runner.paths["claude"] = "/opt/homebrew/bin/claude"

	return runner
}

// recordingObserver is what a spinner does in production, without the terminal: it remembers what it
// was told and in which order.
type recordingObserver struct {
	started  []domain.Step
	finished []domain.StepResult
}

func (o *recordingObserver) StepStarted(step domain.Step) {
	o.started = append(o.started, step)
}

func (o *recordingObserver) StepFinished(result domain.StepResult) {
	o.finished = append(o.finished, result)
}

// --- the run -----------------------------------------------------------------------------------

func TestRunWritesSettingsAndReportsWhatItWrote(t *testing.T) {
	files := newFakeFileSystem()
	observer := &recordingObserver{}

	report, err := NewInitialize(files, claudeInstalled()).Run(
		t.Context(),
		Request{
			Dir:           workingDir,
			Tracker:       domain.TrackerGitHub,
			GitHubRepo:    mo.Some("lividlabs/codefall-cli"),
			GitHubProject: mo.Some(3),
			Harness:       domain.HarnessClaudeCode,
		},
		observer,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	results := report.Results()
	if len(results) != 2 {
		t.Fatalf("Results() = %+v, want the settings and the plugin result", results)
	}

	if results[0].Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", results[0].Outcome)
	}

	want := "wrote .codefall/settings.json (tracker: github, repo: lividlabs/codefall-cli, project: 3)"
	if results[0].Detail != want {
		t.Errorf("detail = %q, want %q", results[0].Detail, want)
	}

	if len(files.made) != 1 || files.made[0] != codefallDir {
		t.Errorf("created %q, want %q", files.made, []string{codefallDir})
	}

	// The observer sees every step start and finish, in order, and what it sees on finishing is the
	// result the report carries.
	if len(observer.started) != 2 ||
		observer.started[0] != domain.SettingsStep ||
		observer.started[1] != domain.PluginStep {
		t.Errorf("started = %+v, want the settings step then the plugin step", observer.started)
	}

	if len(observer.finished) != 2 || observer.finished[0] != results[0] || observer.finished[1] != results[1] {
		t.Errorf("finished = %+v, want %+v", observer.finished, results)
	}
}

func TestRunEncodesGitHubSettings(t *testing.T) {
	files := newFakeFileSystem()

	if _, err := NewInitialize(files, claudeInstalled()).Run(
		t.Context(),
		Request{
			Dir:           workingDir,
			Tracker:       domain.TrackerGitHub,
			GitHubRepo:    mo.Some("owner/name"),
			GitHubProject: mo.Some(3),
			Harness:       domain.HarnessClaudeCode,
		},
		nil,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := `{
  "$schema": "` + domain.SettingsSchemaID + `",
  "version": 1,
  "tracker": "github",
  "github": {
    "repo": "owner/name",
    "project": 3
  }
}
`

	if got := string(files.files[settingsFull]); got != want {
		t.Errorf("settings.json =\n%s\nwant\n%s", got, want)
	}
}

// An absent project is omitted rather than written as null, and the empty beads block is written
// rather than omitted — the schema requires the block that the tracker selects.
func TestRunEncodesTheOptionalFieldsTheWayTheSchemaExpects(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request Request
		want    string
	}{
		{
			name: "github without a project",
			request: Request{
				Dir:        workingDir,
				Tracker:    domain.TrackerGitHub,
				GitHubRepo: mo.Some("owner/name"),
				Harness:    domain.HarnessClaudeCode,
			},
			want: `  "tracker": "github",
  "github": {
    "repo": "owner/name"
  }
}
`,
		},
		{
			name:    "beads",
			request: Request{Dir: workingDir, Tracker: domain.TrackerBeads, Harness: domain.HarnessClaudeCode},
			want: `  "tracker": "beads",
  "beads": {}
}
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFileSystem()

			if _, err := NewInitialize(files, claudeInstalled()).Run(t.Context(), tc.request, nil); err != nil {
				t.Fatalf("Run: %v", err)
			}

			got := string(files.files[settingsFull])
			if !strings.HasSuffix(got, tc.want) {
				t.Errorf("settings.json =\n%s\nwant it to end with\n%s", got, tc.want)
			}

			if strings.Contains(got, "null") {
				t.Errorf("settings.json =\n%s\nwant no null in it", got)
			}
		})
	}
}

func TestRunSkipsSettingsThatAreAlreadyThere(t *testing.T) {
	files := newFakeFileSystem()
	files.files[settingsFull] = []byte("{}\n")

	report, err := NewInitialize(files, claudeInstalled()).Run(
		t.Context(),
		Request{Dir: workingDir, Tracker: domain.TrackerBeads, Harness: domain.HarnessClaudeCode},
		nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	results := report.Results()
	if len(results) != 2 || results[0].Outcome != domain.OutcomeSkipped {
		t.Fatalf("Results() = %+v, want the settings step to have skipped", results)
	}

	want := ".codefall/settings.json already exists (use --force to rewrite it)"
	if results[0].Detail != want {
		t.Errorf("detail = %q, want %q", results[0].Detail, want)
	}

	if got := string(files.files[settingsFull]); got != "{}\n" {
		t.Errorf("settings.json = %q, want it untouched", got)
	}
}

func TestRunRewritesSettingsWithForce(t *testing.T) {
	files := newFakeFileSystem()
	files.files[settingsFull] = []byte("{}\n")

	report, err := NewInitialize(files, claudeInstalled()).Run(
		t.Context(),
		Request{Dir: workingDir, Tracker: domain.TrackerBeads, Harness: domain.HarnessClaudeCode, Force: true},
		nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := report.Results()[0].Outcome; got != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", got)
	}

	if got := string(files.files[settingsFull]); !strings.Contains(got, `"tracker": "beads"`) {
		t.Errorf("settings.json = %q, want it rewritten", got)
	}
}

// A step that cannot finish ends the run, and the error names the step so the reader knows how far
// init got. The observer is told the step started and never told it finished.
func TestRunStopsOnAStepThatFails(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*fakeFileSystem)
		want  string
	}{
		{
			name:  "the settings could not be built",
			setup: func(*fakeFileSystem) {},
			want:  "needs a repository",
		},
		{
			name: "the directory could not be made",
			setup: func(f *fakeFileSystem) {
				f.errs["mkdir "+codefallDir] = errors.New("read-only file system")
			},
			want: "create .codefall/",
		},
		{
			name: "the file could not be written",
			setup: func(f *fakeFileSystem) {
				f.errs["write "+settingsFull] = errors.New("no space left on device")
			},
			want: "write .codefall/settings.json",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFileSystem()
			tc.setup(files)

			observer := &recordingObserver{}

			request := Request{Dir: workingDir, Tracker: domain.TrackerGitHub, Harness: domain.HarnessClaudeCode}
			if tc.name != "the settings could not be built" {
				request.GitHubRepo = mo.Some("owner/name")
			}

			report, err := NewInitialize(files, claudeInstalled()).Run(t.Context(), request, observer)
			if err == nil {
				t.Fatalf("Run = %+v, want an error", report)
			}

			if !strings.HasPrefix(err.Error(), domain.SettingsStep.ID+": ") {
				t.Errorf("error = %q, want it to name the step it failed in", err)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}

			if len(report.Results()) != 0 {
				t.Errorf("Results() = %+v, want none", report.Results())
			}

			if len(observer.started) != 1 || len(observer.finished) != 0 {
				t.Errorf("observer saw %d started and %d finished, want 1 and 0",
					len(observer.started), len(observer.finished))
			}
		})
	}
}

func TestRunReportsAnUnreadableSettingsFile(t *testing.T) {
	files := newFakeFileSystem()
	files.errs[settingsFull] = errors.New("permission denied")

	if _, err := NewInitialize(files, claudeInstalled()).Run(
		t.Context(),
		Request{Dir: workingDir, Tracker: domain.TrackerBeads, Harness: domain.HarnessClaudeCode},
		nil,
	); err == nil || !strings.Contains(err.Error(), "read .codefall/settings.json") {
		t.Errorf("Run error = %v, want it to say the settings could not be read", err)
	}
}

func TestRunStopsOnACancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := NewInitialize(newFakeFileSystem(), claudeInstalled()).Run(
		ctx,
		Request{Dir: workingDir, Tracker: domain.TrackerBeads, Harness: domain.HarnessClaudeCode},
		nil,
	)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run error = %v, want a context.Canceled", err)
	}
}

// --- the questions presentation asks -----------------------------------------------------------

func TestSettingsExist(t *testing.T) {
	files := newFakeFileSystem()

	exists, err := NewInitialize(files, newFakeCommandRunner()).SettingsExist(workingDir)
	if err != nil || exists {
		t.Errorf("SettingsExist of an empty directory = %v, %v, want false, nil", exists, err)
	}

	files.files[settingsFull] = []byte("{}")

	exists, err = NewInitialize(files, newFakeCommandRunner()).SettingsExist(workingDir)
	if err != nil || !exists {
		t.Errorf("SettingsExist with a settings file = %v, %v, want true, nil", exists, err)
	}

	files.errs[settingsFull] = errors.New("permission denied")

	if _, err := NewInitialize(files, newFakeCommandRunner()).SettingsExist(workingDir); err == nil {
		t.Error("SettingsExist of an unreadable file = nil error, want an error")
	}
}

func TestSuggestGitHubRepo(t *testing.T) {
	const command = "gh repo view --json nameWithOwner --jq .nameWithOwner"

	installed := func() *fakeCommandRunner {
		runner := newFakeCommandRunner()
		runner.paths["gh"] = "/opt/homebrew/bin/gh"

		return runner
	}

	t.Run("the repository gh names, trimmed", func(t *testing.T) {
		runner := installed()
		runner.runs[command] = CommandResult{Stdout: "lividlabs/codefall-cli\n"}

		got := NewInitialize(newFakeFileSystem(), runner).SuggestGitHubRepo(t.Context(), workingDir)
		if repo, ok := got.Get(); !ok || repo != "lividlabs/codefall-cli" {
			t.Errorf("SuggestGitHubRepo = %v, want Some(%q)", got, "lividlabs/codefall-cli")
		}

		// gh answers about the directory init is running in, not about wherever the process started.
		if len(runner.calls) != 1 || runner.calls[0].dir != workingDir {
			t.Errorf("calls = %+v, want one in %q", runner.calls, workingDir)
		}
	})

	for _, tc := range []struct {
		name   string
		runner func() *fakeCommandRunner
	}{
		{
			name:   "gh is not installed",
			runner: newFakeCommandRunner,
		},
		{
			name: "the directory is not a GitHub repository",
			runner: func() *fakeCommandRunner {
				runner := installed()
				runner.runs[command] = CommandResult{ExitCode: 1, Stderr: "not a git repository\n"}

				return runner
			},
		},
		{
			name: "gh could not be started",
			runner: func() *fakeCommandRunner {
				runner := installed()
				runner.errs[command] = errors.New("broken pipe")

				return runner
			},
		},
		{
			name: "gh said nothing",
			runner: func() *fakeCommandRunner {
				runner := installed()
				runner.runs[command] = CommandResult{Stdout: "  \n"}

				return runner
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := NewInitialize(newFakeFileSystem(), tc.runner()).SuggestGitHubRepo(t.Context(), workingDir)
			if got.IsPresent() {
				t.Errorf("SuggestGitHubRepo = %v, want None", got)
			}
		})
	}
}
