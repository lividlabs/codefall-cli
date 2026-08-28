package setup_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// schemas/settings.schema.json is the published definition of .codefall/settings.json, and setup is
// what writes that file. This test holds the two equal from setup's side, as doctor's holds them
// equal from the side that reads it — a file setup wrote and doctor rejected would be the worst
// possible bug in either component.
//
// It lives in the facade package rather than beside the domain because depguard denies os in domain
// tests and //go:embed cannot reach outside its own package directory. `go test` runs with the
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

func TestSchemaIdentityMatchesWhatSetupWrites(t *testing.T) {
	schema := loadSchema(t)

	if got := text(t, schema, "$id"); got != domain.SettingsSchemaID {
		t.Errorf("$id = %q, want %q", got, domain.SettingsSchemaID)
	}
}

func TestSchemaMatchesTheDomainFields(t *testing.T) {
	properties := object(t, loadSchema(t), "properties")

	if got := number(t, object(t, properties, "version"), "const"); got != domain.SettingsVersion {
		t.Errorf("properties.version.const = %v, want %d", got, domain.SettingsVersion)
	}

	if got, want := list(t, object(t, properties, "tracker"), "enum"), domain.Trackers(); !slices.Equal(got, want) {
		t.Errorf("properties.tracker.enum = %q, want %q", got, want)
	}

	blockProperties := object(t, object(t, properties, domain.TrackerGitHub), "properties")

	if got := text(t, object(t, blockProperties, "repo"), "pattern"); got != domain.RepoPattern {
		t.Errorf("properties.%s.properties.repo.pattern = %q, want %q", domain.TrackerGitHub, got, domain.RepoPattern)
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
