package application

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// agentsDir is where a harness that reads the .agents/skills convention keeps its skills, relative
// to the directory init runs in. The display form is what a person reads in a report; the path form
// is what the file system is given.
const (
	agentsDir      = ".agents"
	agentsFullName = agentsDir + "/"
)

// skillsDirPlugin is the plugin step for every harness that reads the .agents/skills convention:
// it copies the embedded plugin tree under the project's .agents/. There is nothing to compare — a
// copy of fifty files is cheap, an install-what-you-have is not a state to negotiate with.
func (i *Initialize) skillsDirPlugin(ctx context.Context, request Request) (domain.StepResult, error) {
	if err := i.fetcher.Fetch(ctx, filepath.Join(request.Dir, agentsDir)); err != nil {
		return domain.StepResult{}, fmt.Errorf("install the embedded plugin: %w", err)
	}

	return domain.PluginStep.Done("installed codefall's skills into " + agentsFullName), nil
}
