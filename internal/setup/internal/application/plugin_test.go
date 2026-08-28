package application

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// pluginRequest is a run that has nothing to do but install the plugin: the settings are already
// there, so the first step skips and what the test watches is the second.
func pluginRequest() Request {
	return Request{Dir: workingDir, Tracker: domain.TrackerBeads, Harness: domain.HarnessClaudeCode}
}

// settled is a file system whose .codefall/settings.json is already written, with whatever
// .claude/settings.json the test wants beside it.
func settled(claudeSettings string) *fakeFileSystem {
	files := newFakeFileSystem()
	files.files[settingsFull] = []byte("{}\n")

	if claudeSettings != "" {
		files.files[claudeFull] = []byte(claudeSettings)
	}

	return files
}

// probes are the questions a run asks about the directory rather than work it does in it: preflight
// asks all three, and the beads step asks `bd info` again to decide whether it has anything to do.
var probes = []string{gitWorkTree, gitStaged, beadsInfo}

// commands is the work the runner was asked to do, in order, with the probes left out — a test about
// one step reads better without the questions every run asks.
func commands(runner *fakeCommandRunner) []string {
	got := make([]string, 0, len(runner.calls))

	for _, call := range runner.calls {
		if !slices.Contains(probes, call.command) {
			got = append(got, call.command)
		}
	}

	return got
}

// A plugin the project already enables is left alone: the file says so, and nothing is run.
func TestPluginStepSkipsAPluginThatIsAlreadyEnabled(t *testing.T) {
	runner := toolsInstalled()

	report, err := NewInitialize(
		settled(`{"enabledPlugins": {"codefall@codefall": true}}`), runner,
	).Run(t.Context(), pluginRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := report.Results()[1]
	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	want := "codefall@codefall is already enabled in .claude/settings.json"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if got := commands(runner); len(got) != 0 {
		t.Errorf("ran %q, want nothing", got)
	}
}

// The marketplace is declared before the plugin is installed, and only when the project does not
// declare it already. Both commands run in the directory init was pointed at, not wherever the
// process started.
func TestPluginStepInstallsThePlugin(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings string
		want     []string
	}{
		{
			name:     "no .claude/settings.json at all",
			settings: "",
			want:     []string{marketplaceAdd, pluginInstall},
		},
		{
			name:     "settings that declare nothing",
			settings: `{"permissions": {"allow": []}}`,
			want:     []string{marketplaceAdd, pluginInstall},
		},
		{
			name:     "the plugin is off rather than absent",
			settings: `{"enabledPlugins": {"codefall@codefall": false}}`,
			want:     []string{marketplaceAdd, pluginInstall},
		},
		{
			name: "the marketplace is already declared",
			settings: `{"extraKnownMarketplaces": {"codefall": ` +
				`{"source": {"source": "github", "repo": "lividlabs/codefall-plugin"}}}}`,
			want: []string{pluginInstall},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := settled(tc.settings)
			runner := toolsInstalled()

			report, err := NewInitialize(files, runner).Run(t.Context(), pluginRequest(), nil)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}

			result := report.Results()[1]
			if result.Outcome != domain.OutcomeDone {
				t.Errorf("outcome = %v, want DONE", result.Outcome)
			}

			want := "installed codefall@codefall for this project (.claude/settings.json)"
			if result.Detail != want {
				t.Errorf("detail = %q, want %q", result.Detail, want)
			}

			if got := commands(runner); !slices.Equal(got, tc.want) {
				t.Errorf("ran %q, want %q", got, tc.want)
			}

			for _, call := range runner.calls {
				if call.dir != workingDir {
					t.Errorf("ran %q in %q, want %q", call.command, call.dir, workingDir)
				}
			}

			// The CLI writes the plugin's half of the file; this step never does. The hook step
			// writes the same file afterwards, and what it leaves behind still says everything the
			// file said before the run.
			assertKeysSurvive(t, files.files[claudeFull], tc.settings)
		})
	}
}

