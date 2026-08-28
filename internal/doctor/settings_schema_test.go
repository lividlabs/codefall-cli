package doctor_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/domain"
)

// schemas/settings.schema.json is the published definition of .codefall/settings.json; the domain
// validates the same shape by hand, so no schema library ships in the binary. This test holds the
// two equal — without it they would drift apart the first time a field is added on one side.
//
// It lives in the facade package rather than beside the validator because depguard denies os in
// domain tests and //go:embed cannot reach outside its own package directory; a copy of the schema
// under domain/ would be exactly the drift this test exists to catch. `go test` runs with the
// package directory as its working directory, so the relative path is stable.
const schemaPath = "../../schemas/settings.schema.json"

func loadSchema(t *testing.T) map[string]any {
	t.Helper()

	data, err := os.ReadFile(filepath.FromSlash(schemaPath))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("decode schema: %v", err)
	}

	return schema
}

func TestSchemaIdentity(t *testing.T) {
	schema := loadSchema(t)

	if got := text(t, schema, "$id"); got != domain.SettingsSchemaID {
		t.Errorf("$id = %q, want %q", got, domain.SettingsSchemaID)
	}

	const draft = "https://json-schema.org/draft/2020-12/schema"
	if got := text(t, schema, "$schema"); got != draft {
		t.Errorf("$schema = %q, want %q", got, draft)
	}
}

func TestSchemaMatchesTheDomainFields(t *testing.T) {
	schema := loadSchema(t)

	if got, want := list(t, schema, "required"), domain.RequiredSettingsFields(); !slices.Equal(got, want) {
		t.Errorf("required = %q, want %q", got, want)
	}

	properties := object(t, schema, "properties")

	if got := number(t, object(t, properties, "version"), "const"); got != domain.SettingsVersion {
		t.Errorf("properties.version.const = %v, want %d", got, domain.SettingsVersion)
	}

	if got, want := list(t, object(t, properties, "tracker"), "enum"), domain.Trackers(); !slices.Equal(got, want) {
		t.Errorf("properties.tracker.enum = %q, want %q", got, want)
	}

	beadsBlock := object(t, properties, domain.TrackerBeads)

	if got, want := requiredList(t, beadsBlock, "required"), domain.RequiredTrackerFields(domain.TrackerBeads); !slices.Equal(got, want) {
		t.Errorf("properties.%s.required = %q, want %q", domain.TrackerBeads, got, want)
	}

	block := object(t, properties, domain.TrackerGitHub)

	if got, want := list(t, block, "required"), domain.RequiredTrackerFields(domain.TrackerGitHub); !slices.Equal(got, want) {
		t.Errorf("properties.%s.required = %q, want %q", domain.TrackerGitHub, got, want)
	}

	blockProperties := object(t, block, "properties")

	if got := text(t, object(t, blockProperties, "repo"), "pattern"); got != domain.RepoPattern {
		t.Errorf("properties.%s.properties.repo.pattern = %q, want %q", domain.TrackerGitHub, got, domain.RepoPattern)
	}

	if got := number(t, object(t, blockProperties, "project"), "minimum"); got != 1 {
		t.Errorf("properties.%s.properties.project.minimum = %v, want 1", domain.TrackerGitHub, got)
	}
}

// One oneOf branch per tracker, each pinning "tracker" to itself, requiring its own block, and
// forbidding every other tracker's block — which is what makes "exactly one tracker block" hold in
// the schema as well as in the validator. With one tracker the forbidding loop is empty.
func TestSchemaHasOneBranchPerTracker(t *testing.T) {
	schema := loadSchema(t)
	trackers := domain.Trackers()

	branches, ok := schema["oneOf"].([]any)
	if !ok {
		t.Fatalf("oneOf is %T, want a list", schema["oneOf"])
	}

	if len(branches) != len(trackers) {
		t.Fatalf("oneOf has %d branches, want %d (one per tracker)", len(branches), len(trackers))
	}

	for i, tracker := range trackers {
		branch, ok := branches[i].(map[string]any)
		if !ok {
			t.Fatalf("oneOf[%d] is %T, want an object", i, branches[i])
		}

		properties := object(t, branch, "properties")

		if got := text(t, object(t, properties, "tracker"), "const"); got != tracker {
			t.Errorf("oneOf[%d].properties.tracker.const = %q, want %q", i, got, tracker)
		}

		required := list(t, branch, "required")
		for _, want := range []string{"tracker", tracker} {
			if !slices.Contains(required, want) {
				t.Errorf("oneOf[%d].required = %q, want it to contain %q", i, required, want)
			}
		}

		for _, other := range trackers {
			if other == tracker {
				continue
			}

			if forbidden, ok := properties[other].(bool); !ok || forbidden {
				t.Errorf("oneOf[%d].properties.%s = %v, want false", i, other, properties[other])
			}
		}
	}
}

func object(t *testing.T, parent map[string]any, key string) map[string]any {
	t.Helper()

	value, ok := parent[key].(map[string]any)
	if !ok {
		t.Fatalf("%q is %T, want an object", key, parent[key])
	}

	return value
}

func text(t *testing.T, parent map[string]any, key string) string {
	t.Helper()

	value, ok := parent[key].(string)
	if !ok {
		t.Fatalf("%q is %T, want a string", key, parent[key])
	}

	return value
}

func number(t *testing.T, parent map[string]any, key string) int {
	t.Helper()

	value, ok := parent[key].(float64)
	if !ok {
		t.Fatalf("%q is %T, want a number", key, parent[key])
	}

	return int(value)
}

func list(t *testing.T, parent map[string]any, key string) []string {
	t.Helper()

	raw, ok := parent[key].([]any)
	if !ok {
		t.Fatalf("%q is %T, want a list", key, parent[key])
	}

	items := make([]string, 0, len(raw))

	for i, item := range raw {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("%q[%d] is %T, want a string", key, i, item)
		}

		items = append(items, text)
	}

	return items
}

// requiredList behaves like list, but a block that omits "required" entirely — rather than writing
// an empty array — is treated as requiring nothing. The beads block has no fields today, so it omits
// the key; this keeps that block's required-fields assertion pinned to the domain without forcing an
// empty "required": [] into the schema for a tracker that carries no fields.
func requiredList(t *testing.T, parent map[string]any, key string) []string {
	t.Helper()

	if _, present := parent[key]; !present {
		return []string{}
	}

	return list(t, parent, key)
}
