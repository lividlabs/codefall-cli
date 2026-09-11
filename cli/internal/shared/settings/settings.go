// Package settings is the one definition of the .codefall/settings.json format: the constants the
// file carries, the tables of what each tracker's block may hold, and the rules that decide whether
// a document or a single field is acceptable. Both components need it — initcmd writes that file and
// doctor reads it — so it lives here rather than in either one's domain (ADR-003).
//
// It is a pure shared module: it imports the standard library and nothing else, which is what lets
// domain/ and application/ name it. Adding a dependency here breaks that permission and fails the
// pure-shared-modules rule in .golangci.yml.
//
// What belongs here is the format. How a component builds a settings value from a survey's answers,
// or reports on one it read, belongs to that component.
package settings

import (
	"fmt"
	"maps"
	"math"
	"regexp"
	"slices"
	"strings"
)

// Document is a decoded settings file: the generic shape a JSON decoder produces (strings, float64
// numbers, bools, nil, and nested maps). Validation reads that shape without knowing it came from
// JSON, so the field tables below are the one definition of what settings may contain.
type Document map[string]any

// The published constants of the settings format. The schema test in this package holds
// schemas/settings.schema.json equal to these, so the two definitions cannot drift apart silently.
const (
	Version       = 1
	TrackerBeads  = "beads"
	TrackerGitHub = "github"
	RepoPattern   = `^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`
	SchemaID      = "https://raw.githubusercontent.com/lividlabs/codefall-cli/main/schemas/settings.schema.json"
)

var repoRegexp = regexp.MustCompile(RepoPattern)

// fieldSpec is one row of the settings shape. Check returns "" when the value is acceptable and the
// reason it is not otherwise.
type fieldSpec struct {
	Name     string
	Required bool
	Check    func(v any) string
}

// topLevelFields is the settings file's own shape. Definition order is report order.
var topLevelFields = []fieldSpec{
	{"$schema", false, isString},
	{"version", true, isVersion},
	{"tracker", true, isTracker},
}

// trackerFields is the shape of each known tracker's block, keyed by the value of "tracker" that
// selects it. Adding a tracker is one row here plus one oneOf branch in the schema; the
// "exactly one tracker block" rule then applies to it without further code.
var trackerFields = map[string][]fieldSpec{
	// bd keeps its own configuration under .beads/, so the block carries no fields of its own today;
	// it exists so the "exactly one tracker block" rule applies to it uniformly.
	TrackerBeads: {},
	TrackerGitHub: {
		{"repo", true, matches(repoRegexp, "owner/name")},
		{"project", false, isPositiveInteger},
	},
}

// Trackers returns the known tracker names, sorted.
func Trackers() []string {
	return slices.Sorted(maps.Keys(trackerFields))
}

// RequiredFields returns the top-level fields settings must carry, in definition order.
func RequiredFields() []string {
	return requiredNames(topLevelFields)
}

// RequiredTrackerFields returns the fields the named tracker's block must carry, in definition
// order. An unknown tracker has none.
func RequiredTrackerFields(tracker string) []string {
	return requiredNames(trackerFields[tracker])
}

func requiredNames(fields []fieldSpec) []string {
	names := []string{}

	for _, field := range fields {
		if field.Required {
			names = append(names, field.Name)
		}
	}

	return names
}

// ParseTracker returns the tracker name when it is one codefall knows, and an error listing the
// known ones when it is not.
func ParseTracker(name string) (string, error) {
	if _, known := trackerFields[name]; known {
		return name, nil
	}

	return "", fmt.Errorf("unknown tracker %q (known trackers: %s)", name, strings.Join(Trackers(), ", "))
}

// ValidateRepo reports whether a repository is written the way GitHub names one.
//
// This and ValidateProject are the field-level halves of what Validate checks across a whole
// document. They are separate because a form validates a field as it is typed, before there are
// enough answers to have a document at all — and because what a person reads while typing says how
// to write the value, where a report says which path is wrong.
func ValidateRepo(repo string) error {
	if !repoRegexp.MatchString(repo) {
		return fmt.Errorf("repository %q must be written as owner/name", repo)
	}

	return nil
}