// A refusal from the harness CLI stops the run, and the message carries what the tool said —
// stderr when there is any, and stdout when there is not.
func TestPluginStepStopsTheRunWhenTheHarnessRefuses(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
		result  CommandResult
		want    string
	}{
		{
			name:    "the marketplace could not be added",
			command: marketplaceAdd,
			result:  CommandResult{ExitCode: 1, Stderr: "repository not found\nsee the docs\n"},
			want:    "claude plugin marketplace add lividlabs/codefall-plugin --scope project exited 1: repository not found",
		},
		{
			name:    "the plugin could not be installed",
			command: pluginInstall,
			result:  CommandResult{ExitCode: 2, Stdout: "no such plugin\n"},
			want:    "claude plugin install codefall@codefall --scope project -y exited 2: no such plugin",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := toolsInstalled()
			runner.runs[tc.command] = tc.result

			observer := &recordingObserver{}

			report, err := NewInitialize(settled(""), runner).Run(t.Context(), pluginRequest(), observer)
			if err == nil {
				t.Fatalf("Run = %+v, want an error", report)
			}

			if !strings.HasPrefix(err.Error(), domain.PluginStep.ID+": ") {
				t.Errorf("error = %q, want it to name the step it failed in", err)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}

			if len(report.Results()) != 0 {
				t.Errorf("Results() = %+v, want none", report.Results())
			}

			// The settings step finished; the plugin step started and never did.
			if len(observer.started) != 2 || len(observer.finished) != 1 {
				t.Errorf("observer saw %d started and %d finished, want 2 and 1",
					len(observer.started), len(observer.finished))
			}
		})
	}
}

// A CLI that could not be started at all is a different failure from one that refused, and says so.
func TestPluginStepStopsTheRunWhenTheHarnessCannotBeStarted(t *testing.T) {
	runner := toolsInstalled()
	runner.errs[marketplaceAdd] = errors.New("broken pipe")

	_, err := NewInitialize(settled(""), runner).Run(t.Context(), pluginRequest(), nil)
	if err == nil || !strings.Contains(err.Error(), "run claude plugin marketplace add") {
		t.Errorf("Run error = %v, want it to say the command could not be run", err)
	}
}

// Settings that cannot be read, or are not settings, stop the run naming the file: the CLI is about
// to merge into it.
func TestPluginStepReportsSettingsItCannotRead(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files func() *fakeFileSystem
		want  string
	}{
		{
			name: "the file cannot be read",
			files: func() *fakeFileSystem {
				files := settled("")
				files.errs[claudeFull] = errors.New("permission denied")

				return files
			},
			want: "read .claude/settings.json",
		},
		{
			name:  "the file is not JSON",
			files: func() *fakeFileSystem { return settled("{ not json") },
			want:  "decode .claude/settings.json",
		},
		{
			name:  "the file is JSON but not an object",
			files: func() *fakeFileSystem { return settled(`["codefall@codefall"]`) },
			want:  "decode .claude/settings.json",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := toolsInstalled()

			_, err := NewInitialize(tc.files(), runner).Run(t.Context(), pluginRequest(), nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Run error = %v, want it to mention %q", err, tc.want)
			}

			if got := commands(runner); len(got) != 0 {
				t.Errorf("ran %q, want nothing", got)
			}
		})
	}
}

// The switch is where the next harness lands; today anything but Claude Code is a failure rather
// than a step that quietly does nothing. Presentation refuses these values before the use case sees
// them, so this is a guard and not a path a person can take.
func TestPluginStepRefusesAHarnessItDoesNotKnow(t *testing.T) {
	request := pluginRequest()
	request.Harness = "aider"

	_, err := NewInitialize(settled(""), toolsInstalled()).Run(t.Context(), request, nil)
	if err == nil || !strings.Contains(err.Error(), `harness "aider" has no plugin to install`) {
		t.Errorf("Run error = %v, want it to say the harness has no plugin", err)
	}
}
