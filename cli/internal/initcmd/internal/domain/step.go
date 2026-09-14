package domain

// Step identifies one unit of work an init run performs. The ID is stable and machine-readable; the
// Title is what a person reads while the step runs.
type Step struct {
	ID    string
	Title string
}

// The steps an init run performs, in the order it performs them. The settings file comes first, the
// harness extension second, Beads third, the hooks that tell the harness about Beads and guard the
// default branch fourth, and the section that tells an agent how to use it fifth. The hook step
// follows the extension step because the scripts the hooks run are what the extension step copies.
// The agents step follows Beads because bd init commits what it staged, and codefall's edit to
// AGENTS.md belongs in the author's own commit rather than bd's — which is the same reason the
// ignore step comes last: .ignore is the author's file to commit, not bd's.
var (
	SettingsStep  = Step{ID: "settings", Title: "Writing .codefall/settings.json"}
	ExtensionStep = Step{ID: "extension", Title: "Installing the codefall extension"}
	BeadsStep     = Step{ID: "beads", Title: "Initializing Beads"}
	HookStep      = Step{ID: "hook", Title: "Registering codefall's hooks"}
	AgentsStep    = Step{ID: "agents", Title: "Writing the Beads section to AGENTS.md"}
	IgnoreStep    = Step{ID: "ignore", Title: "Hiding review findings from codebase search"}
)

// What codefall's extension for Claude Code is called and where it comes from. These are facts about
// codefall itself, which is why they live here; how the harness CLI is asked to install them is the
// application layer's business.
const (
	// ManifestName is the install record the extension step writes, so upgrade and drift checks can
	// know exactly which files belong to it. It lives in .codefall/, next to settings.json.
	ManifestName = ".codefall/manifest.json"
)

// Outcome is what a step did. A step that could not do its work returns an error instead: there is
// no failed outcome, because a failure stops the run.
//
// It has no String: presentation prints a mark rather than a word, so a name for each outcome would
// have nothing but a test to read it.
type Outcome int

// The two outcomes, in the order a reader is most likely to see them.
const (
	OutcomeDone Outcome = iota
	OutcomeSkipped
)

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
