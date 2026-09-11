package process

import (
	"errors"
	"io/fs"
	"os"
)

// The permissions codefall creates things with. A repository's configuration is not a secret and is
// read by the tools codefall drives, so a directory is traversable and a file is world-readable —
// the modes git itself uses for a checkout. A use case does not choose them: it asks for a directory
// and a file, and this is where the operating system's terms are decided.
const (
	dirMode  fs.FileMode = 0o755
	fileMode fs.FileMode = 0o644
)

// FileSystem reads and writes the working directory through the operating system. It is the union of
// what the commands need; each one's gateway exposes only its own half.
type FileSystem struct{}

// NewFileSystem builds the real file system.
func NewFileSystem() *FileSystem {
	return &FileSystem{}
}

// DirExists reports whether path is a directory. A path that is not there is not an error — that is
// the answer the caller asked for.
func (*FileSystem) DirExists(path string) (bool, error) {
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
// errors.Is(err, fs.ErrNotExist), which is what the use cases branch on.
func (*FileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// MkdirAll creates path and every parent it needs. An existing directory is success, which is what
// makes init safe to run again.
func (*FileSystem) MkdirAll(path string) error {
	return os.MkdirAll(path, dirMode)
}

// WriteFile writes data to path, replacing whatever was there. The mode applies only to a file this
// call creates; the permissions of one that already exists are left alone, as os.WriteFile has it.
func (*FileSystem) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, fileMode)
}

// MakeExecutable marks a copied script runnable. WriteFile's mode is deliberately not executable —
// most of what codefall writes is configuration — so the step that installs a script says so here,
// rather than the use case choosing modes it was promised not to think about.
func (*FileSystem) MakeExecutable(path string) error {
	return os.Chmod(path, dirMode)
}
