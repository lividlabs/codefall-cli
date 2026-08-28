package infrastructure

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/application"
)

var _ application.FileSystem = (*OSFileSystem)(nil)

func TestOSFileSystemReadFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.json")

	if err := os.WriteFile(path, []byte(`{"version":1}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	files := NewOSFileSystem()

	data, err := files.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(data) != `{"version":1}` {
		t.Errorf("ReadFile = %q, want %q", data, `{"version":1}`)
	}

	// The use case branches on this exact sentinel to decide whether settings are already there, so
	// it is the behaviour worth pinning.
	if _, err := files.ReadFile(filepath.Join(root, "missing.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadFile of a missing file = %v, want an fs.ErrNotExist", err)
	}
}

// The permissions are the reason this gateway exists rather than the use case calling os itself, so
// they are asserted here. The umask can only take bits away, and the modes below are what a normal
// umask of 022 leaves alone.
func TestOSFileSystemMkdirAllCreatesATraversableDirectory(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", ".codefall")

	files := NewOSFileSystem()

	if err := files.MkdirAll(nested); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	info, err := os.Stat(nested)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("MkdirAll made a %v, want a directory", info.Mode())
	}

	if got := info.Mode().Perm(); got != dirMode {
		t.Errorf("mode = %v, want %v", got, dirMode)
	}

	// Running init twice must not fail on the directory it made the first time.
	if err := files.MkdirAll(nested); err != nil {
		t.Errorf("MkdirAll on an existing directory = %v, want nil", err)
	}
}

func TestOSFileSystemWriteFileCreatesAndReplaces(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	files := NewOSFileSystem()

	if err := files.WriteFile(path, []byte("first\n")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if got := info.Mode().Perm(); got != fileMode {
		t.Errorf("mode = %v, want %v", got, fileMode)
	}

	// --force rewrites the file, so the second write must leave nothing of the first behind.
	if err := files.WriteFile(path, []byte("second\n")); err != nil {
		t.Fatalf("WriteFile again: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(data) != "second\n" {
		t.Errorf("contents = %q, want %q", data, "second\n")
	}
}

func TestOSFileSystemWriteFileReportsAMissingDirectory(t *testing.T) {
	files := NewOSFileSystem()

	if err := files.WriteFile(filepath.Join(t.TempDir(), "nope", "settings.json"), nil); err == nil {
		t.Error("WriteFile into a missing directory = nil, want an error")
	}
}
