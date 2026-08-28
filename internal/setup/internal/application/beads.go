package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// beadsCommand is how Beads is invoked, and the arguments init gives it.
//
// --non-interactive because init has already asked its questions. --skip-agents because bd would
// otherwise append its own section to AGENTS.md and CLAUDE.md and write .claude/settings.json,
// .codex/, and .agents/ — codefall owns those files, and the context bd's section points at reaches
// a session through the hook step instead.
const beadsCommand = "bd"

var beadsInitArgs = []string{"init", "--non-interactive", "--skip-agents"}

// What bd says when it committed what it wrote, and the message it committed under. bd commits
// unconditionally and has no flag to stop it, so the report says so rather than letting a commit
// nobody asked for turn up in the log unexplained. What went into it is left unnamed: bd stages
// .beads/ and .gitignore, and also every one of the files preflight guards, so the list depends on
// what the directory had.
const (
	beadsCommitted     = "Committed beads files to git"
	beadsCommitMessage = "bd init: initialize beads issue tracking"
)

// beads is the third step of a run: it creates the issue database that codefall's agents work
// against. Preflight has already established that bd is on PATH, that this is a git work tree, and
// that nothing is staged — the three things bd init would otherwise decide for itself.
func (i *Initialize) beads(ctx context.Context, request Request) (domain.StepResult, error) {
	// Preflight asked bd this same question, to decide whether what the directory had waiting was
	// any of its business. It is asked again rather than carried across, because a step takes no
	// state from preflight: what preflight leaves behind is a run that was allowed to start.
	initialized, err := i.beadsInitialized(ctx, request.Dir)
	if err != nil {
		return domain.StepResult{}, err
	}

	if initialized {
		return domain.BeadsStep.Skipped("Beads is already initialized here"), nil
	}

	result, err := i.probe(ctx, request.Dir, beadsCommand, beadsInitArgs...)
	if err != nil {
		return domain.StepResult{}, err
	}

	if result.ExitCode != 0 {
		detail := firstLine(result.Stderr)
		if detail == "" {
			detail = firstLine(result.Stdout)
		}

		return domain.StepResult{}, fmt.Errorf("%s exited %d: %s",
			commandLine(beadsCommand, beadsInitArgs), result.ExitCode, detail)
	}

	// bd warns on stderr about things a fresh database has not got yet — no Dolt remote, for one —
	// and it does that on a run that worked. A successful run has nothing to complain about, so what
	// it said there is dropped.
	if strings.Contains(result.Stdout, beadsCommitted) {
		return domain.BeadsStep.Done(fmt.Sprintf(
			"initialized Beads (bd init committed what it wrote as %q)", beadsCommitMessage)), nil
	}

	return domain.BeadsStep.Done("initialized Beads"), nil
}

// beadsInitialized asks bd whether this directory already has a database. `bd info` is the question
// doctor asks for the same reason: BEADS_DIR relocates .beads/, so looking for the directory is not
// the same question, and re-running bd init where it has already run is an error rather than a
// no-op.
func (i *Initialize) beadsInitialized(ctx context.Context, dir string) (bool, error) {
	result, err := i.probe(ctx, dir, beadsCommand, "info")
	if err != nil {
		return false, err
	}

	return result.ExitCode == 0, nil
}
