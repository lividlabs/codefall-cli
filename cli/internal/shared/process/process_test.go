package process

import (
	"os"
	"path/filepath"
	"testing"
)

// These shell out to sh, which every platform this tool targets has. Windows is out of scope.

func TestLookPath(t *testing.T) {
	if path, ok := LookPath("sh").Get(); !ok || path == "" {
		t.Errorf(`LookPath("sh") = %v, want a path`, LookPath("sh"))
	}

	if got := LookPath("no-such-binary-codefall"); got.IsPresent() {
		t.Errorf("LookPath of a missing binary = %v, want None", got)
	}
}

func TestRunHonoursDir(t *testing.T) {
	withMarker := t.TempDir()
	if err := os.WriteFile(filepath.Join(withMarker, "marker"), nil, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	found, err := Run(t.Context(), withMarker, "sh", "-c", "test -f marker")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if found.ExitCode != 0 {
		t.Errorf("exit code in the directory with the marker = %d, want 0", found.ExitCode)
	}

	absent, err := Run(t.Context(), t.TempDir(), "sh", "-c", "test -f marker")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if absent.ExitCode != 1 {
		t.Errorf("exit code in the directory without the marker = %d, want 1", absent.ExitCode)
	}
}

func TestRunCapturesBothStreamsAndTheExitCode(t *testing.T) {
	result, err := Run(t.Context(), t.TempDir(), "sh", "-c", "echo out; echo err 1>&2; exit 3")
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

func TestRunReportsACommandItCannotStart(t *testing.T) {
	if _, err := Run(t.Context(), t.TempDir(), "no-such-binary-codefall"); err == nil {
		t.Error("Run of a missing binary = nil error, want an error")
	}
}

// The environment is inherited rather than replaced, which is what makes BEADS_DIR and GH_TOKEN
// reach the tools codefall drives.
func TestRunInheritsTheEnvironment(t *testing.T) {
	t.Setenv("CODEFALL_PROCESS_TEST", "inherited")

	result, err := Run(t.Context(), t.TempDir(), "sh", "-c", "printf %s \"$CODEFALL_PROCESS_TEST\"")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Stdout != "inherited" {
		t.Errorf("Stdout = %q, want the inherited value %q", result.Stdout, "inherited")
	}
}
