package application

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

// skillsRequest is a run on a harness that reads the .agents/skills convention.
func skillsRequest() Request {
	return Request{Dir: workingDir, Tracker: settings.TrackerBeads, Harness: domain.HarnessCodex}
}

// The step copies the embedded plugin tree into the project's .agents/, one Fetch call, one
// destination.
func TestSkillsStepCopiesTheEmbeddedTree(t *testing.T) {
	fetcher := newFakePluginFetcher()

	report, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := report.Results()[1]
	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	want := "installed codefall's skills into .agents/"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if len(fetcher.calls) != 1 || fetcher.calls[0].dir != filepath.Join(workingDir, ".agents") {
		t.Errorf("fetcher calls = %+v, want one fetch into %q",
			fetcher.calls, filepath.Join(workingDir, ".agents"))
	}
}

// The fetcher failing stops the run in the plugin step, which is the second step of five, so the
// report carries the one step that is already done.
func TestSkillsStepStopsTheRunWhenTheCopyFails(t *testing.T) {
	fetcher := newFakePluginFetcher()
	fetcher.err = errors.New("disk full")

	_, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
	if err == nil ||
		!strings.HasPrefix(err.Error(), domain.PluginStep.ID+": ") ||
		!strings.Contains(err.Error(), "install the embedded plugin: disk full") {
		t.Errorf("Run error = %v, want it to name the plugin step and the copy that failed", err)
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
