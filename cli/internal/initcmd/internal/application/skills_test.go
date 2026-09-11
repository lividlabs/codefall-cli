package application

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

// skillsRequest is a run on a harness that reads the .agents/skills convention. What the test
// varies is the plugin version and the state of the tree.
func skillsRequest() Request {
	return Request{Dir: workingDir, Tracker: settings.TrackerBeads, Harness: domain.HarnessCodex}
}

var agentsManifest = filepath.Join(workingDir, ".agents", ".claude-plugin", "plugin.json")

// The step mirrors the plugin's release into the project's .agents/, and the fetcher is given the
// pinned version when the request has none.
func TestSkillsStepInstallsThePinnedRelease(t *testing.T) {
	fetcher := newFakePluginFetcher()

	report, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := report.Results()[1]
	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	want := "installed codefall's skills at 0.7.0 into .agents/"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if len(fetcher.calls) != 1 || fetcher.calls[0].version != domain.PluginVersion ||
		fetcher.calls[0].dir != filepath.Join(workingDir, ".agents") {
		t.Errorf("fetcher calls = %+v, want one fetch of %q into %q",
			fetcher.calls, domain.PluginVersion, filepath.Join(workingDir, ".agents"))
	}
}

// A version the request names is the one it is fetched from.
func TestSkillsStepUsesTheVersionItIsGiven(t *testing.T) {
	fetcher := newFakePluginFetcher()
	request := skillsRequest()
	request.PluginVersion = mo.Some("0.9.0")

	if _, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(
		t.Context(), request, nil,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(fetcher.calls) != 1 || fetcher.calls[0].version != "0.9.0" {
		t.Errorf("fetcher calls = %+v, want one fetch of 0.9.0", fetcher.calls)
	}
}

// A run that already has the pinned release is skipped: the manifest the previous run left behind
// says so, and the fetcher is not asked.
func TestSkillsStepSkipsTheReleaseItAlreadyHas(t *testing.T) {
	files := settled("")
	files.files[agentsManifest] = []byte(`{"version": "0.7.0"}`)
	fetcher := newFakePluginFetcher()

	report, err := NewInitialize(files, toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := report.Results()[1]
	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	want := "codefall's skills are already at 0.7.0 in .agents/"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if len(fetcher.calls) != 0 {
		t.Errorf("fetcher calls = %+v, want none", fetcher.calls)
	}
}

// A tree made from another release is rebuilt from the one the run is pointed at.
func TestSkillsStepRebuildsAnotherRelease(t *testing.T) {
	files := settled("")
	files.files[agentsManifest] = []byte(`{"version": "0.6.0"}`)
	fetcher := newFakePluginFetcher()

	report, err := NewInitialize(files, toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := report.Results()[1].Outcome; got != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", got)
	}

	if len(fetcher.calls) != 1 || fetcher.calls[0].version != domain.PluginVersion {
		t.Errorf("fetcher calls = %+v, want one fetch of %q", fetcher.calls, domain.PluginVersion)
	}
}

// A manifest with no version is the same as no manifest: the install runs rather than the version
// being compared against a blank.
func TestSkillsStepInstallsWhenTheManifestHasNoVersion(t *testing.T) {
	files := settled("")
	files.files[agentsManifest] = []byte(`{"name": "codefall"}`)
	fetcher := newFakePluginFetcher()

	if _, err := NewInitialize(files, toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(fetcher.calls) != 1 {
		t.Errorf("fetcher calls = %+v, want one fetch", fetcher.calls)
	}
}

// A manifest that cannot be read, or is not one, stops the run naming the file: the step is about
// to write the tree beside it.
func TestSkillsStepReportsAManifestItCannotRead(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files func() *fakeFileSystem
		want  string
	}{
		{
			name: "the file cannot be read",
			files: func() *fakeFileSystem {
				files := settled("")
				files.errs[agentsManifest] = errors.New("permission denied")

				return files
			},
			want: "read .agents/.claude-plugin/plugin.json",
		},
		{
			name:  "the file is not JSON",
			files: func() *fakeFileSystem { return settledManifest("{ not json") },
			want:  "decode .agents/.claude-plugin/plugin.json",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fetcher := newFakePluginFetcher()

			_, err := NewInitialize(tc.files(), toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Run error = %v, want it to mention %q", err, tc.want)
			}

			if len(fetcher.calls) != 0 {
				t.Errorf("fetcher calls = %+v, want none", fetcher.calls)
			}
		})
	}
}

func settledManifest(manifest string) *fakeFileSystem {
	files := settled("")
	files.files[agentsManifest] = []byte(manifest)

	return files
}

// The fetcher failing stops the run in the plugin step, which is the second step of five, so the
// report carries the one step that is already done.
func TestSkillsStepStopsTheRunWhenTheFetchFails(t *testing.T) {
	fetcher := newFakePluginFetcher()
	fetcher.err = errors.New("connection refused")

	_, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
	if err == nil ||
		!strings.HasPrefix(err.Error(), domain.PluginStep.ID+": ") ||
		!strings.Contains(err.Error(), "fetch the plugin at 0.7.0: connection refused") {
		t.Errorf("Run error = %v, want it to name the plugin step and the fetch that failed", err)
	}
}

// A skills-directory harness runs all five steps: the plugin step installs where the mechanism
// installs, the hook step skips, and the rest are untouched.
func TestSkillsRunStillRunsEveryStep(t *testing.T) {
	report, err := NewInitialize(settled(""), toolsInstalled(), newFakePluginFetcher()).Run(
		t.Context(), skillsRequest(), nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	results := report.Results()
	if len(results) != 5 {
		t.Fatalf("Results() = %+v, want a result for each of the five steps", results)
	}

	if got := results[3].Outcome; got != domain.OutcomeSkipped {
		t.Errorf("the hook step = %v, want SKIPPED", got)
	}

	want := "codefall has no session hook for codex yet"
	if got := results[3].Detail; got != want {
		t.Errorf("the hook step's detail = %q, want %q", got, want)
	}

	steps := []domain.Step{
		domain.SettingsStep, domain.PluginStep, domain.BeadsStep, domain.HookStep, domain.AgentsStep,
	}
	got := make([]domain.Step, 0, len(results))
	for _, result := range results {
		got = append(got, result.Step)
	}
	if !slices.Equal(got, steps) {
		t.Errorf("steps = %+v, want %+v", got, steps)
	}
}
