package application

import (
	"context"
	"encoding/json"
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
// it copies the embedded plugin tree under the project's .agents/ and records the file list in the
// install manifest at .codefall/manifest.json, which later commands (upgrade, drift checks) can
// trust to state ownership.
func (i *Initialize) skillsDirPlugin(ctx context.Context, request Request) (domain.StepResult, error) {
	installed, err := i.fetcher.Fetch(ctx, filepath.Join(request.Dir, agentsDir))
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("install the embedded plugin: %w", err)
	}

	if err := i.writeManifest(request.Dir, request.Harness, installed); err != nil {
		return domain.StepResult{}, fmt.Errorf("record the installation to %s: %w",
			domain.ManifestName, err)
	}

	return domain.PluginStep.Done(fmt.Sprintf(
		"installed codefall's skills into %s and recorded them to %s",
		agentsFullName, domain.ManifestName)), nil
}

// manifest is the record of what a run installed under which harness. Files are relative to the
// skills directory, sorted, so the diff between two runs names the change it would be.
type manifest struct {
	Harness string   `json:"harness"`
	Files   []string `json:"files"`
}

// writeManifest writes .codefall/manifest.json in the project's directory; the file is
// mergeable by hand (one JSON object), so a clobbered install is rebuilt cleanly when rerun.
func (i *Initialize) writeManifest(dir, harness string, files []string) error {
	body, err := json.MarshalIndent(manifest{Harness: harness, Files: files}, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, filepath.Dir(domain.ManifestName))
	if err := i.files.MkdirAll(path); err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	if err := i.files.WriteFile(filepath.Join(dir, domain.ManifestName), body); err != nil {
		return fmt.Errorf("write %s: %w", domain.ManifestName, err)
	}

	return nil
}
