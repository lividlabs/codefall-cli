package application

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/doctor/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/harness"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/manifest"
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
	// a second complaint about the same file would be noise. The checks are absent from the report
	// rather than present with a status of their own, which is what every other skipped check does.
	if !settingsAreComplete(results) {
		return results
	}

	chosen, recorded := d.chosenHarnesses(dir)
	if !recorded {
		return results
	}

	return d.leftOver(dir, chosen, d.installed(dir, chosen, results))
}

// installed is check 6: codefall's extension is where each chosen harness reads it.
func (d *Diagnose) installed(dir string, chosen []string, results []domain.Result) []domain.Result {
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

// leftOver is check 7: nothing codefall installed is still sitting there for a harness the settings
// no longer name. A project set up for two harnesses that later drops one keeps everything codefall
// wrote for it, because codefall only ever writes what it owns and never deletes.
//
// It warns rather than fails, because nothing stops working. What goes wrong is quieter: an agent
// reading the skills directory finds codefall's skills for a harness the project has dropped, and the
// hook codefall registered for that harness still runs.
//
// The remedy points at the manifest rather than naming a directory to delete. The manifest is what
// says which files codefall wrote for that harness, and harnesses share directories — four of the
// five read .agents/ — so the directory a dropped harness read may still be another's. Which of those
// files are safe to remove is the reader's judgement, not doctor's.
func (d *Diagnose) leftOver(dir string, chosen []string, results []domain.Result) []domain.Result {
	// A manifest that is missing, unreadable, or not a manifest any more records no install to be
	// left over. Init fails on the same file the next time it writes one, so nothing goes unsaid.
	recorded, read := d.recordedManifest(dir)
	if !read {
		return results
	}

	var dropped []string

	for _, name := range recorded.Recorded() {
		if !slices.Contains(chosen, name) {
			dropped = append(dropped, name)
		}
	}

	if len(dropped) == 0 {
		return append(results, domain.HarnessesLeftOver.Pass())
	}

	return append(results, domain.HarnessesLeftOver.Warn(
		fmt.Sprintf("codefall is still installed for %s, which the settings no longer name",
			strings.Join(dropped, ", ")),
		mo.Some(fmt.Sprintf("remove the files %s lists for %s; codefall never deletes what it wrote",
			manifest.Name, strings.Join(dropped, ", ")))))
}

// recordedManifest is what a finished run left on record. Anything the reader cannot make sense of is
// no record rather than a second complaint: doctor has no check for the manifest's own shape, and the
// install check above has already reported whatever is missing on disk.
func (d *Diagnose) recordedManifest(dir string) (manifest.Document, bool) {
	data, err := d.files.ReadFile(filepath.Join(dir, manifest.Name))
	if err != nil {
		return manifest.Document{}, false
	}

	recorded, err := manifest.Decode(data)
	if err != nil {
		return manifest.Document{}, false
	}

	return recorded, true
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
