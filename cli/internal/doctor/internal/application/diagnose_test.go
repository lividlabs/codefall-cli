package application

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/doctor/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/settings"
)

const workingDir = "/work"

var (
	codefallDir  = filepath.Join(workingDir, ".codefall")
	settingsPath = filepath.Join(codefallDir, "settings.json")
	schemaRemedy = "create .codefall/settings.json; schema: " + settings.SchemaID
	fixRemedy    = "fix the fields above; schema: " + settings.SchemaID
)

const validSettings = `{
  "$schema": "` + settings.SchemaID + `",
  "version": 1,
  "tracker": "github",
  "github": { "repo": "lividlabs/codefall-cli", "project": 3 }
}`

// --- fakes -------------------------------------------------------------------------------------

type fakeFileSystem struct {
	dirs  map[string]bool
	files map[string][]byte
	errs  map[string]error
}

func (f *fakeFileSystem) DirExists(path string) (bool, error) {
	if err, ok := f.errs[path]; ok {
		return false, err
	}

	return f.dirs[path], nil
}

func (f *fakeFileSystem) ReadFile(path string) ([]byte, error) {
	if err, ok := f.errs[path]; ok {
		return nil, err
	}

	data, ok := f.files[path]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}

	return data, nil
}

type runCall struct {
	dir     string
	command string
}

type fakeCommandRunner struct {
	paths map[string]string
	runs  map[string]CommandResult
	errs  map[string]error
	calls []runCall
}

func (r *fakeCommandRunner) LookPath(name string) mo.Option[string] {
	if path, ok := r.paths[name]; ok {
		return mo.Some(path)
	}

	return mo.None[string]()
}

func (r *fakeCommandRunner) Run(
	_ context.Context, dir, name string, args ...string,
) (CommandResult, error) {
	command := strings.TrimSpace(name + " " + strings.Join(args, " "))
	r.calls = append(r.calls, runCall{dir: dir, command: command})

	if err, ok := r.errs[command]; ok {
		return CommandResult{}, err
	}

	return r.runs[command], nil
}

// healthy is the fixture where every check passes; each case mutates it.
func healthy() (*fakeFileSystem, *fakeCommandRunner) {
	files := &fakeFileSystem{
		dirs:  map[string]bool{codefallDir: true},
		files: map[string][]byte{settingsPath: []byte(validSettings)},
		errs:  map[string]error{},
	}

	runner := &fakeCommandRunner{
		paths: map[string]string{"bd": "/opt/homebrew/bin/bd", "gh": "/opt/homebrew/bin/gh"},
		runs: map[string]CommandResult{
			"bd version":                  {Stdout: "bd version 1.2.2 (Homebrew)\n"},
			"bd info":                     {Stdout: "beads: 12 issues open\n"},
			"gh --version":                {Stdout: "gh version 2.97.0 (2026-07-31)\nhttps://github.com/cli/cli\n"},
			"gh auth status --json hosts": {Stdout: ghHosts(`{"state":"success","active":true,"host":"github.com","login":"djensen47","tokenSource":"keyring","scopes":"gist, project, read:org, repo, workflow","gitProtocol":"ssh"}`)},
		},
		errs: map[string]error{},
	}

	return files, runner
}

func ghHosts(accounts ...string) string {
	return `{"hosts":{"github.com":[` + strings.Join(accounts, ",") + `]}}`
}

// --- the table ---------------------------------------------------------------------------------

type outcome struct {
	id     string
	status domain.Status
}

var allPass = []outcome{
	{domain.CodefallDir.ID, domain.StatusPass},
	{domain.SettingsFile.ID, domain.StatusPass},
	{domain.SettingsJSON.ID, domain.StatusPass},
	{domain.SettingsComplete.ID, domain.StatusPass},
	{domain.BeadsInstalled.ID, domain.StatusPass},
	{domain.BeadsInitialized.ID, domain.StatusPass},
	{domain.GHInstalled.ID, domain.StatusPass},
	{domain.GHAuthenticated.ID, domain.StatusPass},
	{domain.GHScopes.ID, domain.StatusPass},
}

// outcomes restates allPass with some statuses changed and the skipped checks removed — a check that
// never ran is absent from the report, not present with a status of its own.
func outcomes(changes map[string]domain.Status, absent ...string) []outcome {
	result := make([]outcome, 0, len(allPass))

	for _, o := range allPass {
		if slices.Contains(absent, o.id) {
			continue
		}

		if status, ok := changes[o.id]; ok {
			o.status = status
		}

		result = append(result, o)
	}

	return result
}

