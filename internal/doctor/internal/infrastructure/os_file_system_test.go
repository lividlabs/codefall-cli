package infrastructure

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/application"
)

var _ application.FileSystem = (*OSFileSystem)(nil)

func TestOSFileSystemDirExists(t *testing.T) {
	root := t.TempDir()

	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	files := NewOSFileSystem()

	for _, tc := range []struct {
		name string
		path string
		want bool
	}{
		{"a directory", nested, true},
		{"a file is not a directory", file, false},
		{"a missing path is not an error", filepath.Join(root, "nope"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := files.DirExists(tc.path)
			if err != nil {
				t.Fatalf("DirExists: %v", err)
			}

			if got != tc.want {
				t.Errorf("DirExists(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

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

	// The use case branches on this exact sentinel, so it is the behaviour worth pinning.
	if _, err := files.ReadFile(filepath.Join(root, "missing.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadFile of a missing file = %v, want an fs.ErrNotExist", err)
	}
}
