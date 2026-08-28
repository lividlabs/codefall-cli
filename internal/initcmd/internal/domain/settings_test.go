package domain

import (
	"slices"
	"strings"
	"testing"

	"github.com/samber/mo"
)

func TestTrackersAndHarnessesAreSortedAndCopied(t *testing.T) {
	if got, want := Trackers(), []string{TrackerBeads, TrackerGitHub}; !slices.Equal(got, want) {
		t.Errorf("Trackers() = %q, want %q", got, want)
	}

	if got, want := Harnesses(), []string{HarnessClaudeCode}; !slices.Equal(got, want) {
		t.Errorf("Harnesses() = %q, want %q", got, want)
	}

	// The caller gets a copy: mutating it must not change what the next caller sees.
	Trackers()[0] = "mutated"

	if got := Trackers()[0]; got != TrackerBeads {
		t.Errorf("Trackers()[0] after a caller mutated its copy = %q, want %q", got, TrackerBeads)
	}
}

func TestParseTracker(t *testing.T) {
	for _, tracker := range Trackers() {
		if got, err := ParseTracker(tracker); err != nil || got != tracker {
			t.Errorf("ParseTracker(%q) = %q, %v, want %q, nil", tracker, got, err, tracker)
		}
	}

	_, err := ParseTracker("jira")
	if err == nil {
		t.Fatal(`ParseTracker("jira") = nil error, want an error`)
	}

	// The message lists what would have worked, because that is what the reader needs next.
	for _, want := range append([]string{"jira"}, Trackers()...) {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ParseTracker(%q) error = %q, want it to mention %q", "jira", err, want)
		}
	}
}

func TestParseHarness(t *testing.T) {
	if got, err := ParseHarness(HarnessClaudeCode); err != nil || got != HarnessClaudeCode {
		t.Errorf("ParseHarness(%q) = %q, %v, want %q, nil", HarnessClaudeCode, got, err, HarnessClaudeCode)
	}

	err := errorFrom(ParseHarness("codex"))

	want := `harness "codex" is not supported yet (supported: claude-code)`
	if err == nil || err.Error() != want {
		t.Errorf("ParseHarness(%q) error = %v, want %q", "codex", err, want)
	}
}

func TestNewSettingsForGitHub(t *testing.T) {
	settings, err := NewSettings(TrackerGitHub, mo.Some("lividlabs/codefall-cli"), mo.Some(3))
	if err != nil {
		t.Fatalf("NewSettings: %v", err)
	}

	if settings.Tracker != TrackerGitHub {
		t.Errorf("Tracker = %q, want %q", settings.Tracker, TrackerGitHub)
	}

	github, ok := settings.GitHub.Get()
	if !ok {
		t.Fatalf("GitHub = %v, want a block", settings.GitHub)
	}

	if github.Repo != "lividlabs/codefall-cli" {
		t.Errorf("GitHub.Repo = %q, want %q", github.Repo, "lividlabs/codefall-cli")
	}

	if got, ok := github.Project.Get(); !ok || got != 3 {
		t.Errorf("GitHub.Project = %v, want Some(3)", github.Project)
	}
}

func TestNewSettingsForGitHubWithoutAProject(t *testing.T) {
	settings, err := NewSettings(TrackerGitHub, mo.Some("owner/name"), mo.None[int]())
	if err != nil {
		t.Fatalf("NewSettings: %v", err)
	}

	github, _ := settings.GitHub.Get()
	if github.Project.IsPresent() {
		t.Errorf("GitHub.Project = %v, want None", github.Project)
	}
}

func TestNewSettingsForBeadsCarriesNoGitHubBlock(t *testing.T) {
	settings, err := NewSettings(TrackerBeads, mo.None[string](), mo.None[int]())
	if err != nil {
		t.Fatalf("NewSettings: %v", err)
	}

	if settings.Tracker != TrackerBeads {
		t.Errorf("Tracker = %q, want %q", settings.Tracker, TrackerBeads)
	}

	if settings.GitHub.IsPresent() {
		t.Errorf("GitHub = %v, want None", settings.GitHub)
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
			tracker: TrackerGitHub,
			repo:    mo.None[string](),
			project: mo.None[int](),
			want:    "needs a repository",
		},
		{
			name:    "a repository that is not owner/name",
			tracker: TrackerGitHub,
			repo:    mo.Some("codefall-cli"),
			project: mo.None[int](),
			want:    "owner/name",
		},
		{
			name:    "a project number below one",
			tracker: TrackerGitHub,
			repo:    mo.Some("owner/name"),
			project: mo.Some(0),
			want:    "positive integer",
		},
		{
			name:    "a repository under beads",
			tracker: TrackerBeads,
			repo:    mo.Some("owner/name"),
			project: mo.None[int](),
			want:    "a repository is only used",
		},
		{
			name:    "a project number under beads",
			tracker: TrackerBeads,
			repo:    mo.None[string](),
			project: mo.Some(3),
			want:    "a project number is only used",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings, err := NewSettings(tc.tracker, tc.repo, tc.project)
			if err == nil {
				t.Fatalf("NewSettings = %+v, want an error", settings)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("NewSettings error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestValidateRepo(t *testing.T) {
	for _, repo := range []string{"owner/name", "lividlabs/codefall-cli", "a_b.c-d/e.f_g-h"} {
		if err := ValidateRepo(repo); err != nil {
			t.Errorf("ValidateRepo(%q) = %v, want nil", repo, err)
		}
	}

	for _, repo := range []string{"", "name", "owner/name/extra", "owner /name", "owner/na me"} {
		if err := ValidateRepo(repo); err == nil {
			t.Errorf("ValidateRepo(%q) = nil, want an error", repo)
		}
	}
}

// errorFrom drops the value of a (T, error) pair so a table can assert on the error alone.
func errorFrom(_ string, err error) error {
	return err
}
