package domain

import "testing"

func TestOutcomeString(t *testing.T) {
	for _, tc := range []struct {
		outcome Outcome
		want    string
	}{
		{OutcomeDone, "DONE"},
		{OutcomeSkipped, "SKIPPED"},
		{Outcome(42), "UNKNOWN"},
	} {
		if got := tc.outcome.String(); got != tc.want {
			t.Errorf("Outcome(%d).String() = %q, want %q", tc.outcome, got, tc.want)
		}
	}
}

func TestStepResults(t *testing.T) {
	done := SettingsStep.Done("wrote .codefall/settings.json")
	if done.Outcome != OutcomeDone || done.Step != SettingsStep || done.Detail != "wrote .codefall/settings.json" {
		t.Errorf("Done() = %+v, want the step, DONE, and the detail", done)
	}

	skipped := SettingsStep.Skipped("already exists")
	if skipped.Outcome != OutcomeSkipped || skipped.Detail != "already exists" {
		t.Errorf("Skipped() = %+v, want SKIPPED and the detail", skipped)
	}
}

func TestReportKeepsStepOrder(t *testing.T) {
	first := SettingsStep.Done("one")
	second := Step{ID: "later", Title: "Later"}.Skipped("two")

	results := NewReport(first, second).Results()
	if len(results) != 2 || results[0] != first || results[1] != second {
		t.Errorf("Results() = %+v, want the results in the order they were given", results)
	}
}

func TestZeroReportHasNoResults(t *testing.T) {
	if got := (Report{}).Results(); len(got) != 0 {
		t.Errorf("Report{}.Results() = %+v, want none", got)
	}
}
