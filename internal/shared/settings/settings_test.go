package settings

import (
	"slices"
	"strings"
	"testing"
)

// complete returns the settings from the schema's example, which every case here mutates.
func complete() Document {
	return Document{
		"$schema": SchemaID,
		"version": 1.0,
		"tracker": "github",
		"github": map[string]any{
			"repo":    "lividlabs/codefall-cli",
			"project": 3.0,
		},
	}
}

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name string
		doc  Document
		want []string
	}{
		{
			name: "complete",
			doc:  complete(),
		},
		{
			name: "complete without the optional fields",
			doc: Document{
				"version": 1.0,
				"tracker": "github",
				"github":  map[string]any{"repo": "lividlabs/codefall-cli"},
			},
		},
		{
			name: "unknown top-level keys are ignored",
			doc: Document{
				"version":  1.0,
				"tracker":  "github",
				"github":   map[string]any{"repo": "a/b"},
				"nonsense": "ignored",
			},
		},
		{
			name: "a block for a tracker nobody knows about is ignored",
			doc: Document{
				"version": 1.0,
				"tracker": "github",
				"github":  map[string]any{"repo": "a/b"},
				"gitlab":  map[string]any{"project": "x"},
			},
		},
		{
			name: "empty document",
			doc:  Document{},
			want: []string{"version: missing", "tracker: missing"},
		},
		{
			name: "version missing",
			doc:  without(complete(), "version"),
			want: []string{"version: missing"},
		},
		{
			name: "version empty string counts as missing",
			doc:  with(complete(), "version", ""),
			want: []string{"version: missing"},
		},
		{
			name: "version is another number",
			doc:  with(complete(), "version", 2.0),
			want: []string{"version: must be 1"},
		},
		{
			name: "version is a string",
			doc:  with(complete(), "version", "1"),
			want: []string{"version: must be 1"},
		},
		{
			name: "version is fractional",
			doc:  with(complete(), "version", 1.5),
			want: []string{"version: must be 1"},
		},
		{
			name: "tracker missing suppresses the block problems",
			doc:  without(complete(), "tracker"),
			want: []string{"tracker: missing"},
		},
		{
			name: "tracker is not a string",
			doc:  with(complete(), "tracker", 42.0),
			want: []string{"tracker: must be a string"},
		},
		{
			name: "unknown tracker suppresses the block problems",
			doc:  without(with(complete(), "tracker", "gitlab"), "github"),
			want: []string{`tracker: unknown value "gitlab" (expected "beads", "github")`},
		},
		{
			name: "beads tracker with an empty beads block is complete",
			doc: Document{
				"version": 1.0,
				"tracker": "beads",
				"beads":   map[string]any{},
			},
		},
		{
			name: "beads block missing",
			doc: Document{
				"version": 1.0,
				"tracker": "beads",
			},
			want: []string{`beads: missing (required when tracker is "beads")`},
		},
		{
			name: "github block present while tracker is beads",
			doc: Document{
				"version": 1.0,
				"tracker": "beads",
				"beads":   map[string]any{},
				"github":  map[string]any{"repo": "a/b"},
			},
			want: []string{`github: present but tracker is "beads" — remove it`},
		},
		{
			name: "block missing",
			doc:  without(complete(), "github"),
			want: []string{`github: missing (required when tracker is "github")`},
		},
		{
			name: "block is not an object",
			doc:  with(complete(), "github", "lividlabs/codefall-cli"),
			want: []string{"github: must be an object"},
		},
		{
			name: "repo missing",
			doc:  with(complete(), "github", map[string]any{"project": 3.0}),
			want: []string{"github.repo: missing"},
		},
		{
			name: "repo empty counts as missing",
			doc:  with(complete(), "github", map[string]any{"repo": ""}),
			want: []string{"github.repo: missing"},
		},
		{
			name: "repo without a slash",
			doc:  with(complete(), "github", map[string]any{"repo": "no-slash"}),
			want: []string{"github.repo: must match owner/name"},
		},
		{
			name: "repo with two slashes",
			doc:  with(complete(), "github", map[string]any{"repo": "a/b/c"}),
			want: []string{"github.repo: must match owner/name"},
		},
		{
			name: "repo is not a string",
			doc:  with(complete(), "github", map[string]any{"repo": 3.0}),
			want: []string{"github.repo: must be a string"},
		},
		{
			name: "project is zero",
			doc:  with(complete(), "github", map[string]any{"repo": "a/b", "project": 0.0}),
			want: []string{"github.project: must be a positive integer"},
		},
		{
			name: "project is fractional",
			doc:  with(complete(), "github", map[string]any{"repo": "a/b", "project": 1.5}),
			want: []string{"github.project: must be a positive integer"},
		},
		{
			name: "project is a string",
			doc:  with(complete(), "github", map[string]any{"repo": "a/b", "project": "3"}),
			want: []string{"github.project: must be a positive integer"},
		},
		{
			name: "$schema is not a string",
			doc:  with(complete(), "$schema", 1.0),
			want: []string{"$schema: must be a string"},
		},
		{
			name: "every problem is reported at once, in definition order",
			doc: Document{
				"$schema": 1.0,
				"version": 2.0,
				"tracker": "github",
				"github":  map[string]any{"repo": "no-slash", "project": 0.0},
			},
			want: []string{
				"$schema: must be a string",
				"version: must be 1",
				"github.repo: must match owner/name",
				"github.project: must be a positive integer",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Validate(tc.doc); !slices.Equal(got, tc.want) {
				t.Errorf("Validate() = %q, want %q", got, tc.want)
			}
		})
	}
}

