package domain

import (
	"testing"

	"github.com/samber/mo"
)

func TestStatusString(t *testing.T) {
	for _, tc := range []struct {
		status Status
		want   string
	}{
		{StatusPass, "PASS"},
		{StatusWarn, "WARN"},
		{StatusFail, "FAIL"},
		{Status(42), "UNKNOWN"},
	} {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestPassHasNoDetailAndNoRemedy(t *testing.T) {
	result := BeadsInstalled.Pass()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want PASS", result.Status)
	}

	if result.Detail.IsPresent() {
		t.Errorf("Detail = %v, want None", result.Detail)
	}

	if result.Remedy.IsPresent() {
		t.Errorf("Remedy = %v, want None", result.Remedy)
	}

	if result.Check != BeadsInstalled {
		t.Errorf("Check = %v, want %v", result.Check, BeadsInstalled)
	}
}

func TestPassWithDetailCarriesTheDetailOnly(t *testing.T) {
	result := BeadsInstalled.PassWithDetail("bd version 1.2.2")

	if got, ok := result.Detail.Get(); !ok || got != "bd version 1.2.2" {
		t.Errorf("Detail = %v, want Some(%q)", result.Detail, "bd version 1.2.2")
	}

	if result.Remedy.IsPresent() {
		t.Errorf("Remedy = %v, want None", result.Remedy)
	}
}

func TestWarnAndFailCarryDetailAndRemedy(t *testing.T) {
	warn := BeadsInitialized.Warn("no database", mo.Some("bd init"))
	if warn.Status != StatusWarn {
		t.Errorf("Warn status = %v, want WARN", warn.Status)
	}

	if got, ok := warn.Remedy.Get(); !ok || got != "bd init" {
		t.Errorf("Warn remedy = %v, want Some(%q)", warn.Remedy, "bd init")
	}

	fail := GHScopes.Fail("missing repo", mo.None[string]())
	if fail.Status != StatusFail {
		t.Errorf("Fail status = %v, want FAIL", fail.Status)
	}

	if got, ok := fail.Detail.Get(); !ok || got != "missing repo" {
		t.Errorf("Fail detail = %v, want Some(%q)", fail.Detail, "missing repo")
	}

	if fail.Remedy.IsPresent() {
		t.Errorf("Fail remedy = %v, want None", fail.Remedy)
	}
}

func TestCheckIdentifiersAreDistinct(t *testing.T) {
	checks := []Check{
		CodefallDir, SettingsFile, SettingsJSON, SettingsComplete,
		BeadsInstalled, BeadsInitialized, GHInstalled, GHAuthenticated, GHScopes,
	}

	seen := map[string]bool{}

	for _, check := range checks {
		if check.ID == "" || check.Title == "" {
			t.Errorf("check %+v has an empty field", check)
		}

		if seen[check.ID] {
			t.Errorf("duplicate check ID %q", check.ID)
		}

		seen[check.ID] = true
	}
}
