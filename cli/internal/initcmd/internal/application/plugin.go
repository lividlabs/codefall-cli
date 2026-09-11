package application

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// Where Claude Code keeps a project's settings, relative to the directory init is run in. The
// display form is what a person reads in a report; the path form is what the file system is given.
const (
	claudeDir      = ".claude"
	agentsDir      = ".agents"
	claudeFile     = "settings.json"
	claudeFullName = claudeDir + "/" + claudeFile
)

// where each harness's skills live, the convention each of those directories carries being the
// same to the little letters of the job. Claude Code's own is .claude/; an agent that reads the
// .agents/skills convention takes .agents/, and adding one of those is one row.
var pluginDestDirs = map[string]string{
	domain.HarnessAntigravity: agentsDir,
	domain.HarnessClaudeCode:  claudeDir,
	domain.HarnessCodex:       agentsDir,
	domain.HarnessMuse:        agentsDir,
	domain.HarnessOpenCode:    agentsDir,
}

// plugin is the second step of a run: it mirrors the embedded plugin tree where the harness the
// project uses reads skills, so it never touches a network or a foreign CLI.
func (i *Initialize) plugin(ctx context.Context, request Request) (domain.StepResult, error) {
	if _, known := pluginDestDirs[request.Harness]; !known {
		return domain.StepResult{}, fmt.Errorf("harness %q has no plugin mechanism", request.Harness)
	}

	return i.skillsDirPlugin(ctx, request)
}

// claudePath is the one read of the harness's settings file, shared by the steps that merge into
// it — the hook step encodes the plain object with this at both ends.
//
// Three files say the same nothing: one that is not there, one that is empty or only whitespace,
// and one holding the JSON literal null. They are one answer here rather than three behaviours
// further down, because json.Unmarshal accepts null into anything and leaves it as it was while
// rejecting the other two (ADR-GO-03). Anything else is handed back for the caller to make sense
// of.
func (i *Initialize) readClaudeFile(dir string) (mo.Option[[]byte], error) {
	data, err := i.files.ReadFile(claudePath(dir))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return mo.None[[]byte](), nil
	case err != nil:
		return mo.None[[]byte](), fmt.Errorf("read %s: %w", claudeFullName, err)
	}

	if trimmed := strings.TrimSpace(string(data)); trimmed == "" || trimmed == "null" {
		return mo.None[[]byte](), nil
	}

	return mo.Some(data), nil
}

func claudePath(dir string) string {
	return filepath.Join(dir, claudeDir, claudeFile)
}
