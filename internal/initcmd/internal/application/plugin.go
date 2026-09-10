package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/internal/shared/text"
)

// claudeCommand is how Claude Code is invoked. The CLI's name and the shape of its arguments are
// delivery detail, not facts about codefall, which is why they live here and the plugin's own
// identifiers live in the domain.
const claudeCommand = "claude"

// Where Claude Code keeps a project's settings, relative to the directory init is run in. The
// display form is what a person reads in a report; the path form is what the file system is given.
const (
	claudeDir      = ".claude"
	claudeFile     = "settings.json"
	claudeFullName = claudeDir + "/" + claudeFile
)

// how a harness gets codefall's plugin. Claude Code takes a marketplace install through its own
// CLI; every other harness takes the plugin's files, mirrored under the project's .agents/
// directory. The mapping from harness to mechanism is the table below: adding a harness that
// uses the skills directory is one row, and adding a mechanism is a new case in plugin().
type mechanism int

const (
	claudeMarketplace mechanism = iota
	skillsIntoAgents
)

var harnessMechanisms = map[string]mechanism{
	domain.HarnessAntigravity: skillsIntoAgents,
	domain.HarnessClaudeCode:  claudeMarketplace,
	domain.HarnessCodex:       skillsIntoAgents,
	domain.HarnessMuse:        skillsIntoAgents,
	domain.HarnessOpenCode:    skillsIntoAgents,
}

// plugin is the second step of a run: it makes codefall's skills available to the harness the
// project uses.
func (i *Initialize) plugin(ctx context.Context, request Request) (domain.StepResult, error) {
	mechanism, known := harnessMechanisms[request.Harness]
	if !known {
		return domain.StepResult{}, fmt.Errorf("harness %q has no plugin mechanism", request.Harness)
	}

	switch mechanism {
	case claudeMarketplace:
		return i.claudeCodePlugin(ctx, request.Dir)
	case skillsIntoAgents:
		return i.skillsDirPlugin(ctx, request)
	default:
		return domain.StepResult{}, fmt.Errorf("mechanism %d has no installer", mechanism)
	}
}

// claudeCodePlugin declares the codefall marketplace and installs the plugin, both at project scope
// so that .claude/settings.json carries them and everyone who clones the repository gets them.
//
// What is already done is read out of that file rather than asked of `claude plugin list`, whose
// enabled flag is computed from the current working directory and stamped onto every row. The file
// is the project-local truth; the CLI merges into it, so init never writes it itself.
//
// The two halves are decided separately, because a project can have either without the other. A
// plugin enabled with no marketplace declared at project scope is what a user-scope marketplace
// leaves behind: it works for whoever ran the install and for nobody who clones the repository. The
// step is skipped only when the file already carries both.
func (i *Initialize) claudeCodePlugin(ctx context.Context, dir string) (domain.StepResult, error) {
	settings, err := i.readClaudeSettings(dir)
	if err != nil {
		return domain.StepResult{}, err
	}

	_, declared := settings.ExtraKnownMarketplaces[domain.MarketplaceName]
	enabled := settings.EnabledPlugins[domain.PluginID]

	if declared && enabled {
		return domain.PluginStep.Skipped(fmt.Sprintf(
			"the %s marketplace is declared and %s enabled in %s",
			domain.MarketplaceName, domain.PluginID, claudeFullName)), nil
	}

	if !declared {
		if err := i.claude(ctx, dir,
			"plugin", "marketplace", "add", domain.MarketplaceSource, "--scope", "project"); err != nil {
			return domain.StepResult{}, err
		}
	}

	if !enabled {
		if err := i.claude(ctx, dir,
			"plugin", "install", domain.PluginID, "--scope", "project", "-y"); err != nil {
			return domain.StepResult{}, err
		}
	}

	return domain.PluginStep.Done(pluginDetail(declared, enabled)), nil
}

// pluginDetail is what the step reports it did, which is whichever of the two things it found
// missing. Both already there is not one of these sentences: that is the skip.
func pluginDetail(declared, enabled bool) string {
	var did []string

	if !declared {
		did = append(did, "declared the "+domain.MarketplaceName+" marketplace")
	}

	if !enabled {
		did = append(did, "installed "+domain.PluginID)
	}

	return fmt.Sprintf("%s for this project (%s)", strings.Join(did, " and "), claudeFullName)
}

// claude runs the harness CLI in the project's directory and turns a refusal into the error that
// stops the run. A tool's complaint is summarised to its first line, the way doctor summarises one.
func (i *Initialize) claude(ctx context.Context, dir string, args ...string) error {
	command := strings.Join(append([]string{claudeCommand}, args...), " ")

	result, err := i.runner.Run(ctx, dir, claudeCommand, args...)
	if err != nil {
		return fmt.Errorf("run %s: %w", command, err)
	}

	if result.ExitCode != 0 {
		detail := text.FirstLine(result.Stderr)
		if detail == "" {
			detail = text.FirstLine(result.Stdout)
		}

		return fmt.Errorf("%s exited %d: %s", command, result.ExitCode, detail)
	}

	return nil
}

// claudeSettings is the part of .claude/settings.json init reads: which plugins the project enables
// and which marketplaces it declares. Decoding lives here rather than in the domain, which must not
// name an encoding, or in infrastructure, which takes bytes — the same reason settings.go decodes
// where it does. Every other key in the file is none of init's business.
type claudeSettings struct {
	EnabledPlugins         map[string]bool            `json:"enabledPlugins"`
	ExtraKnownMarketplaces map[string]json.RawMessage `json:"extraKnownMarketplaces"`
}

// readClaudeSettings reads what the project has already told Claude Code. A file with nothing to say
// is not a problem: the CLI creates it, and declares nothing until it does. A file that cannot be
// read or is not settings at all is a problem, because the CLI is about to merge into it.
func (i *Initialize) readClaudeSettings(dir string) (claudeSettings, error) {
	body, err := i.readClaudeFile(dir)
	if err != nil {
		return claudeSettings{}, err
	}

	data, said := body.Get()
	if !said {
		return claudeSettings{}, nil
	}

	var settings claudeSettings

	if err := json.Unmarshal(data, &settings); err != nil {
		return claudeSettings{}, fmt.Errorf("decode %s: %w", claudeFullName, err)
	}

	return settings, nil
}

// readClaudeFile is the one read of the harness's settings file, shared by the two steps that write
// it — the plugin step decodes it into a struct, the hook step into a plain object.
//
// Three files say the same nothing: one that is not there, one that is empty or only whitespace, and
// one holding the JSON literal null. They are one answer here rather than three behaviours further
// down, because json.Unmarshal accepts null into anything and leaves it as it was while rejecting
// the other two (ADR-GO-03). Anything else is handed back for the caller to make sense of.
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
