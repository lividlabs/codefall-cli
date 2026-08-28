// Package domain holds initcmd's entities and value objects: the settings a project is initialised
// with, the harnesses codefall knows about, and the steps a run is made of. It performs no IO and
// names no delivery, encoding, or infrastructure type (ADR-BASE-01).
//
// What .codefall/settings.json may contain is not initcmd's — doctor reads the same file — so the
// format lives in the pure shared module internal/shared/settings and the value objects here are
// built from its rules (ADR-003).
package domain

import (
	"fmt"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/internal/shared/settings"
)

// GitHubSettings is the tracker block for GitHub Issues. The project number is absent when the
// repository's issues are not organised into a GitHub Project.
type GitHubSettings struct {
	Repo    string
	Project mo.Option[int]
}

// Settings is a complete, valid .codefall/settings.json as a value: the tracker, and the block that
// tracker selects. Exactly one tracker block is present, which is what the schema's oneOf says as
// well. Beads carries no fields of its own — bd keeps its configuration under .beads/ — so it needs
// no member here.
type Settings struct {
	Tracker string
	GitHub  mo.Option[GitHubSettings]
}

// NewSettings builds settings from the values a survey or a set of flags collected, and is the only
// way to obtain them: a Settings value that exists is one that may be written. Which values are
// acceptable is the format's to say; which combinations of them make sense for a run is initcmd's,
// and that is what this constructor adds.
func NewSettings(tracker string, repo mo.Option[string], project mo.Option[int]) (Settings, error) {
	name, err := settings.ParseTracker(tracker)
	if err != nil {
		return Settings{}, err
	}

	if name != settings.TrackerGitHub {
		if repo.IsPresent() {
			return Settings{}, fmt.Errorf("a repository is only used when the tracker is %q, not %q", settings.TrackerGitHub, name)
		}

		if project.IsPresent() {
			return Settings{}, fmt.Errorf("a project number is only used when the tracker is %q, not %q", settings.TrackerGitHub, name)
		}

		return Settings{Tracker: name}, nil
	}

	github, err := newGitHubSettings(repo, project)
	if err != nil {
		return Settings{}, err
	}

	return Settings{Tracker: name, GitHub: mo.Some(github)}, nil
}

func newGitHubSettings(repo mo.Option[string], project mo.Option[int]) (GitHubSettings, error) {
	name, ok := repo.Get()
	if !ok {
		return GitHubSettings{}, fmt.Errorf("tracker %q needs a repository", settings.TrackerGitHub)
	}

	if err := settings.ValidateRepo(name); err != nil {
		return GitHubSettings{}, err
	}

	if number, ok := project.Get(); ok {
		if err := settings.ValidateProject(number); err != nil {
			return GitHubSettings{}, err
		}
	}

	return GitHubSettings{Repo: name, Project: project}, nil
}
