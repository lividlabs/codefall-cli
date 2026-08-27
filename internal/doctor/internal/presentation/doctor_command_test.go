package presentation

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/domain"
)

type fakeDiagnose struct {
	report domain.Report
	err    error
	gotDir string
}

func (f *fakeDiagnose) Run(_ context.Context, dir string) (domain.Report, error) {
	f.gotDir = dir

	return f.report, f.err
}

// run executes the command against a fake use case and returns everything it wrote.
//
// CLICOLOR_FORCE and TTY_FORCE are cleared so a developer's forced-colour environment cannot leak
// ANSI into the buffer: colorprofile honours both regardless of whether the writer is a terminal.
func run(t *testing.T, diagnose DiagnoseUseCase, args ...string) (string, error) {
	t.Helper()
	t.Setenv("CLICOLOR_FORCE", "0")
	t.Setenv("TTY_FORCE", "0")

	var out bytes.Buffer

	cmd := NewDoctorCommand(diagnose)
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	// Fang sets both on the root in production, because it renders errors and usage itself.
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()

	return out.String(), err
}

func TestDoctorCommandPrintsAPassingReport(t *testing.T) {
	diagnose := &fakeDiagnose{report: domain.NewReport(
		domain.CodefallDir.Pass(),
		domain.BeadsInstalled.PassWithDetail("bd version 1.2.2 (Homebrew)"),
	)}

	out, err := run(t, diagnose)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "PASS  .codefall/ exists\n" +
		"PASS  bd is on PATH — bd version 1.2.2 (Homebrew)\n"
	if out != want {
		t.Errorf("output =\n%q\nwant\n%q", out, want)
	}
}

func TestDoctorCommandPrintsWarningsWithTheirRemedy(t *testing.T) {
	diagnose := &fakeDiagnose{report: domain.NewReport(
		domain.BeadsInitialized.Warn("bd info exited 1: Error: no beads database found", mo.Some("bd init")),
	)}

	out, err := run(t, diagnose)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "WARN  Beads is initialized here — bd info exited 1: Error: no beads database found (fix: bd init)\n"
	if out != want {
		t.Errorf("output =\n%q\nwant\n%q", out, want)
	}
}

func TestDoctorCommandFailsWhenAnyCheckFails(t *testing.T) {
	for _, tc := range []struct {
		name    string
		report  domain.Report
		wantErr string
	}{
		{
			name:    "one failure",
			report:  domain.NewReport(domain.SettingsFile.Fail("not found", mo.None[string]())),
			wantErr: "doctor: 1 check failed",
		},
		{
			name: "two failures",
			report: domain.NewReport(
				domain.SettingsFile.Fail("not found", mo.None[string]()),
				domain.BeadsInitialized.Warn("no database", mo.Some("bd init")),
				domain.GHScopes.Fail("missing repo", mo.Some("gh auth refresh -s repo")),
			),
			wantErr: "doctor: 2 checks failed",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := run(t, &fakeDiagnose{report: tc.report})
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("Execute error = %v, want %q", err, tc.wantErr)
			}

			// Every line is printed before the error is returned.
			for _, result := range tc.report.Results() {
				if !strings.Contains(out, result.Check.Title) {
					t.Errorf("output does not mention %q:\n%s", result.Check.Title, out)
				}
			}
		})
	}
}

func TestDoctorCommandWrapsAUseCaseError(t *testing.T) {
	failure := errors.New("diagnose: context canceled")

	out, err := run(t, &fakeDiagnose{err: failure})
	if !errors.Is(err, failure) {
		t.Fatalf("Execute error = %v, want it to wrap %v", err, failure)
	}

	if out != "" {
		t.Errorf("output = %q, want nothing printed", out)
	}
}

func TestDoctorCommandTakesNoArguments(t *testing.T) {
	if _, err := run(t, &fakeDiagnose{}, "extra"); err == nil {
		t.Error("Execute with an argument = nil error, want an error")
	}
}

func TestDoctorCommandDiagnosesTheWorkingDirectory(t *testing.T) {
	want, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	diagnose := &fakeDiagnose{}

	if _, err := run(t, diagnose); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if diagnose.gotDir != want {
		t.Errorf("diagnosed %q, want %q", diagnose.gotDir, want)
	}
}

func TestFormatLine(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result domain.Result
		want   string
	}{
		{
			name:   "a bare pass",
			result: domain.CodefallDir.Pass(),
			want:   "PASS  .codefall/ exists",
		},
		{
			name:   "a pass with detail",
			result: domain.GHAuthenticated.PassWithDetail("as djensen47"),
			want:   "PASS  gh is logged in to github.com — as djensen47",
		},
		{
			name:   "a warning with a remedy",
			result: domain.BeadsInitialized.Warn("no database", mo.Some("bd init")),
			want:   "WARN  Beads is initialized here — no database (fix: bd init)",
		},
		{
			name:   "a failure without a remedy",
			result: domain.SettingsJSON.Fail("invalid JSON at byte 4: unexpected end", mo.None[string]()),
			want:   "FAIL  settings.json is valid JSON — invalid JSON at byte 4: unexpected end",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// formatLine styles the status word; strip the styling to compare the text.
			if got := stripANSI(formatLine(tc.result)); got != tc.want {
				t.Errorf("formatLine() = %q, want %q", got, tc.want)
			}
		})
	}
}

// stripANSI removes the escape sequences lipgloss renders, which the colorprofile writer would
// normally downsample away on its way to a non-terminal.
func stripANSI(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			for i < len(s) && s[i] != 'm' {
				i++
			}

			continue
		}

		b.WriteByte(s[i])
	}

	return b.String()
}
