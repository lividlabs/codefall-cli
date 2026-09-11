package infrastructure

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/process"
)

// CodeloadPluginFetcher retrieves the plugin's release from GitHub's tarball service and mirrors
// the plugin's tree into the project's directory. The URL shape and the root directory a tarball
// arrives wrapped in are codeload's conventions; a tag being spelled "v" + version and the plugin
// tree living under plugins/codefall/ are the plugin repository's conventions.
type CodeloadPluginFetcher struct {
	host   string
	client *http.Client
	files  *process.FileSystem
}

// NewCodeloadPluginFetcher builds the real fetch gateway.
func NewCodeloadPluginFetcher() *CodeloadPluginFetcher {
	return &CodeloadPluginFetcher{
		host:   "https://codeload.github.com",
		client: http.DefaultClient,
		files:  process.NewFileSystem(),
	}
}

// Fetch downloads the plugin's release tarball and rebases every regular file under
// plugins/codefall/ onto destDir. Entries outside the plugin tree, and headers that are not
// regular files, are skipped: the mirror is of the plugin's files and nothing else.
func (f *CodeloadPluginFetcher) Fetch(ctx context.Context, version, destDir string) error {
	ref := "v" + version

	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		f.host+"/"+domain.MarketplaceSource+"/tar.gz/refs/tags/"+ref, nil)
	if err != nil {
		return fmt.Errorf("build the %s request: %w", ref, err)
	}

	response, err := f.client.Do(request)
	if err != nil {
		return fmt.Errorf("GET %s: %w", ref, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", ref, response.StatusCode)
	}

	uncompressed, err := gzip.NewReader(response.Body)
	if err != nil {
		return fmt.Errorf("open the %s tarball: %w", ref, err)
	}
	defer func() { _ = uncompressed.Close() }()

	entries := tar.NewReader(uncompressed)

	for {
		header, err := entries.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read the %s tarball: %w", ref, err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		rebased, plugin := rebase(header.Name)
		if !plugin {
			continue
		}

		target := filepath.Join(destDir, rebased)

		if err := f.files.MkdirAll(filepath.Dir(target)); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(target), err)
		}

		data, err := io.ReadAll(entries)
		if err != nil {
			return fmt.Errorf("read %s: %w", header.Name, err)
		}

		if err := f.files.WriteFile(target, data); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}
	}

	return nil
}

// rebase returns the path of an entry relative to the plugin tree — its codeload root directory
// removed and its plugins/codefall/ prefix confirmed — or no path when the entry is not under the
// plugin tree at all. A name that would walk out of the tree is treated as not one of the plugin's,
// because a tarball is read from the network whatever its origin says.
func rebase(name string) (string, bool) {
	parts := strings.Split(name, "/")
	if len(parts) < 4 || parts[1] != "plugins" || parts[2] != "codefall" {
		return "", false
	}

	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}

	return strings.Join(parts[3:], "/"), true
}
