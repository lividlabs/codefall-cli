// Package process is the shared module every component's infrastructure reaches the operating system
// through: the external tools codefall drives and the files it reads and writes. A component wraps
// these in a gateway of its own, because the interface and the result type are the consumer's to
// declare (ADR-GO-02 rule 4: this module never imports a component).
package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/samber/mo"
)

// Result is what one external command produced.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// LookPath resolves a tool on PATH. Any failure means the same thing to the caller — the tool is not
// available — so the reason is dropped (ADR-GO-03).
func LookPath(name string) mo.Option[string] {
	path, err := exec.LookPath(name)
	if err != nil {
		return mo.None[string]()
	}

	return mo.Some(path)
}

// Run executes a tool in dir and captures both streams. A non-zero exit is a normal result; the
// error is reserved for a command that could not be started at all.
//
// cmd.Env is left nil so the child inherits the environment: PATH, BEADS_DIR, and GH_TOKEN all
// matter to the tools codefall asks about.
func Run(ctx context.Context, dir, name string, args ...string) (Result, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{Stdout: stdout.String(), Stderr: stderr.String()}

	var exitErr *exec.ExitError

	switch {
	case err == nil:
		return result, nil
	case errors.As(err, &exitErr):
		// A tool that ran and disagreed is a normal result, not a failure to run it.
		result.ExitCode = exitErr.ExitCode()

		return result, nil
	default:
		return Result{}, fmt.Errorf("run %s: %w", name, err)
	}
}
