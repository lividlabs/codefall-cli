package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/application"
)

var _ application.CommandRunner = (*ExecCommandRunner)(nil)

// These shell out to sh, which every platform this tool targets has. Windows is out of scope.

func TestExecCommandRunnerLookPath(t *testing.T) {
	runner := NewExecCommandRunner()

	if path, ok := runner.LookPath("sh").Get(); !ok || path == "" {
		t.Errorf(`LookPath("sh") = %v, want a path`, runner.LookPath("sh"))
	}

	if got := runner.LookPath("no-such-binary-codefall"); got.IsPresent() {
		t.Errorf("LookPath of a missing binary = %v, want None", got)
	}
}

func TestExecCommandRunnerHonoursDir(t *testing.T) {
	withMarker := t.TempDir()
	if err := os.WriteFile(filepath.Join(withMarker, "marker"), nil, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	runner := NewExecCommandRunner()

	found, err := runner.Run(t.Context(), withMarker, "sh", "-c", "test -f marker")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if found.ExitCode != 0 {
		t.Errorf("exit code in the directory with the marker = %d, want 0", found.ExitCode)
	}

	absent, err := runner.Run(t.Context(), t.TempDir(), "sh", "-c", "test -f marker")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if absent.ExitCode != 1 {
		t.Errorf("exit code in the directory without the marker = %d, want 1", absent.ExitCode)
	}
}

func TestExecCommandRunnerCapturesBothStreamsAndTheExitCode(t *testing.T) {
	runner := NewExecCommandRunner()

	result, err := runner.Run(t.Context(), t.TempDir(), "sh", "-c", "echo out; echo err 1>&2; exit 3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Stdout != "out\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "out\n")
	}

	if result.Stderr != "err\n" {
		t.Errorf("Stderr = %q, want %q", result.Stderr, "err\n")
	}

	if result.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", result.ExitCode)
	}
}

func TestExecCommandRunnerReportsACommandItCannotStart(t *testing.T) {
	runner := NewExecCommandRunner()

	if _, err := runner.Run(t.Context(), t.TempDir(), "no-such-binary-codefall"); err == nil {
		t.Error("Run of a missing binary = nil error, want an error")
	}
}
