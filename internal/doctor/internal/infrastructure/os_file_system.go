// Package infrastructure implements doctor's gateways against the real process: the file system and
// the external tools codefall depends on.
package infrastructure

import (
	"errors"
	"io/fs"
	"os"
)

// OSFileSystem reads the working directory through the operating system.
type OSFileSystem struct{}

// NewOSFileSystem builds the real file system gateway.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// DirExists reports whether path is a directory. A path that is not there is not an error — that is
// the answer the caller asked for.
func (*OSFileSystem) DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return info.IsDir(), nil
}

// ReadFile returns a file's bytes. os.ReadFile's *PathError already satisfies
// errors.Is(err, fs.ErrNotExist), which is what the use case branches on.
func (*OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
