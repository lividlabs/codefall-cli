// Package domain holds doctor's entities and value objects: what a check is, what its outcome looks
// like, and the rules for reading settings and token scopes. It performs no IO and names no
// delivery or infrastructure type (ADR-BASE-01).
package domain

import "github.com/samber/mo"

// Status is the outcome of a single check. A WARN is a condition worth reporting that does not stop
// codefall from working; a FAIL does.
type Status int

// The three outcomes, ordered from healthiest to worst.
const (
	StatusPass Status = iota
	StatusWarn
	StatusFail
)

// String returns the fixed-width word the report prints for the status.
func (s Status) String() string {
	switch s {
	case StatusPass:
		return "PASS"
	case StatusWarn:
		return "WARN"
	case StatusFail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// Check identifies one thing doctor looks at. The ID is stable and machine-readable; the Title is
// what a person reads.
type Check struct {
	ID    string
	Title string
}

// Result is the outcome of running one check. Detail is absent on a bare PASS, and Remedy is absent
// whenever nothing useful can be suggested — doctor never invents a fix it does not believe in.
type Result struct {
	Check  Check
	Status Status
	Detail mo.Option[string]
	Remedy mo.Option[string]
}

// Pass records a check that succeeded with nothing worth saying about it.
func (c Check) Pass() Result {
	return Result{Check: c, Status: StatusPass, Detail: mo.None[string](), Remedy: mo.None[string]()}
}

// PassWithDetail records a check that succeeded and has something to report, such as a version.
func (c Check) PassWithDetail(detail string) Result {
	return Result{Check: c, Status: StatusPass, Detail: mo.Some(detail), Remedy: mo.None[string]()}
}

// Warn records a condition that does not stop codefall from working.
func (c Check) Warn(detail string, remedy mo.Option[string]) Result {
	return Result{Check: c, Status: StatusWarn, Detail: mo.Some(detail), Remedy: remedy}
}

// Fail records a condition that does stop codefall from working.
func (c Check) Fail(detail string, remedy mo.Option[string]) Result {
	return Result{Check: c, Status: StatusFail, Detail: mo.Some(detail), Remedy: remedy}
}

// The nine checks doctor runs, in the order it runs them.
var (
	CodefallDir      = Check{ID: "codefall-dir", Title: ".codefall/ exists"}
	SettingsFile     = Check{ID: "settings-file", Title: ".codefall/settings.json exists"}
	SettingsJSON     = Check{ID: "settings-json", Title: "settings.json is valid JSON"}
	SettingsComplete = Check{ID: "settings-complete", Title: "settings.json is complete"}
	BeadsInstalled   = Check{ID: "bd-installed", Title: "bd is on PATH"}
	BeadsInitialized = Check{ID: "beads-initialized", Title: "Beads is initialized here"}
	GHInstalled      = Check{ID: "gh-installed", Title: "gh is on PATH"}
	GHAuthenticated  = Check{ID: "gh-auth", Title: "gh is logged in to github.com"}
	GHScopes         = Check{ID: "gh-scopes", Title: "gh token has the required scopes"}
)
