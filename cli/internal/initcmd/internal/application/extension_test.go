package application

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

// extensionRequest is a run that has nothing to do but install the extension: the settings are already
// there, so the first step skips and what the test watches is the second.
func extensionRequest() Request {
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

// The step contains one mechanism: embed-copy from the binary onto the harness's own skills
// directory. For Claude Code that is .claude/, so both the destination and the message don't mean
// everything: they mean that the embedded extension.
func TestExtensionStepCopiesIntoTheHarnessSkillsDirectory(t *testing.T) {
	fetcher := newFakeExtensionSource()

	report, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(t.Context(), extensionRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := report.Results()[1]
	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	want := "installed codefall's skills into .claude/"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if len(fetcher.calls) != 1 || fetcher.calls[0].dir != filepath.Join(workingDir, ".claude") {
		t.Errorf("fetcher calls = %+v, want one fetch into %q",
			fetcher.calls, filepath.Join(workingDir, ".claude"))
	}
}

// A claude-code harness that cannot be read runs no extension step: the report commands stay empty.
// Its companion tests for enforcement about manifest write aside — done behavior.
func TestExtensionStepStopsTheRunWhenTheCopyFails(t *testing.T) {
	fetcher := newFakeExtensionSource()
	fetcher.err = errors.New("disk full")

	_, err := NewInitialize(settled(""), toolsInstalled(), fetcher).Run(t.Context(), extensionRequest(), nil)
	if err == nil ||
		!strings.HasPrefix(err.Error(), domain.ExtensionStep.ID+": ") ||
		!strings.Contains(err.Error(), "install the embedded extension: disk full") {
		t.Errorf("Run error = %v, want it to name the extension step and the copy that failed", err)
	}
}

// The mechanism table is a guard: presentation refuses an unknown harness before the use case sees
// it, so this is a path a run should never take.
func TestExtensionStepRefusesAHarnessItDoesNotKnow(t *testing.T) {
	request := extensionRequest()
	request.Harness = "aider"

	_, err := NewInitialize(settled(""), toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), request, nil)
	if err == nil || !strings.Contains(err.Error(), `harness "aider" has no extension mechanism`) {
		t.Errorf("Run error = %v, want it to say the harness has no extension mechanism", err)
	}
}
