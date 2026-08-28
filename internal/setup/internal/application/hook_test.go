package application

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// beadsHook is the file the step writes into a project that had nothing to say about hooks. Key
// order is encoding/json's, which sorts them.
const beadsHook = `{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "command": "bd prime --hook-json",
            "type": "command"
          }
        ],
        "matcher": ""
      }
    ]
  }
}
`

// hookResult is what the fourth step did in a run that got that far.
func hookResult(t *testing.T, report domain.Report) domain.StepResult {
	t.Helper()

	results := report.Results()
	if len(results) != 4 {
		t.Fatalf("Results() = %+v, want a result for each of the four steps", results)
	}

	return results[3]
}

// assertKeysSurvive holds the file to what it said before the run: every key it had is still there,
// with the same value. It is decoded rather than compared as bytes, because encoding/json sorts an
// object's keys and the file comes back in its order, not the one it was written in.
func assertKeysSurvive(t *testing.T, got []byte, before string) {
	t.Helper()

	if before == "" {
		return
	}

	var was, now map[string]any

	if err := json.Unmarshal([]byte(before), &was); err != nil {
		t.Fatalf("the settings the test started from are not an object: %v", err)
	}

	if err := json.Unmarshal(got, &now); err != nil {
		t.Fatalf("settings after the run are not an object: %v", err)
	}

	for _, key := range slices.Sorted(maps.Keys(was)) {
		if !reflect.DeepEqual(now[key], was[key]) {
			t.Errorf("%s = %#v, want %#v — the run lost what the file said", key, now[key], was[key])
		}
	}
}

// A project with no hooks of its own gets the file the harness expects, and the directory to put it
// in when the harness CLI has not made one.
func TestHookStepAddsTheHookToAFileThatHasNone(t *testing.T) {
	files := settled("")

	report, err := NewInitialize(files, toolsInstalled()).Run(t.Context(), beadsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := hookResult(t, report)
	if result.Outcome != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", result.Outcome)
	}

	if want := "added the bd prime SessionStart hook to .claude/settings.json"; result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if got := string(files.files[claudeFull]); got != beadsHook {
		t.Errorf(".claude/settings.json =\n%s\nwant\n%s", got, beadsHook)
	}

	if !slices.Contains(files.made, claudeSettingsDir) {
		t.Errorf("created %q, want %q among them", files.made, claudeSettingsDir)
	}
}

// The hook joins what the project already has rather than replacing it: another SessionStart hook
// stays, and so does every key the file keeps for its own reasons.
func TestHookStepKeepsWhatTheFileAlreadySays(t *testing.T) {
	const before = `{
  "enabledPlugins": {"codefall@codefall": true},
  "permissions": {"allow": ["Bash(git status:*)"]},
  "hooks": {
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "guard"}]}],
    "SessionStart": [{"matcher": "", "hooks": [{"type": "command", "command": "echo hello"}]}]
  }
}`

	files := settled(before)

	report, err := NewInitialize(files, toolsInstalled()).Run(t.Context(), beadsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := hookResult(t, report).Outcome; got != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", got)
	}

	got := files.files[claudeFull]

	// Everything outside hooks is untouched.
	assertKeysSurvive(t, got, `{
  "enabledPlugins": {"codefall@codefall": true},
  "permissions": {"allow": ["Bash(git status:*)"]}
}`)

	var document struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}

	if err := json.Unmarshal(got, &document); err != nil {
		t.Fatalf("decode what was written: %v", err)
	}

	if len(document.Hooks["PreToolUse"]) != 1 {
		t.Errorf("PreToolUse = %+v, want the entry the project already had", document.Hooks["PreToolUse"])
	}

	sessions := document.Hooks[domain.BeadsHookEvent]
	if len(sessions) != 2 {
		t.Fatalf("SessionStart = %+v, want the project's entry and codefall's", sessions)
	}

	if sessions[0].Hooks[0].Command != "echo hello" {
		t.Errorf("the first entry runs %q, want the project's own hook first", sessions[0].Hooks[0].Command)
	}

	if sessions[1].Matcher != "" || sessions[1].Hooks[0].Type != "command" ||
		sessions[1].Hooks[0].Command != domain.BeadsHookCommand {
		t.Errorf("the appended entry = %+v, want codefall's hook", sessions[1])
	}
}

// A hook that is already there is left where it is, whatever else the entry it sits in says.
func TestHookStepSkipsAHookThatIsAlreadyThere(t *testing.T) {
	const before = `{"hooks": {"SessionStart": [` +
		`{"matcher": "startup", "hooks": [{"type": "command", "command": "bd prime --hook-json"}]}]}}`

	files := settled(before)

	report, err := NewInitialize(files, toolsInstalled()).Run(t.Context(), beadsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := hookResult(t, report)
	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	want := "the bd prime SessionStart hook is already in .claude/settings.json"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if got := string(files.files[claudeFull]); got != before {
		t.Errorf(".claude/settings.json = %q, want it untouched", got)
	}
}

// A file the step cannot make sense of stops the run rather than being written over: it is somebody
// else's file, and init is not the thing that should decide what it meant.
func TestHookStepReportsAFileItCannotWorkWith(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings string
		want     string
	}{
		{
			name:     "the file is not JSON",
			settings: "{ not json",
			want:     "decode .claude/settings.json",
		},
		{
			name:     "the file is JSON but not an object",
			settings: `[{"hooks": {}}]`,
			want:     "decode .claude/settings.json",
		},
		{
			name:     "hooks is not an object",
			settings: `{"hooks": ["SessionStart"]}`,
			want:     ".claude/settings.json: hooks is not an object",
		},
		{
			name:     "the event is not an array",
			settings: `{"hooks": {"SessionStart": {"matcher": ""}}}`,
			want:     ".claude/settings.json: hooks.SessionStart is not an array",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := settled(tc.settings)

			_, err := NewInitialize(files, toolsInstalled()).Run(t.Context(), beadsRequest(), nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Run error = %v, want it to mention %q", err, tc.want)
			}

			if got := string(files.files[claudeFull]); got != tc.settings {
				t.Errorf(".claude/settings.json = %q, want it untouched", got)
			}
		})
	}
}

// The switch is where the next harness lands, the way the plugin step's is.
func TestHookStepRefusesAHarnessItDoesNotKnow(t *testing.T) {
	request := beadsRequest()
	request.Harness = "aider"

	// The plugin step refuses this harness first, so the hook step is asked on its own.
	result, err := NewInitialize(settled(""), toolsInstalled()).hook(t.Context(), request)
	if err == nil || !strings.Contains(err.Error(), `harness "aider" has no session hook to add`) {
		t.Errorf("hook = %+v, %v, want it to say the harness has no hook", result, err)
	}
}
