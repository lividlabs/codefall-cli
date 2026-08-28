package application

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/internal/shared/settings"
)

// pluginRequest is a run that has nothing to do but install the plugin: the settings are already
// there, so the first step skips and what the test watches is the second.
func pluginRequest() Request {
	return Request{Dir: workingDir, Tracker: settings.TrackerBeads, Harness: domain.HarnessClaudeCode}
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
// asks all four, and the beads step asks `bd info` again to decide whether it has anything to do.
var probes = []string{gitWorkTree, gitStaged, gitStatus, beadsInfo}

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

// The two halves of the file are read separately, so what the step does is whichever of them is
// missing. The project this repository is in — a plugin enabled with no marketplace declared at
// project scope — is the case that used to be skipped entirely.
//
// Both commands run in the directory init was pointed at, not wherever the process started, and the
// marketplace is declared before the plugin that comes from it is installed.
func TestPluginStepDeclaresWhateverTheProjectIsMissing(t *testing.T) {
	const (
		declared = `"extraKnownMarketplaces": {"codefall": ` +
			`{"source": {"source": "github", "repo": "lividlabs/codefall-plugin"}}}`
		enabled = `"enabledPlugins": {"codefall@codefall": true}`
	)

	for _, tc := range []struct {
		name     string
		settings string
		commands []string
		want     string
	}{
		{
			name:     "no .claude/settings.json at all",
			settings: "",
			commands: []string{marketplaceAdd, pluginInstall},
			want: "declared the codefall marketplace and installed codefall@codefall " +
				"for this project (.claude/settings.json)",
		},
		{
			name:     "settings that declare nothing",
			settings: `{"permissions": {"allow": []}}`,
			commands: []string{marketplaceAdd, pluginInstall},
			want: "declared the codefall marketplace and installed codefall@codefall " +
				"for this project (.claude/settings.json)",
		},
		{
			name:     "the plugin is off rather than absent",
			settings: `{"enabledPlugins": {"codefall@codefall": false}}`,
			commands: []string{marketplaceAdd, pluginInstall},
			want: "declared the codefall marketplace and installed codefall@codefall " +
				"for this project (.claude/settings.json)",
		},
		{
			name:     "the marketplace is already declared",
			settings: "{" + declared + "}",
			commands: []string{pluginInstall},
			want:     "installed codefall@codefall for this project (.claude/settings.json)",
		},
		{
			name:     "the plugin is enabled but the marketplace is not declared",
			settings: "{" + enabled + "}",
			commands: []string{marketplaceAdd},
			want:     "declared the codefall marketplace for this project (.claude/settings.json)",
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

			if result.Detail != tc.want {
				t.Errorf("detail = %q, want %q", result.Detail, tc.want)
			}

			if got := commands(runner); !slices.Equal(got, tc.commands) {
				t.Errorf("ran %q, want %q", got, tc.commands)
			}

			for _, call := range runner.calls {
				if call.dir != workingDir {
					t.Errorf("ran %q in %q, want %q", call.command, call.dir, workingDir)
				}
			}

			// The CLI writes the plugin's half of the file; this step never does. The hook step
			// writes the same file afterwards, and what it leaves behind still says everything the
			// file said before the run. A run that started from nothing has nothing to hold it to.
			if tc.settings != "" {
				assertKeysSurvive(t, files.files[claudeFull], tc.settings)
			}
		})
	}
}

// A project that already has both is the only one with nothing to do: the file says the marketplace
// is declared and the plugin enabled, and nothing is run.
func TestPluginStepSkipsAProjectThatHasBoth(t *testing.T) {
	runner := toolsInstalled()

	report, err := NewInitialize(settled(
		`{"enabledPlugins": {"codefall@codefall": true},`+
			`"extraKnownMarketplaces": {"codefall": {"source": "lividlabs/codefall-plugin"}}}`,
	), runner).Run(t.Context(), pluginRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := report.Results()[1]
	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	want := "the codefall marketplace is declared and codefall@codefall enabled in .claude/settings.json"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if got := commands(runner); len(got) != 0 {
		t.Errorf("ran %q, want nothing", got)
	}
}

// A file that is there and says nothing is the same answer as one that is not there. Both steps that
// read it go on to do their work, rather than one refusing what encoding/json would have accepted.
func TestBothStepsTreatAFileThatSaysNothingAsAMissingOne(t *testing.T) {
	for _, tc := range []struct{ name, settings string }{
		{"empty", ""},
		{"whitespace only", " \n\t "},
		{"the JSON literal null", "null\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := settled("")
			files.files[claudeFull] = []byte(tc.settings)

			runner := toolsInstalled()

			report, err := NewInitialize(files, runner).Run(t.Context(), pluginRequest(), nil)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}

			if got := commands(runner); !slices.Equal(got, []string{marketplaceAdd, pluginInstall}) {
				t.Errorf("ran %q, want both commands", got)
			}

			if got := report.Results()[3].Outcome; got != domain.OutcomeDone {
				t.Errorf("the hook step = %v, want DONE", got)
			}

			if got := string(files.files[claudeFull]); got != beadsHook {
				t.Errorf(".claude/settings.json =\n%s\nwant\n%s", got, beadsHook)
			}
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
			name:  "the file is a JSON array",
			files: func() *fakeFileSystem { return settled(`["codefall@codefall"]`) },
			want:  "decode .claude/settings.json",
		},
		{
			name:  "the file is a JSON string",
			files: func() *fakeFileSystem { return settled(`"codefall@codefall"`) },
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
