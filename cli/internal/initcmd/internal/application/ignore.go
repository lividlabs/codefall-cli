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

// ignore is the step that keeps codefall-review's findings out of codebase search.
//
// The file may already be the project's own, holding entries that have nothing to do with codefall,
// so this step appends rather than writes: overwriting a file that has drifted is the one thing the
// extension's rules never allow. A file that already names the directory is left exactly as it is,
// which is what makes a rerun a no-op rather than a growing list of duplicates.
func (i *Initialize) ignore(_ context.Context, request Request) (domain.StepResult, error) {
	path := filepath.Join(request.Dir, settings.IgnoreName)

	existing, err := i.files.ReadFile(path)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		body := settings.IgnoreComment + "\n" + settings.IgnoreEntry + "\n"
		if err := i.files.WriteFile(path, []byte(body)); err != nil {
			return domain.StepResult{}, fmt.Errorf("write %s: %w", settings.IgnoreName, err)
		}

		return domain.IgnoreStep.Done("wrote " + settings.IgnoreName), nil

	case err != nil:
		return domain.StepResult{}, fmt.Errorf("read %s: %w", settings.IgnoreName, err)
	}

	if settings.IgnoresReviews(string(existing)) {
		return domain.IgnoreStep.Skipped(
			fmt.Sprintf("%s already names %s", settings.IgnoreName, settings.IgnoreEntry)), nil
	}

	// A file that does not end in a newline would otherwise have its last entry joined to ours.
	body := string(existing)
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	body += "\n" + settings.IgnoreComment + "\n" + settings.IgnoreEntry + "\n"

	if err := i.files.WriteFile(path, []byte(body)); err != nil {
		return domain.StepResult{}, fmt.Errorf("write %s: %w", settings.IgnoreName, err)
	}

	return domain.IgnoreStep.Done(
		fmt.Sprintf("added %s to %s", settings.IgnoreEntry, settings.IgnoreName)), nil
}
