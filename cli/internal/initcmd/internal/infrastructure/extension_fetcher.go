package infrastructure

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"sort"

	"github.com/lividlabs/codefall-cli/cli/internal/shared/process"
)

// EmbeddedExtensionFetcher serves the extension tree out of the embedded extensions FS.
type EmbeddedExtensionFetcher struct {
	src   fs.FS
	files *process.FileSystem
}

// NewEmbeddedExtensionFetcher builds the gateway over the tree the binary was compiled with. The
// source fs is injected so tests can drive an in-memory tree instead of the real one.
func NewEmbeddedExtensionFetcher(src fs.FS) *EmbeddedExtensionFetcher {
	return &EmbeddedExtensionFetcher{src: src, files: process.NewFileSystem()}
}

// Fetch mirrors every file in the embedded tree onto destDir, except anything under the excluded
// prefixes, and returns the relative paths it wrote (a manifest-of-one-copy) sorted so two
// sequential runs produce the same record.
func (f *EmbeddedExtensionFetcher) Fetch(ctx context.Context, destDir string, exclude []string) ([]string, error) {
	var installed []string

	err := fs.WalkDir(f.src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		for _, prefix := range exclude {
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
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

		// The hooks the definitions register are invoked by path, so a copied script has to be
		// runnable where it lands: the embedded tree carries no modes to copy.
		if strings.HasSuffix(path, ".sh") {
			if err := f.files.MakeExecutable(target); err != nil {
				return fmt.Errorf("make %s executable: %w", target, err)
			}
		}

		installed = append(installed, path)
		return nil
	})

	sort.Strings(installed)
	return installed, err
}

// Read returns one file from the embedded tree: what the hook step consumes per harness.
func (f *EmbeddedExtensionFetcher) Read(path string) ([]byte, error) {
	return fs.ReadFile(f.src, path)
}
