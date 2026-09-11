package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// agentsDir is where a harness that reads the .agents/skills convention keeps its skills, relative
// to the directory init runs in. The display form is what a person reads in a report; the path form
// is what the file system is given. The plugin manifest made alongside them is what a rerun checks
// its version against.
const (
	agentsDir      = ".agents"
	agentsFullName = agentsDir + "/"
	pluginManifest = ".claude-plugin/plugin.json"
)

// skillsDirPlugin is the plugin step for every harness that reads the .agents/skills convention: it
// fetches the plugin's release from the plugin repository and mirrors its tree under the project's
// .agents/. A run that already has the release it is pointed at is skipped: the tree it would
// write is the tree that is already there.
func (i *Initialize) skillsDirPlugin(ctx context.Context, request Request) (domain.StepResult, error) {
	version := request.PluginVersion.OrElse(domain.PluginVersion)

	installed, err := i.installedPluginVersion(request.Dir)
	if err != nil {
		return domain.StepResult{}, err
	}

	if previous, ok := installed.Get(); ok && previous == version {
		return domain.PluginStep.Skipped(fmt.Sprintf(
			"codefall's skills are already at %s in %s", version, agentsFullName)), nil
	}

	if err := i.fetcher.Fetch(ctx, version, filepath.Join(request.Dir, agentsDir)); err != nil {
		return domain.StepResult{}, fmt.Errorf("fetch the plugin at %s: %w", version, err)
	}

	return domain.PluginStep.Done(fmt.Sprintf(
		"installed codefall's skills at %s into %s", version, agentsFullName)), nil
}

// installedPluginVersion is the release the mirror in .agents/ came from, read out of the plugin's
// own manifest rather than asked of any harness. A directory with nothing to say says nothing —
// the install then runs — and a file that cannot be read or is not a manifest is an error, because
// the step is about to write over it. A manifest without a version is the same as none at all.
func (i *Initialize) installedPluginVersion(dir string) (mo.Option[string], error) {
	path := filepath.Join(dir, agentsDir, pluginManifest)

	data, err := i.files.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return mo.None[string](), nil
	case err != nil:
		return mo.None[string](), fmt.Errorf("read %s%s: %w", agentsFullName, pluginManifest, err)
	}

	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return mo.None[string](), fmt.Errorf("decode %s%s: %w", agentsFullName, pluginManifest, err)
	}

	if manifest.Version == "" {
		return mo.None[string](), nil
	}

	return mo.Some(manifest.Version), nil
}