// A block for a tracker that is not selected is a problem once that tracker is known — beads and
// github already prove this pairwise above. This test adds a third, synthetic tracker for its
// length to show the rule keeps generalizing, which is also the proof that adding a row is all a
// further tracker needs.
func TestValidateRejectsAnotherKnownTrackersBlock(t *testing.T) {
	trackerFields["fake"] = []fieldSpec{{"token", true, isString}}
	t.Cleanup(func() { delete(trackerFields, "fake") })

	doc := with(complete(), "fake", map[string]any{"token": "x"})

	want := []string{`fake: present but tracker is "github" — remove it`}
	if got := Validate(doc); !slices.Equal(got, want) {
		t.Errorf("Validate() = %q, want %q", got, want)
	}
}

func TestTrackers(t *testing.T) {
	if got, want := Trackers(), []string{TrackerBeads, TrackerGitHub}; !slices.Equal(got, want) {
		t.Errorf("Trackers() = %q, want %q", got, want)
	}

	// The caller gets a copy: mutating it must not change what the next caller sees. Order is not
	// cosmetic — the schema test holds the schema's tracker enum equal to this list.
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

func TestValidateProject(t *testing.T) {
	for _, number := range []int{1, 3, 4096} {
		if err := ValidateProject(number); err != nil {
			t.Errorf("ValidateProject(%d) = %v, want nil", number, err)
		}
	}

	for _, number := range []int{0, -1} {
		if err := ValidateProject(number); err == nil {
			t.Errorf("ValidateProject(%d) = nil, want an error", number)
		}
	}
}

func TestRequiredFields(t *testing.T) {
	if got, want := RequiredFields(), []string{"version", "tracker"}; !slices.Equal(got, want) {
		t.Errorf("RequiredFields() = %q, want %q", got, want)
	}

	if got, want := RequiredTrackerFields(TrackerBeads), []string{}; !slices.Equal(got, want) {
		t.Errorf("RequiredTrackerFields(%q) = %q, want %q", TrackerBeads, got, want)
	}

	if got, want := RequiredTrackerFields(TrackerGitHub), []string{"repo"}; !slices.Equal(got, want) {
		t.Errorf("RequiredTrackerFields(%q) = %q, want %q", TrackerGitHub, got, want)
	}

	if got := RequiredTrackerFields("nonesuch"); len(got) != 0 {
		t.Errorf("RequiredTrackerFields(%q) = %q, want none", "nonesuch", got)
	}
}

func with(doc Document, key string, value any) Document {
	doc[key] = value

	return doc
}

func without(doc Document, key string) Document {
	delete(doc, key)

	return doc
}
