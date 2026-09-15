package domain

// Step identifies one unit of work a create run performs. The ID is stable and machine-readable; the
// Title is what a person reads while the step runs.
type Step struct {
	ID    string
	Title string
}

// The steps a create run performs, in the order it performs them. The initial commit is the last of
// them and comes before init runs, because init's preflight refuses a .gitignore that is not
// committed: bd init would otherwise take it into a commit of its own. The push is not part of the
// run — it is offered once init has finished, so what it pushes includes bd init's commit.
var (
	DirectoryStep  = Step{ID: "directory", Title: "Creating the project directory"}
	RepositoryStep = Step{ID: "repository", Title: "Initializing the git repository"}
	RemoteStep     = Step{ID: "remote", Title: "Adding the remote"}
	FilesStep      = Step{ID: "files", Title: "Writing README.md and .gitignore"}
	CommitStep     = Step{ID: "commit", Title: "Making the initial commit"}
	PushStep       = Step{ID: "push", Title: "Pushing to the remote"}
)

// Outcome is what a step did. A step that could not do its work returns an error instead: there is
// no failed outcome, because a failure stops the run.
type Outcome int

// The two outcomes, in the order a reader is most likely to see them.
const (
	OutcomeDone Outcome = iota
	OutcomeSkipped
)

// StepResult is what one step did and the sentence that says so.
type StepResult struct {
	Step    Step
	Outcome Outcome
	Detail  string
}

// Done records a step that did its work.
func (s Step) Done(detail string) StepResult {
	return StepResult{Step: s, Outcome: OutcomeDone, Detail: detail}
}

// Skipped records a step that had nothing to do and left it alone.
func (s Step) Skipped(detail string) StepResult {
	return StepResult{Step: s, Outcome: OutcomeSkipped, Detail: detail}
}

// Report is the ordered outcome of one create run. A step that never ran, because an earlier one
// failed, is absent from the report.
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
