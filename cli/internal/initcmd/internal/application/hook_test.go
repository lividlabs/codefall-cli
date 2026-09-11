package application

import (
	"encoding/json"
	"errors"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// The definitions the embedded tree serves, one per harness. They mirror extensions/hooks/<harness>/,
// and the merge reads what comes back through the same doorway the real FS uses.
var hookDefinitions = map[string][]byte{
	"hooks/claude/hooks.json": []byte(`{
  "hooks": {
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "guard"}]}],
    "SessionStart": [{"matcher": "", "hooks": [{"type": "command", "command": "bd prime --hook-json"}]}]
  }
}`),
	"hooks/codex/hooks.json": []byte(`{
  "hooks": {
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "guard"}]}],
    "SessionStart": [{"matcher": "", "hooks": [{"type": "command", "command": "bd prime --hook-json"}]}]
  }
}`),
	"hooks/antigravity/hooks.json": []byte(`{
  "codefall-merge-guard": {
    "PreToolUse": [{"matcher": "run_command", "hooks": [{"command": "guard --antigravity"}]}]
  }
}`),
	"hooks/opencode/codefall.js": []byte("// the adapter\n"),
}

// hookResult is what the fourth step did in a run that got that far.
func hookResult(t *testing.T, report domain.Report) domain.StepResult {
	t.Helper()

	results := report.Results()
	if len(results) != 5 {
		t.Fatalf("Results() = %+v, want a result for each of the five steps", results)
	}

	return results[3]
}

// The three JSON harnesses merge their definitions; OpenCode copies its plugin. A harness in the
// table without a definition would be a bug the table would not have room to hide.
func TestHookRegistersWhatTheTableSays(t *testing.T) {
	for _, tc := range []struct {
		harness string
		dest    string
	}{
		{domain.HarnessClaudeCode, claudeFull},
		{domain.HarnessCodex, filepath.Join(workingDir, ".codex/hooks.json")},
		{domain.HarnessAntigravity, filepath.Join(workingDir, ".agents/hooks.json")},
		{domain.HarnessOpenCode, filepath.Join(workingDir, ".opencode/plugins/codefall.js")},
	} {
		t.Run(tc.harness, func(t *testing.T) {
			files := settled("")

			report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), beadsRequestHarness(tc.harness), nil)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}

			result := hookResult(t, report)
			if result.Outcome != domain.OutcomeDone {
				t.Errorf("outcome = %v, want DONE", result.Outcome)
			}

			if _, ok := files.files[tc.dest]; !ok {
				t.Errorf("nothing landed at %q; files = %v", tc.dest, keysOf(files))
			}
		})
	}
}

// beadsRequestHarness is beadsRequest with the harness the test wants.
func beadsRequestHarness(harness string) Request {
	request := beadsRequest()
	request.Harness = harness
	return request
}

// keysOf maps the fake's written files, for failure messages.
func keysOf(files *fakeFileSystem) []string {
	keys := make([]string, 0, len(files.files))
	for path := range files.files {
		keys = append(keys, path)
	}
	slices.Sort(keys)
	return keys
}

// assertKeysSurvive holds the file to what it said before the run: every key it had is still there,
// with the same value. It is decoded rather than compared as bytes, because encoding/json sorts an
// object's keys and the file comes back in its order, not the one it was written in.
func assertKeysSurvive(t *testing.T, got []byte, before string) {
	t.Helper()

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

// The merge joins what the project already has rather than replacing it: another SessionStart hook
// stays, and so does every key the file keeps for its own reasons.
func TestHookMergeKeepsWhatTheFileAlreadySays(t *testing.T) {
	const before = `{
  "enabledPlugins": {"codefall@codefall": true},
  "permissions": {"allow": ["Bash(git status:*)"]},
  "hooks": {
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "other"}]}],
    "SessionStart": [{"matcher": "", "hooks": [{"type": "command", "command": "echo hello"}]}]
  }
}`

	files := settled(before)

	report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), beadsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := hookResult(t, report).Outcome; got != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", got)
	}

	got := files.files[claudeFull]

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

	if got := document.Hooks["PreToolUse"]; len(got) != 2 || got[1].Hooks[0].Command != "guard" {
		t.Errorf("PreToolUse = %+v, want the project's entry kept and codefall's merged", got)
	}

	sessions := document.Hooks["SessionStart"]
	if len(sessions) != 2 || sessions[0].Hooks[0].Command != "echo hello" ||
		sessions[1].Hooks[0].Command != "bd prime --hook-json" {
		t.Errorf("SessionStart = %+v, want the project's entry first, codefall's second", sessions)
	}
}

