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
	BlockReview   = "review"
)

// The .ignore file, which is not settings but is the other file init writes and doctor checks — so
// it lives here for the same reason the settings format does (ADR-003).
//
// codefall-review commits its findings so that patterns across reviews stay visible, which puts old
// findings in the same tree as the code. Ripgrep reads .ignore and git does not, so one line keeps
// them out of every search that goes through ripgrep while leaving them tracked.
const (
	// IgnoreName is the file, beside .codefall/ rather than at the repository root: a run below the
	// root installs there, and ripgrep reads .ignore files down the tree.
	IgnoreName = ".ignore"
	// IgnoreEntry is the line itself.
	IgnoreEntry = ".codefall/reviews/"
	// IgnoreComment says why the line is there, for whoever finds the file later.
	IgnoreComment = "# codefall review findings: tracked in git, skipped by ripgrep."
)

// IgnoresReviews reports whether a .ignore file's contents already name the reviews directory.
//
// The comparison is per line and ignores surrounding space, so an entry someone indented still
// counts and a commented-out one does not — ripgrep does not read the comment either. Both
// components need the same answer: init decides whether to append, doctor decides whether to fail.
func IgnoresReviews(body string) bool {
	for line := range strings.SplitSeq(body, "\n") {
		if strings.TrimSpace(line) == IgnoreEntry {
			return true
		}
	}

	return false
}

var repoRegexp = regexp.MustCompile(RepoPattern)

// fieldSpec is one row of the settings shape. Check returns "" when the value is acceptable and the
// reason it is not otherwise.
type fieldSpec struct {
	Name     string
	Required bool
	Check    func(v any) string
}

// topLevelFields is the settings file's own shape. Definition order is report order.
//
// The review block is optional at the top level and complete when it is there: a project set up
// before the block existed is still valid settings, and one that carries it carries every field.
var topLevelFields = []fieldSpec{
	{"$schema", false, isString},
	{"version", true, isVersion},
	{"tracker", true, isTracker},
	{BlockReview, false, isObject},
}

// reviewFields is the shape of the review block, which codefall-review reads and nothing else
// writes. Posting findings to a pull request is visible to everyone on it, so the field exists to
// make that a decision the project made rather than a default it inherited.
var reviewFields = []fieldSpec{
	{"postToPullRequest", true, isBool},
}

// trackerFields is the shape of each known tracker's block, keyed by the value of "tracker" that
// selects it. Adding a tracker is one row here plus one oneOf branch in the schema; the
// "exactly one tracker block" rule then applies to it without further code.
var trackerFields = map[string][]fieldSpec{
	// bd keeps its own configuration under .beads/, so the block carries no fields of its own today;
	// it exists so the "exactly one tracker block" rule applies to it uniformly.
	TrackerBeads: {},
	// The fields say what they point at rather than which forge holds it: the block is already named
	// github, and "repo" alone read as the repository being set up rather than the one issues are
	// filed in — which need not be the same repository.
	TrackerGitHub: {
		{"issuesRepo", true, matches(repoRegexp, "owner/name")},
		{"issuesProject", false, isPositiveInteger},
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

// RequiredReviewFields returns the fields the review block must carry once it is present, in
// definition order.
func RequiredReviewFields() []string {
	return requiredNames(reviewFields)
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

	// The review block is independent of the tracker, so it is checked before the tracker's own
	// block decides whether there is anything further to say.
	if _, present := lookup(doc, BlockReview); present {
		problems = append(problems, validateBlock(doc, BlockReview, reviewFields)...)
	}

	// A missing or unknown tracker selects no block, so there is nothing further to say.
	if _, known := trackerFields[selected]; !known {
		return problems
	}

	if _, present := lookup(doc, selected); !present {
		problems = append(problems,
			fmt.Sprintf("%s: missing (required when tracker is %q)", selected, selected))

		return problems
	}

	problems = append(problems, validateBlock(doc, selected, trackerFields[selected])...)

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

// validateBlock checks one present block against its field table. The tracker's block and the
// review block have the same shape — a name, an object, a table of fields — so they are one
// function. Whether a block being absent is a problem differs between them and is decided by the
// caller, which is why this one is only ever called on a block that is there.
func validateBlock(doc Document, name string, fields []fieldSpec) []string {
	var problems []string

	value, ok := lookup(doc, name)
	if !ok {
		return nil
	}

	block, ok := value.(map[string]any)
	if !ok {
		return []string{name + ": must be an object"}
	}

	for _, field := range fields {
		path := name + "." + field.Name

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

func isBool(v any) string {
	if _, ok := v.(bool); !ok {
		return "must be true or false"
	}

	return ""
}

func isObject(v any) string {
	if _, ok := v.(map[string]any); !ok {
		return "must be an object"
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
