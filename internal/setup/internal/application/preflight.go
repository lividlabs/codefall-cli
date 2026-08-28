package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// gitCommand is how git is invoked. Init never asks git to change anything: it asks it two questions
// about the directory, because bd init answers them for itself in ways nobody asked for.
const gitCommand = "git"

// requiredTool is an external tool a run cannot do without, and what a reader should do about it
// being missing. A tool with nothing useful to suggest has no remedy, and the complaint is then just
// the missing tool.
type requiredTool struct {
	name   string
	remedy string
}

// preflight checks that every tool this run will reach for is on PATH and that the directory is one
// the run can work in, before any step has done anything. A tool that turns up missing halfway
// through leaves the project half set up, which is the one state init exists to avoid — so the whole
// run refuses to start instead.
//
// Every problem is reported, not just the first: someone clearing prerequisites wants the whole list
// in one go. A missing tool is the exception to that only in one direction — the questions that tool
// would have answered are not asked, so it is one complaint rather than a cascade of them.
func (i *Initialize) preflight(ctx context.Context, request Request) error {
	var problems []string

	add := func(problem string) {
		if problem != "" {
			problems = append(problems, problem)
		}
	}

	for _, tool := range requiredTools(request) {
		if i.runner.LookPath(tool.name).IsAbsent() {
			add(missingTool(tool))
		}
	}

	// git answers both questions below, so they are asked only when git is there.
	if i.runner.LookPath(gitCommand).IsPresent() {
		worktree, problem := i.worktreeProblem(ctx, request.Dir)
		add(problem)

		// bd init commits, so the index only matters when bd init is going to run — which needs bd
		// itself to say whether Beads is already initialised, and needs a work tree to commit in.
		if worktree && i.runner.LookPath(beadsCommand).IsPresent() {
			add(i.stagedChangesProblem(ctx, request.Dir))
		}
	}

	if len(problems) == 0 {
		return nil
	}

	return errors.New(strings.Join(problems, "; "))
}

// missingTool is the complaint about a tool that is not installed.
func missingTool(tool requiredTool) string {
	if tool.remedy == "" {
		return fmt.Sprintf("%s is not on PATH", tool.name)
	}

	return fmt.Sprintf("%s is not on PATH (%s)", tool.name, tool.remedy)
}

// worktreeProblem reports whether dir is inside a git work tree, and what to say when it is not.
// It matters because bd init run outside a repository silently runs git init first, which is a
// decision about the project that init has no business making on someone's behalf.
func (i *Initialize) worktreeProblem(ctx context.Context, dir string) (bool, string) {
	result, err := i.probe(ctx, dir, gitCommand, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, err.Error()
	}

	if result.ExitCode != 0 {
		return false, "not a git repository (bd init would create one; run git init first)"
	}

	return true, ""
}

// stagedChangesProblem reports what to say when the index has changes waiting and bd init is about
// to run. bd init makes its own commit with no pathspec, so anything already staged is swept into a
// commit that says it initialised Beads — which is why a run with a dirty index refuses to start.
//
// Beads that is already initialised means bd init will not run, and then the index is bd's business
// no more than it is init's.
func (i *Initialize) stagedChangesProblem(ctx context.Context, dir string) string {
	initialized, err := i.beadsInitialized(ctx, dir)
	if err != nil {
		return err.Error()
	}

	if initialized {
		return ""
	}

	result, err := i.probe(ctx, dir, gitCommand, "diff", "--cached", "--quiet")
	if err != nil {
		return err.Error()
	}

	switch result.ExitCode {
	case 0:
		return ""
	case 1:
		return "the git index has staged changes; bd init commits the whole index, " +
			"so commit or unstage them first"
	default:
		return fmt.Sprintf("git diff --cached --quiet exited %d: %s",
			result.ExitCode, firstLine(result.Stderr))
	}
}

// probe asks a tool a question in the project's directory. A non-zero exit is the answer, not a
// failure; a tool that could not be started at all is the error, phrased the way plugin.go phrases
// the same thing.
func (i *Initialize) probe(ctx context.Context, dir, name string, args ...string) (CommandResult, error) {
	result, err := i.runner.Run(ctx, dir, name, args...)
	if err != nil {
		return CommandResult{}, fmt.Errorf("run %s: %w", commandLine(name, args), err)
	}

	return result, nil
}

// commandLine is how a command reads back to the person who has to run it themselves.
func commandLine(name string, args []string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

// requiredTools is what this particular run needs, which depends on the answers it was given: the
// harness decides whose CLI installs the plugin. bd and git are needed by every run, because every
// run initialises Beads and Beads is initialised inside a git repository.
func requiredTools(request Request) []requiredTool {
	var tools []requiredTool

	if request.Harness == domain.HarnessClaudeCode {
		tools = append(tools, requiredTool{
			name:   claudeCommand,
			remedy: "install Claude Code: https://docs.anthropic.com/en/docs/claude-code/setup",
		})
	}

	return append(tools,
		requiredTool{name: beadsCommand, remedy: "brew install beads"},
		requiredTool{name: gitCommand},
	)
}
