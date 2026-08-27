package domain

import (
	"testing"

	"github.com/samber/mo"
)

func TestReportPreservesOrder(t *testing.T) {
	report := NewReport(
		CodefallDir.Pass(),
		SettingsFile.Fail("not found", mo.None[string]()),
		BeadsInitialized.Warn("no database", mo.Some("bd init")),
	)

	want := []string{CodefallDir.ID, SettingsFile.ID, BeadsInitialized.ID}

	results := report.Results()
	if len(results) != len(want) {
		t.Fatalf("Results() has %d entries, want %d", len(results), len(want))
	}

	for i, id := range want {
		if results[i].Check.ID != id {
			t.Errorf("Results()[%d].Check.ID = %q, want %q", i, results[i].Check.ID, id)
		}
	}
}

func TestReportFailedCountsOnlyFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		report Report
		want   int
	}{
		{"empty", NewReport(), 0},
		{"all pass", NewReport(CodefallDir.Pass(), SettingsFile.Pass()), 0},
		{"warnings do not count", NewReport(BeadsInitialized.Warn("x", mo.None[string]())), 0},
		{
			"two failures",
			NewReport(
				CodefallDir.Pass(),
				SettingsFile.Fail("not found", mo.None[string]()),
				BeadsInitialized.Warn("x", mo.None[string]()),
				GHScopes.Fail("missing repo", mo.None[string]()),
			),
			2,
		},
	} {
		if got := tc.report.Failed(); got != tc.want {
			t.Errorf("%s: Failed() = %d, want %d", tc.name, got, tc.want)
		}
	}
}
