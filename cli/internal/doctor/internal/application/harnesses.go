package application

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/doctor/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/harness"
)

// initRemedy is what to do about a harness the project chose and codefall was never run for.
const initRemedy = "codefall init"

// harnesses runs check 6: codefall's extension is installed for every harness the settings record.
//
// What it stats is the directory the shared hook scripts land in, under each harness's own skills
// directory. Nothing but codefall writes that directory, which is what makes it evidence of an
// install — where the skills directory itself, .claude/ or .agents/, exists in plenty of projects
// that have never run codefall.
//
// It fails rather than warns. A harness the project chose and codefall never installed for has no
// skills and no hooks there, so every codefall verb is missing in a harness somebody is using.
func (d *Diagnose) harnesses(_ context.Context, dir string, results []domain.Result) []domain.Result {
	// Settings doctor has already complained about say nothing about which harnesses were chosen, and
	// a second complaint about the same file would be noise. The check is absent from the report
	// rather than present with a status of its own, which is what every other skipped check does.
	if !settingsAreComplete(results) {
		return results
	}

	chosen, recorded := d.chosenHarnesses(dir)
	if !recorded {
		return results
	}

	var missing []string

	for _, name := range chosen {
		shared, known := harness.SharedHooksDir(name).Get()
		if !known {
			// A name the settings-complete check has already refused.
			continue
		}

		path := filepath.Join(dir, shared)

		installed, err := d.files.DirExists(path)
		if err != nil {
			return append(results, domain.HarnessesInstalled.Fail(
				fmt.Sprintf("Cannot stat %s: %v", path, err), mo.None[string]()))
		}

		if !installed {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return append(results, domain.HarnessesInstalled.Fail(
			"codefall is not installed for "+strings.Join(missing, ", "), mo.Some(initRemedy)))
	}

	return append(results, domain.HarnessesInstalled.PassWithDetail(strings.Join(chosen, ", ")))
}

// settingsAreComplete reports whether the settings group found the file complete. A check whose
// prerequisite failed is skipped, and the settings this check reads are that prerequisite.
func settingsAreComplete(results []domain.Result) bool {
	for _, result := range results {
		if result.Check.ID == domain.SettingsComplete.ID {
			return result.Status == domain.StatusPass
		}
	}

	return false
}

// chosenHarnesses is the harnesses the settings record. The file has already been read and found
// complete by the settings group, so anything unexpected here is nothing to report a second time —
// it is simply no answer, and the check is skipped.
func (d *Diagnose) chosenHarnesses(dir string) ([]string, bool) {
	data, err := d.files.ReadFile(filepath.Join(dir, ".codefall", "settings.json"))
	if err != nil {
		return nil, false
	}

	var document struct {
		Harnesses []string `json:"harnesses"`
	}

	if err := json.Unmarshal(data, &document); err != nil || len(document.Harnesses) == 0 {
		return nil, false
	}

	return document.Harnesses, true
}
