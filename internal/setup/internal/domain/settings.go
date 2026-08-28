// Package domain holds setup's entities and value objects: the settings a project is initialised
// with, the trackers and harnesses codefall knows about, and the steps a run is made of. It performs
// no IO and names no delivery, encoding, or infrastructure type (ADR-BASE-01).
package domain

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/samber/mo"
)

// The published constants of the settings format. The schema test holds schemas/settings.schema.json
// equal to these, so what setup writes and what the schema promises cannot drift apart.
const (
	SettingsVersion  = 1
	SettingsSchemaID = "https://raw.githubusercontent.com/lividlabs/codefall-cli/main/schemas/settings.schema.json"
	RepoPattern      = `^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`
)

// The issue trackers codefall can be pointed at.
const (
	TrackerBeads  = "beads"
	TrackerGitHub = "github"
)

// The coding harnesses codefall can set up. Claude Code is the only one today; the flag exists so
// the later steps of init have something to branch on.
const HarnessClaudeCode = "claude-code"

var repoRegexp = regexp.MustCompile(RepoPattern)

// The two lists are sorted once, here. Order is not cosmetic for trackers: the schema test holds the
// schema's tracker enum equal to this list, and doctor's own list is sorted the same way.
var (
	trackers  = slices.Sorted(slices.Values([]string{TrackerBeads, TrackerGitHub}))
	harnesses = slices.Sorted(slices.Values([]string{HarnessClaudeCode}))
)

// Trackers returns the known tracker names, sorted.
func Trackers() []string {
	return slices.Clone(trackers)
}

// Harnesses returns the supported harness names, sorted.
func Harnesses() []string {
	return slices.Clone(harnesses)
}

// ParseTracker returns the tracker name when it is one codefall knows, and an error listing the
// known ones when it is not.
func ParseTracker(name string) (string, error) {
	if slices.Contains(trackers, name) {
		return name, nil
	}

	return "", fmt.Errorf("unknown tracker %q (known trackers: %s)", name, strings.Join(trackers, ", "))
}

// ParseHarness returns the harness name when codefall can set it up, and an error naming the ones it
// can when it cannot. A harness codefall does not support yet is not a typo, so the message says so
// rather than calling the value unknown.
func ParseHarness(name string) (string, error) {
	if slices.Contains(harnesses, name) {
		return name, nil
	}

	return "", fmt.Errorf("harness %q is not supported yet (supported: %s)", name, strings.Join(harnesses, ", "))
}

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
// way to obtain them: a Settings value that exists is one that may be written.
func NewSettings(tracker string, repo mo.Option[string], project mo.Option[int]) (Settings, error) {
	name, err := ParseTracker(tracker)
	if err != nil {
		return Settings{}, err
	}

	if name != TrackerGitHub {
		if repo.IsPresent() {
			return Settings{}, fmt.Errorf("a repository is only used when the tracker is %q, not %q", TrackerGitHub, name)
		}

		if project.IsPresent() {
			return Settings{}, fmt.Errorf("a project number is only used when the tracker is %q, not %q", TrackerGitHub, name)
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
		return GitHubSettings{}, fmt.Errorf("tracker %q needs a repository", TrackerGitHub)
	}

	if err := ValidateRepo(name); err != nil {
		return GitHubSettings{}, err
	}

	if number, ok := project.Get(); ok && number < 1 {
		return GitHubSettings{}, fmt.Errorf("project number %d must be a positive integer", number)
	}

	return GitHubSettings{Repo: name, Project: project}, nil
}

// ValidateRepo reports whether a repository is written the way GitHub names one. It is exported
// because a form validates the field as it is typed, before there are enough answers to build
// settings from.
func ValidateRepo(repo string) error {
	if !repoRegexp.MatchString(repo) {
		return fmt.Errorf("repository %q must be written as owner/name", repo)
	}

	return nil
}
