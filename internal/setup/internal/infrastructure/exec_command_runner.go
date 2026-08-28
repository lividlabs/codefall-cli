package infrastructure

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/application"
)

// ExecCommandRunner runs external tools as child processes. This is the only file in the component
// that imports os/exec.
type ExecCommandRunner struct{}

// NewExecCommandRunner builds the real command gateway.
func NewExecCommandRunner() *ExecCommandRunner {
	return &ExecCommandRunner{}
}

// LookPath resolves a tool on PATH. Any failure means the same thing to the caller — the tool is not
// available — so the reason is dropped.
func (*ExecCommandRunner) LookPath(name string) mo.Option[string] {
	path, err := exec.LookPath(name)
	if err != nil {
		return mo.None[string]()
	}

	return mo.Some(path)
}

// Run executes a tool in dir and captures both streams. cmd.Env is left nil so the child inherits
// the environment: PATH, BEADS_DIR, and GH_TOKEN all matter to the tools doctor asks about.
func (*ExecCommandRunner) Run(
	ctx context.Context, dir, name string, args ...string,
) (application.CommandResult, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := application.CommandResult{Stdout: stdout.String(), Stderr: stderr.String()}

	var exitErr *exec.ExitError

	switch {
	case err == nil:
		return result, nil
	case errors.As(err, &exitErr):
		// A tool that ran and disagreed is a normal result, not a failure to run it.
		result.ExitCode = exitErr.ExitCode()

		return result, nil
	default:
		return application.CommandResult{}, fmt.Errorf("run %s: %w", name, err)
	}
}
