package infrastructure

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"sort"

	"github.com/lividlabs/codefall-cli/cli/internal/shared/process"
)

// EmbeddedExtensionFetcher copies the extension tree out of the embedded extensions FS.
type EmbeddedExtensionFetcher struct {
	src   fs.FS
	files *process.FileSystem
}

// NewEmbeddedExtensionFetcher builds the fetch gateway over the tree the binary was compiled with.
// The source fs is injected so tests can drive an in-memory tree instead of the real one.
func NewEmbeddedExtensionFetcher(src fs.FS) *EmbeddedExtensionFetcher {
	return &EmbeddedExtensionFetcher{src: src, files: process.NewFileSystem()}
}

// Fetch mirrors every file in the embedded tree onto destDir, and returns the relative paths it
// wrote (a manifest-of-one-copy) sorted so two sequential runs produce the same record.
func (f *EmbeddedExtensionFetcher) Fetch(ctx context.Context, destDir string) ([]string, error) {
	var installed []string

	err := fs.WalkDir(f.src, ".", func(path string, d fs.DirEntry, err error) error {
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

		installed = append(installed, path)
		return nil
	})

	sort.Strings(installed)
	return installed, err
}
