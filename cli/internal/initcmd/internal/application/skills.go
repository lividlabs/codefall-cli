package application

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// skillsDirPlugin is the plugin step for every harness — each own directory comes from
// pluginDestDirs. It copies the embedded plugin tree, and records the file list in the install
// manifest, which later commands (upgrade, drift checks) can trust to state ownership.
func (i *Initialize) skillsDirPlugin(ctx context.Context, request Request) (domain.StepResult, error) {
	dest := pluginDestDirs[request.Harness]

	installed, err := i.fetcher.Fetch(ctx, filepath.Join(request.Dir, dest))
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("install the embedded plugin: %w", err)
	}

	if err := i.writeManifest(request.Dir, request.Harness, installed); err != nil {
		return domain.StepResult{}, fmt.Errorf("record the installation to %s: %w",
			domain.ManifestName, err)
	}

	return domain.PluginStep.Done(fmt.Sprintf(
		"installed codefall's skills into %s/ and recorded them to %s",
		dest, domain.ManifestName)), nil
}

// manifest is the record of what a run installed under which harness. Files are relative to the
// skills directory, sorted, so the diff between two runs names the change it would be.
type manifest struct {
	Harness string   `json:"harness"`
	Files   []string `json:"files"`
}

// writeManifest writes .codefall/manifest.json in the project's directory; the settings step
// already created the directory in this run. Its row over the plain JSON body makes a clobbered
// install rebuildable by rerun.
func (i *Initialize) writeManifest(dir, harness string, files []string) error {
	body, err := json.MarshalIndent(manifest{Harness: harness, Files: files}, "", "  ")
	if err != nil {
		return err
	}

	if err := i.files.WriteFile(filepath.Join(dir, domain.ManifestName), body); err != nil {
		return fmt.Errorf("write %s: %w", domain.ManifestName, err)
	}

	return nil
}
