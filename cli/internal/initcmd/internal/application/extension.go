package application

import (
	"context"
	"fmt"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/harness"
)

// extension is the second step of a run: it mirrors the embedded extension tree where the harness the
// project uses reads skills, so it never touches a network or a foreign CLI. It also hands back
// every path it wrote, which the run records in the manifest once the last step has succeeded.
//
// Which directory that is belongs to the shared harness module rather than to this step: doctor
// reports on the same directories, so neither component can be the one that decides where they are
// (ADR-003).
func (i *Initialize) extension(ctx context.Context, request Request) (domain.StepResult, []string, error) {
	dest, known := harness.SkillsDir(request.Harness).Get()
	if !known {
		return domain.StepResult{}, nil, fmt.Errorf("harness %q has no extension mechanism", request.Harness)
	}

	return i.skillsDirExtension(ctx, request, dest)
}
