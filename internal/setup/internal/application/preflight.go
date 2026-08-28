package application

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// requiredTool is an external tool a run cannot do without, and what a reader should do about it
// being missing.
type requiredTool struct {
	name   string
	remedy string
}

// preflight checks that every tool this run will reach for is on PATH, before any step has done
// anything. A tool that turns up missing halfway through leaves the project half set up, which is
// the one state init exists to avoid — so the whole run refuses to start instead.
//
// Every missing tool is reported, not just the first: someone installing prerequisites wants the
// whole list in one go.
func (i *Initialize) preflight(request Request) error {
	var missing []string

	for _, tool := range requiredTools(request) {
		if i.runner.LookPath(tool.name).IsAbsent() {
			missing = append(missing, fmt.Sprintf("%s is not on PATH (%s)", tool.name, tool.remedy))
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return errors.New(strings.Join(missing, "; "))
}

// requiredTools is what this particular run needs, which depends on the answers it was given: the
// harness decides whose CLI installs the plugin, and the tracker will decide whose CLI initialises
// the issue database.
func requiredTools(request Request) []requiredTool {
	var tools []requiredTool

	if request.Harness == domain.HarnessClaudeCode {
		tools = append(tools, requiredTool{
			name:   claudeCommand,
			remedy: "install Claude Code: https://docs.anthropic.com/en/docs/claude-code/setup",
		})
	}

	return tools
}
