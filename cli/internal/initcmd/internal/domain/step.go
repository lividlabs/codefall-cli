package domain

// Step identifies one unit of work an init run performs. The ID is stable and machine-readable; the
// Title is what a person reads while the step runs.
type Step struct {
	ID    string
	Title string
}

// The steps an init run performs, in the order it performs them. The settings file comes first, the
// harness plugin second, Beads third, the session hook that tells the harness about Beads fourth,
// and the section that tells an agent how to use it last. The hook step follows the plugin step
// because both of them write the harness's settings file, and the harness CLI's own merge runs
// first. The agents step follows Beads because bd init commits what it staged, and codefall's edit
// to AGENTS.md belongs in the author's own commit rather than bd's.
var (
	SettingsStep = Step{ID: "settings", Title: "Writing .codefall/settings.json"}
	PluginStep   = Step{ID: "plugin", Title: "Installing the codefall plugin"}
	BeadsStep    = Step{ID: "beads", Title: "Initializing Beads"}
	HookStep     = Step{ID: "hook", Title: "Adding the Beads session hook"}
	AgentsStep   = Step{ID: "agents", Title: "Writing the Beads section to AGENTS.md"}
)

// What codefall's plugin for Claude Code is called and where it comes from. These are facts about
// codefall itself, which is why they live here; how the harness CLI is asked to install them is the
// application layer's business.
//
// A plugin is identified as name@marketplace, so PluginID is the plugin's own name joined to
// MarketplaceName — written out rather than composed, because it is the string the harness prints
// and the string a person types.
const (
	PluginID          = "codefall@codefall"
	MarketplaceName   = "codefall"
	MarketplaceSource = "lividlabs/codefall-plugin"

	// PluginVersion is the plugin release the skills-directory harnesses are fetched from. It is
	// pinned rather than latest so that every run installs the same release until a CLI release
	// deliberately moves it; the --plugin-version flag overrides it for one run.
	PluginVersion = "0.7.0"
)

// The session hook codefall installs so that a harness session starts knowing about the project's
// Beads database. These are facts about what codefall installs, like the plugin's identifiers above;
// the file the hook is written into, and the shape that file wants, belong to the application layer.
const (
	BeadsHookEvent   = "SessionStart"
	BeadsHookCommand = "bd prime --hook-json"
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
