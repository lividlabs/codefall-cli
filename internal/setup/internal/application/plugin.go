package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
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

// plugin is the second step of a run: it makes codefall's skills available to the harness the
// project uses.
//
// One case per harness. Claude Code is the only one today, and presentation has already refused
// every other value — so the default is unreachable, and it is where the next harness lands.
func (i *Initialize) plugin(ctx context.Context, request Request) (domain.StepResult, error) {
	switch request.Harness {
	case domain.HarnessClaudeCode:
		return i.claudeCodePlugin(ctx, request.Dir)
	default:
		return domain.StepResult{}, fmt.Errorf("harness %q has no plugin to install", request.Harness)
	}
}

// claudeCodePlugin declares the codefall marketplace and installs the plugin, both at project scope
// so that .claude/settings.json carries them and everyone who clones the repository gets them.
//
// What is already done is read out of that file rather than asked of `claude plugin list`, whose
// enabled flag is computed from the current working directory and stamped onto every row. The file
// is the project-local truth; the CLI merges into it, so init never writes it itself.
func (i *Initialize) claudeCodePlugin(ctx context.Context, dir string) (domain.StepResult, error) {
	settings, err := i.readClaudeSettings(dir)
	if err != nil {
		return domain.StepResult{}, err
	}

	if settings.EnabledPlugins[domain.PluginID] {
		return domain.PluginStep.Skipped(
			fmt.Sprintf("%s is already enabled in %s", domain.PluginID, claudeFullName)), nil
	}

	// Installing works without the declaration, because the marketplace may be registered for this
	// user — but then a teammate who clones the repository does not have it. Declaring it at project
	// scope is what makes the install repeatable for everyone else.
	if _, declared := settings.ExtraKnownMarketplaces[domain.MarketplaceName]; !declared {
		if err := i.claude(ctx, dir,
			"plugin", "marketplace", "add", domain.MarketplaceSource, "--scope", "project"); err != nil {
			return domain.StepResult{}, err
		}
	}

	if err := i.claude(ctx, dir,
		"plugin", "install", domain.PluginID, "--scope", "project", "-y"); err != nil {
		return domain.StepResult{}, err
	}

	return domain.PluginStep.Done(
		fmt.Sprintf("installed %s for this project (%s)", domain.PluginID, claudeFullName)), nil
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
		detail := firstLine(result.Stderr)
		if detail == "" {
			detail = firstLine(result.Stdout)
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

// readClaudeSettings reads what the project has already told Claude Code. A file that is not there
// yet says nothing, which is not a problem: the CLI creates it. A file that cannot be read or is not
// settings at all is a problem, because the CLI is about to merge into it.
func (i *Initialize) readClaudeSettings(dir string) (claudeSettings, error) {
	data, err := i.files.ReadFile(claudePath(dir))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return claudeSettings{}, nil
	case err != nil:
		return claudeSettings{}, fmt.Errorf("read %s: %w", claudeFullName, err)
	}

	var settings claudeSettings

	if err := json.Unmarshal(data, &settings); err != nil {
		return claudeSettings{}, fmt.Errorf("decode %s: %w", claudeFullName, err)
	}

	return settings, nil
}

func claudePath(dir string) string {
	return filepath.Join(dir, claudeDir, claudeFile)
}

// firstLine is how a tool's chatty output becomes one line of an error, as it is in doctor.
func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")

	return strings.TrimSpace(line)
}