// The skip rules, named so each case reads as the rule it exercises.
var (
	afterCodefallDir  = []string{domain.SettingsFile.ID, domain.SettingsJSON.ID, domain.SettingsComplete.ID}
	afterSettingsFile = []string{domain.SettingsJSON.ID, domain.SettingsComplete.ID}
	afterSettingsJSON = []string{domain.SettingsComplete.ID}
	afterBeadsMissing = []string{domain.BeadsInitialized.ID}
	afterGHMissing    = []string{domain.GHAuthenticated.ID, domain.GHScopes.ID}
	afterGHAuth       = []string{domain.GHScopes.ID}
)

func TestDiagnoseRun(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mutate     func(*fakeFileSystem, *fakeCommandRunner)
		want       []outcome
		target     string
		wantDetail string
		wantRemedy mo.Option[string]
	}{
		{
			name:   "everything is in place",
			mutate: func(*fakeFileSystem, *fakeCommandRunner) {},
			want:   allPass,
			target: domain.GHScopes.ID,
		},
		{
			name: "the .codefall directory is missing",
			mutate: func(f *fakeFileSystem, _ *fakeCommandRunner) {
				delete(f.dirs, codefallDir)
			},
			want:       outcomes(map[string]domain.Status{domain.CodefallDir.ID: domain.StatusFail}, afterCodefallDir...),
			target:     domain.CodefallDir.ID,
			wantDetail: ".codefall/ not found",
			wantRemedy: mo.Some(schemaRemedy),
		},
		{
			name: "the .codefall directory cannot be stat'd",
			mutate: func(f *fakeFileSystem, _ *fakeCommandRunner) {
				f.errs[codefallDir] = errors.New("permission denied")
			},
			want:       outcomes(map[string]domain.Status{domain.CodefallDir.ID: domain.StatusFail}, afterCodefallDir...),
			target:     domain.CodefallDir.ID,
			wantDetail: "Cannot stat " + codefallDir + ": permission denied",
			wantRemedy: mo.None[string](),
		},
		{
			name: "settings.json is missing",
			mutate: func(f *fakeFileSystem, _ *fakeCommandRunner) {
				delete(f.files, settingsPath)
			},
			want:       outcomes(map[string]domain.Status{domain.SettingsFile.ID: domain.StatusFail}, afterSettingsFile...),
			target:     domain.SettingsFile.ID,
			wantDetail: ".codefall/settings.json not found",
			wantRemedy: mo.Some(schemaRemedy),
		},
		{
			name: "settings.json cannot be read",
			mutate: func(f *fakeFileSystem, _ *fakeCommandRunner) {
				f.errs[settingsPath] = errors.New("permission denied")
			},
			want:       outcomes(map[string]domain.Status{domain.SettingsFile.ID: domain.StatusFail}, afterSettingsFile...),
			target:     domain.SettingsFile.ID,
			wantDetail: "Cannot read .codefall/settings.json: permission denied",
			wantRemedy: mo.None[string](),
		},
		{
			name: "settings.json is not valid JSON",
			mutate: func(f *fakeFileSystem, _ *fakeCommandRunner) {
				f.files[settingsPath] = []byte("{\n  \"version\": 1,,\n}")
			},
			want:   outcomes(map[string]domain.Status{domain.SettingsJSON.ID: domain.StatusFail}, afterSettingsJSON...),
			target: domain.SettingsJSON.ID,
			wantDetail: "settings.json is not valid JSON at byte 18: " +
				"invalid character ',' looking for beginning of object key string",
			wantRemedy: mo.None[string](),
		},
		{
			name: "settings.json is valid JSON but not an object",
			mutate: func(f *fakeFileSystem, _ *fakeCommandRunner) {
				f.files[settingsPath] = []byte(`["nope"]`)
			},
			want:       outcomes(map[string]domain.Status{domain.SettingsComplete.ID: domain.StatusFail}),
			target:     domain.SettingsComplete.ID,
			wantDetail: "settings.json's top level must be a JSON object",
			wantRemedy: mo.Some(fixRemedy),
		},
		{
			name: "settings.json is incomplete",
			mutate: func(f *fakeFileSystem, _ *fakeCommandRunner) {
				f.files[settingsPath] = []byte(`{"version": 2}`)
			},
			want:       outcomes(map[string]domain.Status{domain.SettingsComplete.ID: domain.StatusFail}),
			target:     domain.SettingsComplete.ID,
			wantDetail: "settings.json is incomplete: version: must be 1; tracker: missing",
			wantRemedy: mo.Some(fixRemedy),
		},
		{
			name: "bd is not installed",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				delete(r.paths, "bd")
			},
			want:       outcomes(map[string]domain.Status{domain.BeadsInstalled.ID: domain.StatusFail}, afterBeadsMissing...),
			target:     domain.BeadsInstalled.ID,
			wantDetail: "bd is not on PATH",
			wantRemedy: mo.Some("brew install beads"),
		},
		{
			name: "bd version fails, so the path is the detail",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				r.runs["bd version"] = CommandResult{ExitCode: 2, Stderr: "unknown command\n"}
			},
			want:       allPass,
			target:     domain.BeadsInstalled.ID,
			wantDetail: "/opt/homebrew/bin/bd",
			wantRemedy: mo.None[string](),
		},
		{
			name: "beads is not initialized here",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				r.runs["bd info"] = CommandResult{
					ExitCode: 1,
					Stderr:   "Error: no beads database found\nrun bd init\n",
				}
			},
			want:       outcomes(map[string]domain.Status{domain.BeadsInitialized.ID: domain.StatusWarn}),
			target:     domain.BeadsInitialized.ID,
			wantDetail: "This repository has no Beads database (bd info exited 1: Error: no beads database found)",
			wantRemedy: mo.Some("bd init"),
		},
		{
			name: "bd info cannot be started",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				r.errs["bd info"] = errors.New("run bd: fork/exec: resource temporarily unavailable")
			},
			want:       outcomes(map[string]domain.Status{domain.BeadsInitialized.ID: domain.StatusWarn}),
			target:     domain.BeadsInitialized.ID,
			wantDetail: "Could not run bd info: run bd: fork/exec: resource temporarily unavailable",
			wantRemedy: mo.Some("bd init"),
		},
		{
			name: "gh is not installed",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				delete(r.paths, "gh")
			},
			want:       outcomes(map[string]domain.Status{domain.GHInstalled.ID: domain.StatusFail}, afterGHMissing...),
			target:     domain.GHInstalled.ID,
			wantDetail: "gh is not on PATH",
			wantRemedy: mo.Some("brew install gh"),
		},
		{
			name: "gh auth status cannot be started",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				r.errs["gh auth status --json hosts"] = errors.New("run gh: signal: killed")
			},
			want:       outcomes(map[string]domain.Status{domain.GHAuthenticated.ID: domain.StatusFail}, afterGHAuth...),
			target:     domain.GHAuthenticated.ID,
			wantDetail: "Could not run gh auth status: run gh: signal: killed",
			wantRemedy: mo.None[string](),
		},
		{
			name: "gh auth status prints something that is not JSON",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				r.runs["gh auth status --json hosts"] = CommandResult{
					ExitCode: 1,
					Stderr:   "unknown flag: --json\n",
				}
			},
			want:       outcomes(map[string]domain.Status{domain.GHAuthenticated.ID: domain.StatusFail}, afterGHAuth...),
			target:     domain.GHAuthenticated.ID,
			wantDetail: "Unexpected output from gh auth status: unknown flag: --json",
			wantRemedy: mo.Some("gh auth login"),
		},
		{
			name: "no github.com account is active",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				r.runs["gh auth status --json hosts"] = CommandResult{
					Stdout: ghHosts(`{"state":"success","active":false,"login":"djensen47","scopes":"repo"}`),
				}
			},
			want:       outcomes(map[string]domain.Status{domain.GHAuthenticated.ID: domain.StatusFail}, afterGHAuth...),
			target:     domain.GHAuthenticated.ID,
			wantDetail: "No active github.com account",
			wantRemedy: mo.Some("gh auth login"),
		},
		{
			name: "the active account is in an error state",
			mutate: func(_ *fakeFileSystem, r *fakeCommandRunner) {
				r.runs["gh auth status --json hosts"] = CommandResult{
					Stdout: ghHosts(`{"state":"error","active":true,"login":"djensen47","scopes":"repo","error":"token expired"}`),
				}
			},
			want:       outcomes(map[string]domain.Status{domain.GHAuthenticated.ID: domain.StatusFail}, afterGHAuth...),
			target:     domain.GHAuthenticated.ID,
			wantDetail: `djensen47: state "error" token expired`,
			wantRemedy: mo.Some("gh auth login"),
		},
		{
			name:       "the token is missing repo",
			mutate:     withScopes("gist, project, read:org"),
			want:       outcomes(map[string]domain.Status{domain.GHScopes.ID: domain.StatusFail}),
			target:     domain.GHScopes.ID,
			wantDetail: "The token is missing repo",
			wantRemedy: mo.Some("gh auth refresh -s repo"),
		},
		{
			name:   "project satisfies read:project",
			mutate: withScopes("repo, project"),
			want:   allPass,
			target: domain.GHScopes.ID,
		},
		{
			name:       "the token has neither project scope",
			mutate:     withScopes("repo, workflow"),
			want:       outcomes(map[string]domain.Status{domain.GHScopes.ID: domain.StatusWarn}),
			target:     domain.GHScopes.ID,
			wantDetail: "The token is missing read:project, project",
			wantRemedy: mo.Some("gh auth refresh -s read:project -s project"),
		},
		{
			name:       "the token has none of the three scopes",
			mutate:     withScopes("gist"),
			want:       outcomes(map[string]domain.Status{domain.GHScopes.ID: domain.StatusFail}),
			target:     domain.GHScopes.ID,
			wantDetail: "The token is missing repo, read:project, project",
			wantRemedy: mo.Some("gh auth refresh -s repo -s read:project -s project"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, runner := healthy()
			tc.mutate(files, runner)

			report, err := NewDiagnose(files, runner).Run(t.Context(), workingDir)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}

			assertOutcomes(t, report, tc.want)

			if tc.wantDetail != "" || tc.wantRemedy.IsPresent() {
				assertResult(t, report, tc.target, tc.wantDetail, tc.wantRemedy)
			}
		})
	}
}

