// Package application holds initcmd's one use case. It owns the gateway interfaces the steps need
// and depends on nothing but the standard library, mo, and initcmd's own domain.
package application

import (
	"context"
	"fmt"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// FileSystem is initcmd's view of the working directory (one gateway role). It reads what is already
// there and writes what init creates; the permissions those writes use are infrastructure's
// decision, not the use case's.
type FileSystem interface {
	// ReadFile returns a file's bytes. A missing file satisfies errors.Is(err, fs.ErrNotExist).
	ReadFile(path string) ([]byte, error)
	// MkdirAll creates path and every parent it needs, and does nothing when it already exists.
	MkdirAll(path string) error
	// WriteFile writes data to path, replacing whatever was there.
	WriteFile(path string, data []byte) error
}

// CommandRunner locates and runs the external tools codefall depends on (one gateway role). It is
// the same role doctor declares, because it is the same job: initcmd asks gh what repository this
// directory belongs to, asks git about the directory, and runs the harness and Beads.
type CommandRunner interface {
	// LookPath returns where a tool lives, or None when it is not on PATH. There is no useful "why"
	// behind a missing binary, so this is an Option and not an error (ADR-GO-03).
	LookPath(name string) mo.Option[string]
	// Run executes a tool and waits for it. A non-zero exit is a normal result; the error is
	// reserved for a command that could not be started at all.
	Run(ctx context.Context, dir, name string, args ...string) (CommandResult, error)
}

// CommandResult is what one external command produced.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// PluginFetcher copies the plugin's files and records what it copied (one gateway role): the
// embedded tree in the binary, so Fetch is a local copy, and the relative paths it wrote are
// what init uses for the install manifest.
type PluginFetcher interface {
	// Fetch mirrors the plugin's tree onto destDir and returns every path it wrote, relative to
	// destDir, so the caller can put it on record.
	Fetch(ctx context.Context, destDir string) ([]string, error)
}

// Request is what presentation hands the use case: every answer a survey or a set of flags could
// collect, already parsed. Absence is an Option, so "no project number" and "project number zero"
// cannot be confused (ADR-GO-03). Tracker may be empty when settings already exist and Force is
// false, because then nothing is built from it — so a step that comes to depend on which tracker the
// project uses must read it out of the settings file that is already there, not out of this field.
type Request struct {
	Dir           string
	Tracker       string
	GitHubRepo    mo.Option[string]
	GitHubProject mo.Option[int]
	Harness       string
	Force         bool
}

// Observer watches a run step by step, so a terminal can show what is happening while it happens.
// It is declared here because the use case is what has something to say; what a listener does with
// it — a spinner, a log line, nothing at all — belongs to the caller.
type Observer interface {
	// StepStarted is called before a step does its work.
	StepStarted(step domain.Step)
	// StepFinished is called after a step has done its work, and not at all for a step that failed.
	StepFinished(result domain.StepResult)
}

// Initialize sets a project up for codefall: it writes .codefall/settings.json, installs the harness
// plugin, initialises Beads, gives the harness the session hook that primes a session with what
// Beads knows, and writes the section of AGENTS.md that says how the project uses it.
type Initialize struct {
	files   FileSystem
	runner  CommandRunner
	fetcher PluginFetcher
}

// NewInitialize builds the use case over its gateways.
func NewInitialize(files FileSystem, runner CommandRunner, fetcher PluginFetcher) *Initialize {
	return &Initialize{files: files, runner: runner, fetcher: fetcher}
}

// step is one unit of work in a run. It returns what it did, or an error that stops the run: a step
// that could not finish leaves the project half set up, and the steps after it would be building on
// something that is not there.
type step struct {
	domain.Step
	run func(ctx context.Context, request Request) (domain.StepResult, error)
}

// Run performs every step in order, telling the observer as each one starts and finishes. The first
// step that fails ends the run, and its error names the step so the reader knows how far init got.
//
// The tools the run needs are checked before the first step starts, so a missing one leaves the
// project untouched rather than half set up. That error already says what is wrong and what to do
// about it, so it is returned as it is, with no step to name in front of it.
func (i *Initialize) Run(ctx context.Context, request Request, observer Observer) (domain.Report, error) {
	if observer == nil {
		observer = silentObserver{}
	}

	if err := i.preflight(ctx, request); err != nil {
		return domain.Report{}, err
	}

	steps := []step{
		{Step: domain.SettingsStep, run: i.settings},
		{Step: domain.PluginStep, run: i.plugin},
		{Step: domain.BeadsStep, run: i.beads},
		{Step: domain.HookStep, run: i.hook},
		{Step: domain.AgentsStep, run: i.agents},
	}

	results := make([]domain.StepResult, 0, len(steps))

	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			return domain.Report{}, fmt.Errorf("initialize: %w", err)
		}

		observer.StepStarted(step.Step)

		result, err := step.run(ctx, request)
		if err != nil {
			return domain.Report{}, fmt.Errorf("%s: %w", step.ID, err)
		}

		observer.StepFinished(result)

		results = append(results, result)
	}

	return domain.NewReport(results...), nil
}

// silentObserver stands in for a caller that has nothing to show, so Run has no nil check in its
// loop.
type silentObserver struct{}

func (silentObserver) StepStarted(domain.Step)        {}
func (silentObserver) StepFinished(domain.StepResult) {}