// ValidateProject reports whether a GitHub project number could identify a project.
func ValidateProject(number int) error {
	if number < 1 {
		return fmt.Errorf("project number %d must be a positive integer", number)
	}

	return nil
}

// Validate reports every problem with a settings document at once, each as "path: reason", in
// definition order. An empty result means the settings are complete.
func Validate(doc Document) []string {
	var problems []string

	selected := ""

	for _, field := range topLevelFields {
		value, ok := lookup(doc, field.Name)
		if !ok {
			if field.Required {
				problems = append(problems, field.Name+": missing")
			}

			continue
		}

		if reason := field.Check(value); reason != "" {
			problems = append(problems, field.Name+": "+reason)

			continue
		}

		if field.Name == "tracker" {
			selected, _ = value.(string)
		}
	}

	// A missing or unknown tracker selects no block, so there is nothing further to say.
	if _, known := trackerFields[selected]; !known {
		return problems
	}

	problems = append(problems, validateTrackerBlock(doc, selected)...)

	for _, other := range Trackers() {
		if other == selected {
			continue
		}

		if _, present := doc[other]; present {
			problems = append(problems, fmt.Sprintf("%s: present but tracker is %q — remove it", other, selected))
		}
	}

	return problems
}

func validateTrackerBlock(doc Document, tracker string) []string {
	var problems []string

	value, ok := lookup(doc, tracker)
	if !ok {
		return []string{fmt.Sprintf("%s: missing (required when tracker is %q)", tracker, tracker)}
	}

	block, ok := value.(map[string]any)
	if !ok {
		return []string{tracker + ": must be an object"}
	}

	for _, field := range trackerFields[tracker] {
		path := tracker + "." + field.Name

		fieldValue, present := lookup(block, field.Name)
		if !present {
			if field.Required {
				problems = append(problems, path+": missing")
			}

			continue
		}

		if reason := field.Check(fieldValue); reason != "" {
			problems = append(problems, path+": "+reason)
		}
	}

	return problems
}

// lookup reports whether a key carries a usable value. A null and an empty string both count as
// missing, so a half-filled template reads the same as an absent field.
func lookup(doc map[string]any, name string) (any, bool) {
	value, ok := doc[name]
	if !ok || value == nil {
		return nil, false
	}

	if text, isText := value.(string); isText && text == "" {
		return nil, false
	}

	return value, true
}

func isString(v any) string {
	if _, ok := v.(string); !ok {
		return "must be a string"
	}

	return ""
}

func isVersion(v any) string {
	number, ok := wholeNumber(v)
	if !ok || number != Version {
		return fmt.Sprintf("must be %d", Version)
	}

	return ""
}

func isTracker(v any) string {
	name, ok := v.(string)
	if !ok {
		return "must be a string"
	}

	if _, known := trackerFields[name]; !known {
		expected := make([]string, 0, len(trackerFields))
		for _, tracker := range Trackers() {
			expected = append(expected, fmt.Sprintf("%q", tracker))
		}

		return fmt.Sprintf("unknown value %q (expected %s)", name, strings.Join(expected, ", "))
	}

	return ""
}

func isPositiveInteger(v any) string {
	number, ok := wholeNumber(v)
	if !ok || number < 1 {
		return "must be a positive integer"
	}

	return ""
}

func matches(pattern *regexp.Regexp, shape string) func(v any) string {
	return func(v any) string {
		text, ok := v.(string)
		if !ok {
			return "must be a string"
		}

		if !pattern.MatchString(text) {
			return "must match " + shape
		}

		return ""
	}
}

// wholeNumber accepts a JSON number that has no fractional part. JSON has one number type, so an
// integer arrives as a float64 and 1.5 has to be rejected here rather than by the decoder.
func wholeNumber(v any) (int, bool) {
	number, ok := v.(float64)
	if !ok || number != math.Trunc(number) {
		return 0, false
	}

	return int(number), true
}
