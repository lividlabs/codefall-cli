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

// skillsDirExtension is the extension step for every harness — each own directory comes from
// extensionDestDirs. It copies the embedded extension tree, minus the per-harness hook definitions
// the hook step consumes, and hands back what it wrote for the manifest the run records at the end.
func (i *Initialize) skillsDirExtension(ctx context.Context, request Request) (domain.StepResult, []string, error) {
	dest := extensionDestDirs[request.Harness]

	installed, err := i.source.Fetch(ctx, filepath.Join(request.Dir, dest), hookSourceDirs)
	if err != nil {
		return domain.StepResult{}, nil, fmt.Errorf("install the embedded extension: %w", err)
	}

	return domain.ExtensionStep.Done(fmt.Sprintf(
		"installed codefall's skills into %s/", dest)), installed, nil
}

// manifest is the install record. Files are relative to the harness's skills directory, sorted,
// and the version is the binary that wrote them, so "from/to" is readable both ways.
type manifest struct {
	Harness string   `json:"harness"`
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

// InstalledVersion is the manifest's version report through the use-case boundary, so presentation
// can compare it with the binary's own tag without knowing the manifest's path.
func (i *Initialize) InstalledVersion(dir string) (mo.Option[string], error) {
	path := filepath.Join(dir, domain.ManifestName)

	data, err := i.files.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return mo.None[string](), nil
	case err != nil:
		return mo.None[string](), fmt.Errorf("read %s: %w", domain.ManifestName, err)
	}

	var previous manifest
	if err := json.Unmarshal(data, &previous); err != nil {
		return mo.None[string](), fmt.Errorf("decode %s: %w", domain.ManifestName, err)
	}

	if previous.Version == "" {
		return mo.None[string](), nil
	}

	return mo.Some(previous.Version), nil
}

// writeManifest writes .codefall/manifest.json in the project's directory. The run calls it once
// every step has succeeded: the file says an install of this version for this harness is complete,
// and the upgrade gate takes it at its word.
func (i *Initialize) writeManifest(dir, harness, version string, files []string) error {
	body, err := json.MarshalIndent(manifest{Harness: harness, Version: version, Files: files},
		"", "  ")
	if err != nil {
		return err
	}

	if err := i.files.WriteFile(filepath.Join(dir, domain.ManifestName), body); err != nil {
		return fmt.Errorf("write %s: %w", domain.ManifestName, err)
	}

	return nil
}
