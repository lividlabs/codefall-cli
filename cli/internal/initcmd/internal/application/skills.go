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

// fetchSkills copies the embedded extension tree into one skills directory, minus the per-harness
// hook definitions the hook step consumes, and hands back what it wrote relative to that directory.
func (i *Initialize) fetchSkills(ctx context.Context, dir, dest string) ([]string, error) {
	installed, err := i.source.Fetch(ctx, filepath.Join(dir, dest), hookSourceDirs)
	if err != nil {
		return nil, fmt.Errorf("install the embedded extension: %w", err)
	}

	return installed, nil
}

// manifest is the install record: what each harness has installed, and which binary installed it.
//
// A run updates the entries for the harnesses it installed for and leaves every other entry as it
// was. A project set up for two harnesses therefore keeps both records, where a file naming one
// harness lost the first record as soon as the second install finished — and the next run for that
// first harness repeated work it had already done.
type manifest struct {
	Harnesses map[string]harnessInstall `json:"harnesses"`
}

// harnessInstall is one harness's entry: the binary that wrote its files, and the files, relative to
// the directory init installed in.
type harnessInstall struct {
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

// Installation is what finished runs recorded, as presentation reads it back: the version each
// harness was installed at. The gate compares it per harness, because a current version for one
// harness says nothing about a project that has never been set up for the harness in front of it
// (ADR-001).
type Installation struct {
	Versions map[string]string
}

// Installed is the manifest through the use-case boundary, so presentation can compare it with the
// binary's own tag without knowing the manifest's path. A harness recorded with no version records
// nothing a comparison can use, so it is left out; a manifest that records no usable version at all
// reads the same as no manifest (ADR-GO-03).
func (i *Initialize) Installed(dir string) (mo.Option[Installation], error) {
	data, err := i.files.ReadFile(filepath.Join(dir, domain.ManifestName))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return mo.None[Installation](), nil
	case err != nil:
		return mo.None[Installation](), fmt.Errorf("read %s: %w", domain.ManifestName, err)
	}

	previous, err := decodeManifest(data)
	if err != nil {
		return mo.None[Installation](), err
	}

	versions := map[string]string{}

	for name, install := range previous.Harnesses {
		if install.Version != "" {
			versions[name] = install.Version
		}
	}

	if len(versions) == 0 {
		return mo.None[Installation](), nil
	}

	return mo.Some(Installation{Versions: versions}), nil
}

// writeManifest records this run in .codefall/manifest.json. The run calls it once every step has
// succeeded: the entries it writes say an install of this version, for those harnesses, is complete,
// and the upgrade gate takes them at their word.
//
// It merges into what is already recorded rather than replacing it. A run for one harness has done
// nothing to another harness's install and has no business erasing the record of it.
func (i *Initialize) writeManifest(dir, version string, installed map[string][]string) error {
	recorded, err := i.recordedManifest(dir)
	if err != nil {
		return err
	}

	for name, files := range installed {
		recorded.Harnesses[name] = harnessInstall{Version: version, Files: files}
	}

	body, err := json.MarshalIndent(recorded, "", "  ")
	if err != nil {
		return err
	}

	if err := i.files.WriteFile(filepath.Join(dir, domain.ManifestName), body); err != nil {
		return fmt.Errorf("write %s: %w", domain.ManifestName, err)
	}

	return nil
}

// recordedManifest is what the file already says, with an entry map ready to write into. A file that
// is not there is an empty record rather than an error, because this is what creates it.
func (i *Initialize) recordedManifest(dir string) (manifest, error) {
	recorded := manifest{Harnesses: map[string]harnessInstall{}}

	data, err := i.files.ReadFile(filepath.Join(dir, domain.ManifestName))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return recorded, nil
	case err != nil:
		return manifest{}, fmt.Errorf("read %s: %w", domain.ManifestName, err)
	}

	previous, err := decodeManifest(data)
	if err != nil {
		return manifest{}, err
	}

	for name, install := range previous.Harnesses {
		recorded.Harnesses[name] = install
	}

	return recorded, nil
}

// decodeManifest reads the record. A manifest written before harnesses were recorded per install
// names none of them, which decodes cleanly to an empty record: the run then repeats every step,
// which they are all built to tolerate.
func decodeManifest(data []byte) (manifest, error) {
	var previous manifest
	if err := json.Unmarshal(data, &previous); err != nil {
		return manifest{}, fmt.Errorf("decode %s: %w", domain.ManifestName, err)
	}

	return previous, nil
}