// withScopes rewrites the fixture's gh auth status output with a different scope string.
func withScopes(scopes string) func(*fakeFileSystem, *fakeCommandRunner) {
	return func(_ *fakeFileSystem, r *fakeCommandRunner) {
		r.runs["gh auth status --json hosts"] = CommandResult{
			Stdout: ghHosts(fmt.Sprintf(
				`{"state":"success","active":true,"login":"djensen47","scopes":%q}`, scopes)),
		}
	}
}

func assertOutcomes(t *testing.T, report domain.Report, want []outcome) {
	t.Helper()

	got := make([]outcome, 0, len(want))
	for _, result := range report.Results() {
		got = append(got, outcome{result.Check.ID, result.Status})
	}

	if !slices.Equal(got, want) {
		t.Errorf("outcomes =\n  %v\nwant\n  %v", got, want)
	}
}

func assertResult(t *testing.T, report domain.Report, id, wantDetail string, wantRemedy mo.Option[string]) {
	t.Helper()

	for _, result := range report.Results() {
		if result.Check.ID != id {
			continue
		}

		if got := result.Detail.OrElse(""); got != wantDetail {
			t.Errorf("%s detail = %q, want %q", id, got, wantDetail)
		}

		if got := result.Remedy; got != wantRemedy {
			t.Errorf("%s remedy = %v, want %v", id, got, wantRemedy)
		}

		return
	}

	t.Errorf("check %q is not in the report", id)
}

// --- the calls the use case makes ----------------------------------------------------------------

func TestDiagnoseRunsToolsInTheWorkingDirectory(t *testing.T) {
	files, runner := healthy()

	if _, err := NewDiagnose(files, runner).Run(t.Context(), workingDir); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []runCall{
		{workingDir, "bd version"},
		{workingDir, "bd info"},
		{workingDir, "gh --version"},
		{workingDir, "gh auth status --json hosts"},
	}

	if !slices.Equal(runner.calls, want) {
		t.Errorf("calls =\n  %v\nwant\n  %v", runner.calls, want)
	}
}

func TestDiagnoseReturnsAnErrorForACancelledContext(t *testing.T) {
	files, runner := healthy()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := NewDiagnose(files, runner).Run(ctx, workingDir)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want context.Canceled", err)
	}
}
