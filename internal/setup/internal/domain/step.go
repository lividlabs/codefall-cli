package domain

// Step identifies one unit of work an init run performs. The ID is stable and machine-readable; the
// Title is what a person reads while the step runs.
type Step struct {
	ID    string
	Title string
}

// The steps an init run performs, in the order it performs them. Writing the settings file is the
// first; installing the Claude Code plugin and initialising Beads follow it.
var SettingsStep = Step{ID: "settings", Title: "Writing .codefall/settings.json"}

// Outcome is what a step did. A step that could not do its work returns an error instead: there is
// no failed outcome, because a failure stops the run.
type Outcome int

// The two outcomes, in the order a reader is most likely to see them.
const (
	OutcomeDone Outcome = iota
	OutcomeSkipped
)

// String returns the word the report prints for the outcome.
func (o Outcome) String() string {
	switch o {
	case OutcomeDone:
		return "DONE"
	case OutcomeSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

// StepResult is what one step did and the sentence that says so. Every result carries a detail:
// a step that ran has something to report, and a step that was skipped has a reason.
type StepResult struct {
	Step    Step
	Outcome Outcome
	Detail  string
}

// Done records a step that did its work.
func (s Step) Done(detail string) StepResult {
	return StepResult{Step: s, Outcome: OutcomeDone, Detail: detail}
}

// Skipped records a step that found its work already done and left it alone.
func (s Step) Skipped(detail string) StepResult {
	return StepResult{Step: s, Outcome: OutcomeSkipped, Detail: detail}
}

// Report is the ordered outcome of one init run. A step that never ran, because an earlier one
// failed, is absent from the report rather than present with an outcome of its own.
type Report struct {
	results []StepResult
}

// NewReport builds a report from results already in step order.
func NewReport(results ...StepResult) Report {
	return Report{results: results}
}

// Results returns the results in step order.
func (r Report) Results() []StepResult {
	return r.results
}
