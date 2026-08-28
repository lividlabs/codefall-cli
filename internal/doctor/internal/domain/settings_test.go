package domain

import (
	"slices"
	"testing"
)

// complete returns the settings from the schema's example, which every case here mutates.
func complete() Document {
	return Document{
		"$schema": SettingsSchemaID,
		"version": 1.0,
		"tracker": "github",
		"github": map[string]any{
			"repo":    "lividlabs/codefall-cli",
			"project": 3.0,
		},
	}
}

func TestValidateSettings(t *testing.T) {
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
			if got := ValidateSettings(tc.doc); !slices.Equal(got, tc.want) {
				t.Errorf("ValidateSettings() = %q, want %q", got, tc.want)
			}
		})
	}
}

// A block for a tracker that is not selected is a problem once that tracker is known — beads and
// github already prove this pairwise above. This test adds a third, synthetic tracker for its
// length to show the rule keeps generalizing, which is also the proof that adding a row is all a
// further tracker needs.
func TestValidateSettingsRejectsAnotherKnownTrackersBlock(t *testing.T) {
	trackerFields["fake"] = []fieldSpec{{"token", true, isString}}
	t.Cleanup(func() { delete(trackerFields, "fake") })

	doc := with(complete(), "fake", map[string]any{"token": "x"})

	want := []string{`fake: present but tracker is "github" — remove it`}
	if got := ValidateSettings(doc); !slices.Equal(got, want) {
		t.Errorf("ValidateSettings() = %q, want %q", got, want)
	}
}

func TestTrackers(t *testing.T) {
	if got, want := Trackers(), []string{TrackerBeads, TrackerGitHub}; !slices.Equal(got, want) {
		t.Errorf("Trackers() = %q, want %q", got, want)
	}
}

func TestRequiredFields(t *testing.T) {
	if got, want := RequiredSettingsFields(), []string{"version", "tracker"}; !slices.Equal(got, want) {
		t.Errorf("RequiredSettingsFields() = %q, want %q", got, want)
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
