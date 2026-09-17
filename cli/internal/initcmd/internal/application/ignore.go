package application

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

// ignoreEntry is one line codefall needs in one ignore file: the file, the entry, the comment that
// says why it is there, and the shared module's reading of whether the file already has it.
type ignoreEntry struct {
	file    string
	entry   string
	comment string
	names   func(body string) bool
}

// ignoreEntries is what the step writes, in the order it reports them. The .ignore entry keeps
// codefall-review's findings out of codebase search; the .gitignore entry keeps the refresh stamp,
// a per-machine file, out of the repository (ADR-005).
var ignoreEntries = []ignoreEntry{
	{file: settings.IgnoreName, entry: settings.IgnoreEntry, comment: settings.IgnoreComment, names: settings.IgnoresReviews},
	{file: settings.GitIgnoreName, entry: settings.RefreshStamp, comment: settings.GitIgnoreComment, names: settings.IgnoresStamp},
}

// ignore is the step that writes the two ignore entries.
//
// Either file may already be the project's own, holding entries that have nothing to do with
// codefall, so this step appends rather than writes: overwriting a file that has drifted is the one
// thing the extension's rules never allow. A file that already names its entry is left exactly as
// it is, which is what makes a rerun a no-op rather than a growing list of duplicates.
func (i *Initialize) ignore(_ context.Context, request Request) (domain.StepResult, error) {
	var done []string

	for _, entry := range ignoreEntries {
		did, err := i.ensureIgnored(request.Dir, entry)
		if err != nil {
			return domain.StepResult{}, err
		}

		if did != "" {
			done = append(done, did)
		}
	}

	if len(done) == 0 {
		return domain.IgnoreStep.Skipped(fmt.Sprintf("%s already names %s and %s already names %s",
			settings.IgnoreName, settings.IgnoreEntry, settings.GitIgnoreName, settings.RefreshStamp)), nil
	}

	return domain.IgnoreStep.Done(sentenceList(done)), nil
}

// ensureIgnored puts one entry in its file and says what that took, or "" when the file already
// had it.
func (i *Initialize) ensureIgnored(dir string, entry ignoreEntry) (string, error) {
	path := filepath.Join(dir, entry.file)

	existing, err := i.files.ReadFile(path)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		body := entry.comment + "\n" + entry.entry + "\n"
		if err := i.files.WriteFile(path, []byte(body)); err != nil {
			return "", fmt.Errorf("write %s: %w", entry.file, err)
		}

		return "wrote " + entry.file, nil

	case err != nil:
		return "", fmt.Errorf("read %s: %w", entry.file, err)
	}

	if entry.names(string(existing)) {
		return "", nil
	}

	// A file that does not end in a newline would otherwise have its last entry joined to ours.
	body := string(existing)
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	body += "\n" + entry.comment + "\n" + entry.entry + "\n"

	if err := i.files.WriteFile(path, []byte(body)); err != nil {
		return "", fmt.Errorf("write %s: %w", entry.file, err)
	}

	return fmt.Sprintf("added %s to %s", entry.entry, entry.file), nil
}
