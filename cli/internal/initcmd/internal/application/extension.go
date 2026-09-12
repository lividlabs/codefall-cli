package application

import (
	"context"
	"fmt"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// Where each harness's files go, relative to the directory init is run in. The display form is
// what a person reads in a report; the path form is what the file system is given.
const (
	claudeDir      = ".claude"
	agentsDir      = ".agents"
	claudeFile     = "settings.json"
	claudeFullName = claudeDir + "/" + claudeFile
)

// where each harness's skills live, the convention each of those directories carries being the
// same to the little letters of the job. Claude Code's own is .claude/; an agent that reads the
// .agents/skills convention takes .agents/, and adding one of those is one row.
var extensionDestDirs = map[string]string{
	domain.HarnessAntigravity: agentsDir,
	domain.HarnessClaudeCode:  claudeDir,
	domain.HarnessCodex:       agentsDir,
	domain.HarnessMuse:        agentsDir,
	domain.HarnessOpenCode:    agentsDir,
}

// extension is the second step of a run: it mirrors the embedded extension tree where the harness the
// project uses reads skills, so it never touches a network or a foreign CLI. It also hands back
// every path it wrote, which the run records in the manifest once the last step has succeeded.
func (i *Initialize) extension(ctx context.Context, request Request) (domain.StepResult, []string, error) {
	if _, known := extensionDestDirs[request.Harness]; !known {
		return domain.StepResult{}, nil, fmt.Errorf("harness %q has no extension mechanism", request.Harness)
	}

	return i.skillsDirExtension(ctx, request)
}
