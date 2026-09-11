package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

// Where settings live, relative to the directory init is run in. The display form is what a person
// reads in a report; the path form is what the file system is given.
const (
	settingsDir  = ".codefall"
	settingsFile = "settings.json"
	settingsName = settingsDir + "/" + settingsFile
)

// SettingsExist reports whether dir already has settings. Presentation asks before it prompts: there
// is no point surveying a project for answers that will not be written.
func (i *Initialize) SettingsExist(dir string) (bool, error) {
	_, err := i.files.ReadFile(settingsPath(dir))

	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	default:
		return false, fmt.Errorf("read %s: %w", settingsName, err)
	}
}

// settings is the first step of a run: it writes .codefall/settings.json, the file doctor checks.
//
// Settings that are already there are left alone unless the run asked for them to be rewritten,
// because init is safe to run again — the steps after this one still have work to do in a project
// that is half set up.
func (i *Initialize) settings(_ context.Context, request Request) (domain.StepResult, error) {
	exists, err := i.SettingsExist(request.Dir)
	if err != nil {
		return domain.StepResult{}, err
	}

	if exists && !request.Force {
		return domain.SettingsStep.Skipped(settingsName + " already exists (use --force to rewrite it)"), nil
	}

	chosen, err := domain.NewSettings(request.Tracker, request.GitHubRepo, request.GitHubProject)
	if err != nil {
		return domain.StepResult{}, err
	}

	data, err := encodeSettings(chosen)
	if err != nil {
		return domain.StepResult{}, err
	}

	if err := i.files.MkdirAll(filepath.Join(request.Dir, settingsDir)); err != nil {
		return domain.StepResult{}, fmt.Errorf("create %s/: %w", settingsDir, err)
	}

	if err := i.files.WriteFile(settingsPath(request.Dir), data); err != nil {
		return domain.StepResult{}, fmt.Errorf("write %s: %w", settingsName, err)
	}

	return domain.SettingsStep.Done(fmt.Sprintf("wrote %s (%s)", settingsName, describe(chosen))), nil
}

func settingsPath(dir string) string {
	return filepath.Join(dir, settingsDir, settingsFile)
}

// describe is what the step reports it wrote, in the terms the person answering the survey used.
//
// A settings value is named `chosen` throughout this file because `settings` is now the shared
// module that defines the format (ADR-003), and a local of that name would hide it.
func describe(chosen domain.Settings) string {
	parts := []string{"tracker: " + chosen.Tracker}

	if github, ok := chosen.GitHub.Get(); ok {
		parts = append(parts, "repo: "+github.Repo)

		if project, ok := github.Project.Get(); ok {
			parts = append(parts, fmt.Sprintf("project: %d", project))
		}
	}

	return strings.Join(parts, ", ")
}

// settingsDocument is the settings file's JSON shape. Encoding lives here rather than in the domain,
// which must not name an encoding, or in infrastructure, which takes bytes — the same reason doctor
// decodes in its application layer. Field order is key order, and it is chosen so the file opens
// with what identifies it and closes with the block the tracker selects.
type settingsDocument struct {
	Schema  string                    `json:"$schema"`
	Version int                       `json:"version"`
	Tracker string                    `json:"tracker"`
	Beads   mo.Option[beadsDocument]  `json:"beads,omitzero"`
	GitHub  mo.Option[gitHubDocument] `json:"github,omitzero"`
}

// beadsDocument is the empty block the schema requires for the Beads tracker: bd keeps its own
// configuration under .beads/, and the block exists so the "exactly one tracker block" rule reads
// the same for both trackers.
//
// IsZero is what keeps it in the file. `omitzero` asks mo.Option, which asks the value it holds, and
// a struct with no fields is always the zero value — so without this the present block would be
// dropped as if it were absent.
type beadsDocument struct{}

func (beadsDocument) IsZero() bool { return false }

// gitHubDocument is the GitHub tracker's block. The project number is omitted rather than written as
// null when the repository's issues are not organised into a project.
type gitHubDocument struct {
	Repo    string         `json:"repo"`
	Project mo.Option[int] `json:"project,omitzero"`
}

func (gitHubDocument) IsZero() bool { return false }

// encodeSettings renders settings as the bytes that go in the file: two-space indent and a trailing
// newline, so the result is what a person would have written by hand and diffs a line at a time.
func encodeSettings(chosen domain.Settings) ([]byte, error) {
	document := settingsDocument{
		Schema:  settings.SchemaID,
		Version: settings.Version,
		Tracker: chosen.Tracker,
	}

	// One case per tracker, so adding a tracker is a case here and a block above rather than a
	// silent omission. NewSettings has already rejected any tracker the format does not know, which
	// is what makes the default unreachable.
	switch chosen.Tracker {
	case settings.TrackerGitHub:
		github, ok := chosen.GitHub.Get()
		if !ok {
			return nil, fmt.Errorf("tracker %q has no github block", chosen.Tracker)
		}

		document.GitHub = mo.Some(gitHubDocument{Repo: github.Repo, Project: github.Project})
	case settings.TrackerBeads:
		document.Beads = mo.Some(beadsDocument{})
	default:
		return nil, fmt.Errorf("tracker %q has no block to write", chosen.Tracker)
	}

	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", settingsName, err)
	}

	return append(data, '\n'), nil
}