// A destination that already has every entry in the definition — dedupe is per entry, not all or
// nothing — is left alone, whatever else the file says.
func TestHookMergeSkipsWhatIsAlreadyThere(t *testing.T) {
	const before = `{"hooks": {
  "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "guard"}]}],
  "SessionStart": [{"matcher": "", "hooks": [{"type": "command", "command": "bd prime --hook-json"}]}]
}}`

	files := settled(before)

	report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), beadsRequest(), nil)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := hookResult(t, report)
	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	want := "codefall's hooks are already in .claude/settings.json"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}

	if got := string(files.files[claudeFull]); got != before {
		t.Errorf(".claude/settings.json = %q, want it untouched", got)
	}
}

// The same command under another matcher is not the same entry: the harness would run it for other
// tool calls, and this event would stay unguarded. It joins rather than counting as present.
func TestHookMergeAddsTheSameCommandUnderADifferentMatcher(t *testing.T) {
	files := settled(`{"hooks": {"SessionStart": [` +
		`{"matcher": "startup", "hooks": [{"type": "command", "command": "bd prime --hook-json"}]}]}}`)

	report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), beadsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := hookResult(t, report).Outcome; got != domain.OutcomeDone {
		t.Fatalf("outcome = %v, want DONE", got)
	}

	var document struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}

	if err := json.Unmarshal(files.files[claudeFull], &document); err != nil {
		t.Fatalf("decode what was written: %v", err)
	}

	sessions := document.Hooks["SessionStart"]
	if len(sessions) != 2 || sessions[0].Matcher != "startup" || sessions[1].Matcher != "" {
		t.Errorf("SessionStart = %+v, want the startup entry kept and the every-session one added", sessions)
	}
	for _, session := range sessions {
		if session.Hooks[0].Command != "bd prime --hook-json" {
			t.Errorf("SessionStart entry = %+v, want both running the prime command", session)
		}
	}
}

// A section the project nulled out says nothing: it merges like an absent key rather than stopping
// the run the way a mistyped one does.
func TestHookMergeTreatsANulledSectionAsAbsent(t *testing.T) {
	files := settled(`{"hooks": null}`)

	report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), beadsRequest(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := hookResult(t, report).Outcome; got != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", got)
	}

	got := string(files.files[claudeFull])
	if !strings.Contains(got, "bd prime --hook-json") || !strings.Contains(got, `"PreToolUse"`) {
		t.Errorf(".claude/settings.json = %s, want both events merged", got)
	}
}

// Antigravity keys its hooks by name at the top level, and the merge grows the same arrays under
// them. A flag the project set on codefall's own hook — disabled it, say — is a scalar, and scalars
// the destination says are kept.
func TestHookMergeKeepsTheProjectsAntigravityFlags(t *testing.T) {
	files := settled("")
	files.files[filepath.Join(workingDir, ".agents/hooks.json")] = []byte(
		`{"codefall-merge-guard": {"enabled": false}}`)

	request := beadsRequestHarness(domain.HarnessAntigravity)
	report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), request, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := hookResult(t, report).Outcome; got != domain.OutcomeDone {
		t.Errorf("outcome = %v, want DONE", got)
	}

	got := string(files.files[filepath.Join(workingDir, ".agents/hooks.json")])
	if !strings.Contains(got, `"enabled": false`) {
		t.Errorf(".agents/hooks.json = %s, want the project's enabled flag kept", got)
	}
	if !strings.Contains(got, `"PreToolUse"`) {
		t.Errorf(".agents/hooks.json = %s, want the guard merged", got)
	}
}

// The OpenCode plugin is copied, and a rerun with nothing new is a skip, not a rewrite.
func TestHookCopySkipsAnInstallationItAlreadyHas(t *testing.T) {
	files := settled("")
	request := beadsRequestHarness(domain.HarnessOpenCode)

	if _, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), request, nil); err != nil {
		t.Fatalf("first run: %v", err)
	}

	report, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), request, nil)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	result := hookResult(t, report)
	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	want := "codefall's plugin at .opencode/plugins/codefall.js is already installed"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}
}

// An installed plugin init cannot read stops the step with the read error, not a copy over it.
func TestHookCopyReportsAFileItCannotRead(t *testing.T) {
	files := settled("")
	path := filepath.Join(workingDir, ".opencode/plugins/codefall.js")
	files.errs[path] = errors.New("permission denied")

	_, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).hook(t.Context(), beadsRequestHarness(domain.HarnessOpenCode))
	if err == nil || !strings.Contains(err.Error(), "read .opencode/plugins/codefall.js") {
		t.Errorf("hook error = %v, want it to say the file could not be read", err)
	}
}

