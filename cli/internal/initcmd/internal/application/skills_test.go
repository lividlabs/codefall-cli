package application

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

// skillsRequest is a run on a harness that reads the .agents/skills convention.
func skillsRequest() Request {
	return Request{Dir: workingDir, Tracker: settings.TrackerBeads, Harness: domain.HarnessCodex}
}

// The step copies the embedded extension tree into the project's .agents/, one Fetch call, one
// destination.
func TestSkillsStepCopiesTheEmbeddedTree(t *testing.T) {
	fetcher := newFakeExtensionSource()

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

// The fetcher failing stops the run in the extension step, which is the second step of five, so the
// report carries the one step that is already done.
func TestSkillsStepStopsTheRunWhenTheCopyFails(t *testing.T) {
	fetcher := newFakeExtensionSource()
	fetcher.err = errors.New("disk full")

	_, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(t.Context(), skillsRequest(), nil)
	if err == nil ||
		!strings.HasPrefix(err.Error(), domain.ExtensionStep.ID+": ") ||
		!strings.Contains(err.Error(), "install the embedded extension: disk full") {
		t.Errorf("Run error = %v, want it to name the extension step and the copy that failed", err)
	}
}

// A skills-directory harness runs all six steps: the extension step installs where the mechanism
// installs, the hook step registers there too, and the rest are untouched.
func TestSkillsRunStillRunsEveryStep(t *testing.T) {
	report, err := NewInitialize(settled(""), toolsInstalled(), newFakeExtensionSource()).Run(
		t.Context(), skillsRequest(), nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	results := report.Results()
	if len(results) != 6 {
		t.Fatalf("Results() = %+v, want a result for each of the six steps", results)
	}

	if got := results[3].Outcome; got != domain.OutcomeDone {
		t.Errorf("the hook step = %v, want DONE", got)
	}

	want := "merged codefall's hooks into .codex/hooks.json"
	if got := results[3].Detail; got != want {
		t.Errorf("the hook step's detail = %q, want %q", got, want)
	}

	steps := []domain.Step{
		domain.SettingsStep, domain.ExtensionStep, domain.BeadsStep, domain.HookStep,
		domain.AgentsStep, domain.IgnoreStep,
	}
	got := make([]domain.Step, 0, len(results))
	for _, result := range results {
		got = append(got, result.Step)
	}
	if !slices.Equal(got, steps) {
		t.Errorf("steps = %+v, want %+v", got, steps)
	}
}

// The manifest is what the upgrade gate reads to decide a rerun has nothing to do, so it is written
// only when every step has succeeded. Written in the extension step, it made that claim with the
// hook and AGENTS.md steps still to come: a run that failed at either left a manifest saying the
// install was current, and the next run reported it was already up to date without registering the
// hooks the failed run never got to.
func TestTheManifestRecordsOnlyARunThatFinished(t *testing.T) {
	t.Run("a run that finished", func(t *testing.T) {
		files := settled("")
		request := beadsRequest()
		request.CLIVersion = "v1.2.3"

		if _, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), request, nil); err != nil {
			t.Fatalf("Run: %v", err)
		}

		var recorded manifest
		if err := json.Unmarshal(files.files[filepath.Join(workingDir, domain.ManifestName)], &recorded); err != nil {
			t.Fatalf("decode %s: %v", domain.ManifestName, err)
		}

		want := manifest{Harness: domain.HarnessClaudeCode, Version: "v1.2.3",
			Files: []string{"skills/design/SKILL.md"}}
		if !reflect.DeepEqual(recorded, want) {
			t.Errorf("manifest = %+v, want %+v", recorded, want)
		}
	})

	t.Run("a run that failed a later step", func(t *testing.T) {
		files := settled(`{"hooks": "not an object"}`)

		_, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), beadsRequest(), nil)
		if err == nil {
			t.Fatal("Run = nil error, want the hook step to stop the run")
		}

		if _, wrote := files.files[filepath.Join(workingDir, domain.ManifestName)]; wrote {
			t.Error("a failed run recorded a manifest, want none until every step has succeeded")
		}
	})
}

// Installed is the record through the use-case boundary: what a finished run installed, and which
// harness it installed for. Both are what the gate compares; neither is any use on its own.
func TestInstalledReportsWhatAFinishedRunRecorded(t *testing.T) {
	for _, tc := range []struct {
		name    string
		body    string
		missing bool
		want    mo.Option[Installation]
	}{
		{name: "a manifest", body: `{"harness": "codex", "version": "v1.2.3"}`,
			want: mo.Some(Installation{Harness: "codex", Version: "v1.2.3"})},
		{name: "no manifest", missing: true, want: mo.None[Installation]()},
		{name: "a manifest naming no version", body: `{"harness": "codex"}`,
			want: mo.None[Installation]()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFileSystem()
			if !tc.missing {
				files.files[filepath.Join(workingDir, domain.ManifestName)] = []byte(tc.body)
			}

			got, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Installed(workingDir)
			if err != nil {
				t.Fatalf("Installed: %v", err)
			}

			if got != tc.want {
				t.Errorf("Installed = %v, want %v", got, tc.want)
			}
		})
	}
}
