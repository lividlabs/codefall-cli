package application

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/shared/manifest"
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

// Installation is what finished runs recorded, as presentation reads it back: the version each
// harness was installed at. The gate compares it per harness, because a current version for one
// harness says nothing about a project that has never been set up for the harness in front of it
// (ADR-001).
type Installation struct {
	Versions map[string]string
}

// Installed is the manifest through the use-case boundary, so presentation can compare it with the
// binary's own tag without knowing the manifest's path. A manifest that records no usable version at
// all reads the same as no manifest (ADR-GO-03).
func (i *Initialize) Installed(dir string) (mo.Option[Installation], error) {
	recorded, read, err := i.recordedManifest(dir)
	if err != nil {
		return mo.None[Installation](), err
	}

	if !read {
		return mo.None[Installation](), nil
	}

	versions := recorded.Versions()
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
	recorded, _, err := i.recordedManifest(dir)
	if err != nil {
		return err
	}

	if recorded.Harnesses == nil {
		recorded.Harnesses = map[string]manifest.Install{}
	}

	for name, files := range installed {
		recorded.Harnesses[name] = manifest.Install{Version: version, Files: files}
	}

	body, err := manifest.Encode(recorded)
	if err != nil {
		return err
	}

	if err := i.files.WriteFile(filepath.Join(dir, manifest.Name), body); err != nil {
		return fmt.Errorf("write %s: %w", manifest.Name, err)
	}

	return nil
}

// recordedManifest is what the file already says. A file that is not there is not an error and not a
// record either, because this is what creates it — so the second return says whether there was one to
// read, which the caller needs when the difference matters (ADR-GO-03).
func (i *Initialize) recordedManifest(dir string) (manifest.Document, bool, error) {
	data, err := i.files.ReadFile(filepath.Join(dir, manifest.Name))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return manifest.Document{}, false, nil
	case err != nil:
		return manifest.Document{}, false, fmt.Errorf("read %s: %w", manifest.Name, err)
	}

	recorded, err := manifest.Decode(data)
	if err != nil {
		return manifest.Document{}, false, err
	}

	return recorded, true, nil
}
