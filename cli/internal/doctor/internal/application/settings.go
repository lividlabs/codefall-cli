package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/domain"
	"github.com/lividlabs/codefall-cli/internal/shared/settings"
)

// settings runs checks 1 to 4. Each one is the prerequisite of the next, so the first failure ends
// the group and the remaining checks are absent from the report.
//
// Every detail is a whole sentence, because the report prints it under a category header without the
// check's title in front of it.
//
// Decoding happens here rather than in infrastructure, which returns bytes, or in the domain, which
// must not name an encoding. It decodes into the generic document so the shared field tables stay
// the single definition of the settings shape and every problem is reported at once.
func (d *Diagnose) settings(_ context.Context, dir string, results []domain.Result) []domain.Result {
	codefallDir := filepath.Join(dir, ".codefall")
	settingsPath := filepath.Join(codefallDir, "settings.json")
	createRemedy := mo.Some("create .codefall/settings.json; schema: " + settings.SchemaID)

	exists, err := d.files.DirExists(codefallDir)

	switch {
	case err != nil:
		return append(results, domain.CodefallDir.Fail(
			fmt.Sprintf("Cannot stat %s: %v", codefallDir, err), mo.None[string]()))
	case !exists:
		return append(results, domain.CodefallDir.Fail(".codefall/ not found", createRemedy))
	}

	results = append(results, domain.CodefallDir.Pass())

	data, err := d.files.ReadFile(settingsPath)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return append(results, domain.SettingsFile.Fail(".codefall/settings.json not found", createRemedy))
	case err != nil:
		return append(results, domain.SettingsFile.Fail(
			fmt.Sprintf("Cannot read .codefall/settings.json: %v", err), mo.None[string]()))
	}

	results = append(results, domain.SettingsFile.Pass())

	fixRemedy := mo.Some("fix the fields above; schema: " + settings.SchemaID)

	var doc settings.Document

	if err := json.Unmarshal(data, &doc); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return append(results, domain.SettingsJSON.Fail(
				fmt.Sprintf("settings.json is not valid JSON at byte %d: %v", syntaxErr.Offset, syntaxErr),
				mo.None[string]()))
		}

		// Well-formed JSON that is not an object parses cleanly and is simply the wrong shape, so
		// check 3 passes and check 4 carries the complaint.
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			results = append(results, domain.SettingsJSON.Pass())

			return append(results, domain.SettingsComplete.Fail(
				"settings.json's top level must be a JSON object", fixRemedy))
		}

		return append(results, domain.SettingsJSON.Fail(
			fmt.Sprintf("settings.json is not valid JSON: %v", err), mo.None[string]()))
	}

	results = append(results, domain.SettingsJSON.Pass())

	if problems := settings.Validate(doc); len(problems) > 0 {
		return append(results, domain.SettingsComplete.Fail(
			"settings.json is incomplete: "+strings.Join(problems, "; "), fixRemedy))
	}

	return append(results, domain.SettingsComplete.Pass())
}
