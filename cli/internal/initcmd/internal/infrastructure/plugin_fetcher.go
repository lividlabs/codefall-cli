package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/lividlabs/codefall-cli/cli/internal/shared/process"
)

// EmbeddedPluginFetcher copies the plugin tree out of the embedded extensions FS. Versioning lives
// in the same tree, in the plugin manifest, where it also lives after install.
type EmbeddedPluginFetcher struct {
	src   fs.FS
	files *process.FileSystem
}

// NewEmbeddedPluginFetcher builds the fetch gateway over the tree the binary was compiled with.
// The source fs is injected so tests can drive an in-memory tree instead of the real one.
func NewEmbeddedPluginFetcher(src fs.FS) *EmbeddedPluginFetcher {
	return &EmbeddedPluginFetcher{src: src, files: process.NewFileSystem()}
}

// pluginManifest is the version file's path inside the plugin tree, before and after install.
const pluginManifest = ".claude-plugin/plugin.json"

// Version returns the plugin release the embedded tree belongs to, read from the manifest before
// anything is copied — the same check the install does.
func (f *EmbeddedPluginFetcher) Version() (string, error) {
	data, err := fs.ReadFile(f.src, pluginManifest)
	if err != nil {
		return "", fmt.Errorf("read %s from the embedded tree: %w", pluginManifest, err)
	}

	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", fmt.Errorf("decode %s from the embedded tree: %w", pluginManifest, err)
	}

	return manifest.Version, nil
}

// Fetch mirrors every file in the embedded tree onto destDir. Entries are files and empty dirs in
// an embed.FS, nothing else — the walk is a copy and nothing else, with the manifest path visible.
func (f *EmbeddedPluginFetcher) Fetch(ctx context.Context, destDir string) error {
	return fs.WalkDir(f.src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(f.src, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		target := filepath.Join(destDir, path)

		if err := f.files.MkdirAll(filepath.Dir(target)); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(target), err)
		}

		if err := f.files.WriteFile(target, data); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}

		return nil
	})
}