// A file the step cannot make sense of stops the run rather than being written over: it is somebody
// else's file, and init is not the thing that should decide what it meant.
//
// The step is called directly rather than through a run, because the extension step reads the same
// file first and refuses the same two bodies with the same words — so a run would prove nothing
// about the branches here.
func TestHookMergeReportsAFileItCannotWorkWith(t *testing.T) {
	for _, tc := range []struct {
		name     string
		harness  string
		have     map[string][]byte
		settings string
		want     string
	}{
		{
			name:     "the file is not JSON",
			harness:  domain.HarnessClaudeCode,
			settings: "{ not json",
			want:     "decode .claude/settings.json",
		},
		{
			name:     "the file is a JSON array",
			harness:  domain.HarnessClaudeCode,
			settings: `[{"hooks": {}}]`,
			want:     "decode .claude/settings.json",
		},
		{
			name:     "the file is a JSON string",
			harness:  domain.HarnessClaudeCode,
			settings: `"hooks"`,
			want:     "decode .claude/settings.json",
		},
		{
			name:     "hooks is not an object",
			harness:  domain.HarnessClaudeCode,
			settings: `{"hooks": ["SessionStart"]}`,
			want:     ".claude/settings.json: hooks is not an object",
		},
		{
			name:     "the event is not an array",
			harness:  domain.HarnessClaudeCode,
			settings: `{"hooks": {"SessionStart": {"matcher": ""}}}`,
			want:     ".claude/settings.json: hooks.SessionStart is not an array",
		},
		{
			name:    "an event is not an array under antigravity's hook name",
			harness: domain.HarnessAntigravity,
			have:    map[string][]byte{filepath.Join(workingDir, ".agents/hooks.json"): []byte(`{"codefall-merge-guard": {"PreToolUse": {"matcher": ""}}}`)},
			want:    ".agents/hooks.json: codefall-merge-guard.PreToolUse is not an array",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := settled(tc.settings)
			for path, body := range tc.have {
				files.files[path] = body
			}

			_, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).hook(t.Context(), beadsRequestHarness(tc.harness))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("hook error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

// A file that cannot be read at all is the step's error too, named so the reader knows which file.
func TestHookMergeReportsAFileItCannotRead(t *testing.T) {
	files := settled("{}")
	files.errs[claudeFull] = errors.New("permission denied")

	_, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).hook(t.Context(), beadsRequest())
	if err == nil || !strings.Contains(err.Error(), "read .claude/settings.json") {
		t.Errorf("hook error = %v, want it to name .claude/settings.json the way the report does", err)
	}
}

// The file is somebody else's, so what it says comes back byte for byte. json.Marshal would escape
// <, > and & into \u sequences and quietly rewrite a permission rule; the step's encoder is told
// not to.
func TestHookMergeDoesNotEscapeWhatTheFileAlreadySays(t *testing.T) {
	const rule = "Bash(test a && b < c > d)"

	files := settled(`{"permissions": {"allow": ["` + rule + `"]}}`)

	if _, err := NewInitialize(files, toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), beadsRequest(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := string(files.files[claudeFull])
	if !strings.Contains(got, rule) {
		t.Errorf(".claude/settings.json =\n%s\nwant %q in it, unescaped", got, rule)
	}

	if strings.Contains(got, `\u00`) {
		t.Errorf(".claude/settings.json =\n%s\nwant no \\u escapes in it", got)
	}
}

// A harness that takes codefall only as skills has no hook definition, and the step says so rather
// than writing somewhere it should not.
func TestHookSkipsAHarnessWithNoDefinition(t *testing.T) {
	request := beadsRequestHarness(domain.HarnessMuse)

	report, err := NewInitialize(settled(""), toolsInstalled(), newFakeExtensionSource()).Run(t.Context(), request, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result := hookResult(t, report)
	if result.Outcome != domain.OutcomeSkipped {
		t.Errorf("outcome = %v, want SKIPPED", result.Outcome)
	}

	want := "codefall has no hooks for muse"
	if result.Detail != want {
		t.Errorf("detail = %q, want %q", result.Detail, want)
	}
}

// The mechanism table guards the step the same way the extension step is guarded.
func TestHookRefusesAHarnessItDoesNotKnow(t *testing.T) {
	request := beadsRequestHarness("aider")

	// The extension step refuses this harness first, so the hook step is asked on its own.
	_, err := NewInitialize(settled(""), toolsInstalled(), newFakeExtensionSource()).hook(t.Context(), request)
	if err == nil || !strings.Contains(err.Error(), `harness "aider" has no extension mechanism`) {
		t.Errorf("hook error = %v, want it to say the harness has no extension mechanism", err)
	}
}
