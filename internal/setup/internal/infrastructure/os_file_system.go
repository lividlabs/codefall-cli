// Package infrastructure implements setup's gateways against the real process: the file system it
// writes into and the external tools codefall depends on.
package infrastructure

import (
	"io/fs"
	"os"
)

// The permissions init creates things with. A repository's configuration is not a secret and is read
// by the tools codefall drives, so the directory is traversable and the file is world-readable — the
// modes git itself uses for a checkout. The use case does not choose them: it asks for a directory
// and a file, and this is where the operating system's terms are decided.
const (
	dirMode  fs.FileMode = 0o755
	fileMode fs.FileMode = 0o644
)

// OSFileSystem reads and writes the working directory through the operating system.
type OSFileSystem struct{}

// NewOSFileSystem builds the real file system gateway.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadFile returns a file's bytes. os.ReadFile's *PathError already satisfies
// errors.Is(err, fs.ErrNotExist), which is what the use case branches on.
func (*OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// MkdirAll creates path and every parent it needs. An existing directory is success, which is what
// makes init safe to run again.
func (*OSFileSystem) MkdirAll(path string) error {
	return os.MkdirAll(path, dirMode)
}

// WriteFile writes data to path, replacing whatever was there. The mode applies only to a file this
// call creates; the permissions of one that already exists are left alone, as os.WriteFile has it.
func (*OSFileSystem) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, fileMode)
}
