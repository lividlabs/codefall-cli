package domain

import (
	"strings"
	"testing"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

func TestNewSettingsForGitHub(t *testing.T) {
	value, err := NewSettings(settings.TrackerGitHub, mo.Some("lividlabs/codefall-cli"), mo.Some(3))
	if err != nil {
		t.Fatalf("NewSettings: %v", err)
	}

	if value.Tracker != settings.TrackerGitHub {
		t.Errorf("Tracker = %q, want %q", value.Tracker, settings.TrackerGitHub)
	}

	github, ok := value.GitHub.Get()
	if !ok {
		t.Fatalf("GitHub = %v, want a block", value.GitHub)
	}

	if github.Repo != "lividlabs/codefall-cli" {
		t.Errorf("GitHub.Repo = %q, want %q", github.Repo, "lividlabs/codefall-cli")
	}

	if got, ok := github.Project.Get(); !ok || got != 3 {
		t.Errorf("GitHub.Project = %v, want Some(3)", github.Project)
	}
}

func TestNewSettingsForGitHubWithoutAProject(t *testing.T) {
	value, err := NewSettings(settings.TrackerGitHub, mo.Some("owner/name"), mo.None[int]())
	if err != nil {
		t.Fatalf("NewSettings: %v", err)
	}

	github, _ := value.GitHub.Get()
	if github.Project.IsPresent() {
		t.Errorf("GitHub.Project = %v, want None", github.Project)
	}
}

func TestNewSettingsForBeadsCarriesNoGitHubBlock(t *testing.T) {
	value, err := NewSettings(settings.TrackerBeads, mo.None[string](), mo.None[int]())
	if err != nil {
		t.Fatalf("NewSettings: %v", err)
	}

	if value.Tracker != settings.TrackerBeads {
		t.Errorf("Tracker = %q, want %q", value.Tracker, settings.TrackerBeads)
	}

	if value.GitHub.IsPresent() {
		t.Errorf("GitHub = %v, want None", value.GitHub)
	}
}

func TestNewSettingsRejects(t *testing.T) {
	for _, tc := range []struct {
		name    string
		tracker string
		repo    mo.Option[string]
		project mo.Option[int]
		want    string
	}{
		{
			name:    "an unknown tracker",
			tracker: "linear",
			repo:    mo.None[string](),
			project: mo.None[int](),
			want:    "unknown tracker",
		},
		{
			name:    "github without a repository",
			tracker: settings.TrackerGitHub,
			repo:    mo.None[string](),
			project: mo.None[int](),
			want:    "needs a repository",
		},
		{
			name:    "a repository that is not owner/name",
			tracker: settings.TrackerGitHub,
			repo:    mo.Some("codefall-cli"),
			project: mo.None[int](),
			want:    "owner/name",
		},
		{
			name:    "a project number below one",
			tracker: settings.TrackerGitHub,
			repo:    mo.Some("owner/name"),
			project: mo.Some(0),
			want:    "positive integer",
		},
		{
			name:    "a repository under beads",
			tracker: settings.TrackerBeads,
			repo:    mo.Some("owner/name"),
			project: mo.None[int](),
			want:    "a repository is only used",
		},
		{
			name:    "a project number under beads",
			tracker: settings.TrackerBeads,
			repo:    mo.None[string](),
			project: mo.Some(3),
			want:    "a project number is only used",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := NewSettings(tc.tracker, tc.repo, tc.project)
			if err == nil {
				t.Fatalf("NewSettings = %+v, want an error", value)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("NewSettings error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}
