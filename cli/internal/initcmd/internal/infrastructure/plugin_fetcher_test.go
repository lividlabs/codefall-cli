package infrastructure

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/application"
)

var _ application.PluginFetcher = (*CodeloadPluginFetcher)(nil)

// archiveServer answers with a tar.gz built from the test's entries, and remembers the path it
// was asked for.
type archiveServer struct {
	*httptest.Server
	requested []string
}

// serveArchive hands the repository back to the fetcher, from a running server. The name shape
// mirrors what codeload gives back: a root directory wrapped around the repository.
func serveArchive(t *testing.T, entries []archiveEntry) *archiveServer {
	t.Helper()

	var body bytes.Buffer

	gzipped := gzip.NewWriter(&body)
	writer := tar.NewWriter(gzipped)

	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.path,
			Mode:     0o644,
			Size:     int64(len(entry.content)),
			Typeflag: entry.typeflag,
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatalf("WriteHeader: %v", err)
		}
		if entry.typeflag == tar.TypeReg {
			if _, err := writer.Write([]byte(entry.content)); err != nil {
				t.Fatalf("Write: %v", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := gzipped.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	server := &archiveServer{}

	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.requested = append(server.requested, r.URL.Path)
		if _, err := w.Write(body.Bytes()); err != nil {
			t.Errorf("server write: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

// archiveEntry is one file serveArchive puts into its tarball.
type archiveEntry struct {
	path     string
	content  string
	typeflag byte
}

func testFetcher(host string) *CodeloadPluginFetcher {
	fetcher := NewCodeloadPluginFetcher()
	fetcher.host = host

	return fetcher
}

// The tarball is rebased file by file: the codeload root is stripped, the plugins/codefall/ prefix
// is confirmed, and every regular file lands under the destination with the rest of its path. A
// repository file above the plugin tree, and a directory entry beneath it, are left behind.
func TestFetchInstallsThePluginTree(t *testing.T) {
	server := serveArchive(t, []archiveEntry{
		{"codefall-plugin-0.7.0/", "", tar.TypeDir},
		{"codefall-plugin-0.7.0/README.md", "# the plugin repo\n", tar.TypeReg},
		{"codefall-plugin-0.7.0/plugins/codefall/skills/design/SKILL.md", "---\nname: design\n---\n", tar.TypeReg},
		{"codefall-plugin-0.7.0/plugins/codefall/shared/preflight.sh", "#!/bin/sh\nexit 0\n", tar.TypeReg},
	})

	dir := t.TempDir()

	if err := testFetcher(server.URL).Fetch(t.Context(), "0.7.0", dir); err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if len(server.requested) != 1 ||
		server.requested[0] != "/lividlabs/codefall-plugin/tar.gz/refs/tags/v0.7.0" {
		t.Errorf("requested %v, want one GET of the pinned tag's tarball", server.requested)
	}

	for path, want := range map[string]string{
		"skills/design/SKILL.md": "---\nname: design\n---\n",
		"shared/preflight.sh":    "#!/bin/sh\nexit 0\n",
	} {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
		if err != nil {
			t.Errorf("read %s: %v", path, err)
			continue
		}
		if string(data) != want {
			t.Errorf("%s = %q, want %q", path, data, want)
		}
	}

	// Nothing above the plugin tree is mirrored, and the prefix itself is gone.
	if _, err := os.Stat(filepath.Join(dir, "README.md")); err == nil {
		t.Errorf("README.md present under dest, want only the plugin's tree")
	}

	if _, err := os.Stat(filepath.Join(dir, "plugins")); err == nil {
		t.Errorf("plugins/ present under dest, want the tree rooted at it")
	}
}

// A version that is not a release is reported as what the server answered.
func TestFetchReportedANonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	err := testFetcher(server.URL).Fetch(t.Context(), "9.9.9", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "GET v9.9.9: HTTP 404") {
		t.Errorf("Fetch = %v, want it to name the tag and the status", err)
	}
}

// An entry that would walk out of the tree is not one of the plugin's files, whatever its root
// prefix says.
func TestFetchSkipsAnEntryOutsideTheTree(t *testing.T) {
	server := serveArchive(t, []archiveEntry{
		{"codefall-plugin-0.7.0/plugins/codefall/../../evil.sh", "evil\n", tar.TypeReg},
		{"codefall-plugin-0.7.0/plugins/codefall/skills/ok.md", "ok\n", tar.TypeReg},
	})

	dir := t.TempDir()

	if err := testFetcher(server.URL).Fetch(t.Context(), "0.7.0", dir); err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "evil.sh")); err == nil {
		t.Errorf("evil.sh present, want it refused")
	}

	if _, err := os.Stat(filepath.Join(dir, "skills", "ok.md")); err != nil {
		t.Errorf("skills/ok.md missing after a clean entry: %v", err)
	}
}
