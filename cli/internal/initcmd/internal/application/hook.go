package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// The keys Claude Code's settings file uses to describe a hook. They are the shape of somebody
// else's file, which is why they are here and not in the domain: what codefall installs is the
// command and the event, and those are domain constants.
const (
	hooksKey        = "hooks"
	matcherKey      = "matcher"
	typeKey         = "type"
	commandKey      = "command"
	commandHookType = "command"
)

// hook is the fourth step of a run, and only Claude Code has a hook to write: it tells the harness
// to prime each session with what Beads knows.
//
// bd would install this hook itself, along with a section in AGENTS.md and CLAUDE.md that codefall
// owns and bd cannot be told to leave alone — so init writes the hook and bd writes nothing outside
// .beads/. A harness that takes codefall's skills as files has no session hook for them yet, so its
// report is a skip rather than an error that would leave the project half set up.
func (i *Initialize) hook(_ context.Context, request Request) (domain.StepResult, error) {
	if _, known := extensionDestDirs[request.Harness]; !known {
		return domain.StepResult{}, fmt.Errorf("harness %q has no extension mechanism", request.Harness)
	}

	if request.Harness != domain.HarnessClaudeCode {
		return domain.HookStep.Skipped(fmt.Sprintf(
			"codefall has no session hook for %s yet", request.Harness)), nil
	}

	return i.claudeCodeHook(request.Dir)
}

// claudeCodeHook appends codefall's SessionStart hook to .claude/settings.json, which the extension
// step's CLI has already created by the time this runs.
//
// The file is read as a plain object rather than into a struct, so that every key it has — the
// extension declarations next to this one, and whatever else the project keeps there — survives being
// written back out. Key order is the one thing that does not: encoding/json sorts an object's keys,
// so a hand-edited file comes back sorted. That is a diff once, and the alternative is a JSON
// editor.
func (i *Initialize) claudeCodeHook(dir string) (domain.StepResult, error) {
	document, err := i.readClaudeDocument(dir)
	if err != nil {
		return domain.StepResult{}, err
	}

	hooks, err := objectAt(document, hooksKey)
	if err != nil {
		return domain.StepResult{}, err
	}

	entries, err := arrayAt(hooks, hooksKey, domain.BeadsHookEvent)
	if err != nil {
		return domain.StepResult{}, err
	}

	if hasBeadsHook(entries) {
		return domain.HookStep.Skipped(fmt.Sprintf("the bd prime %s hook is already in %s",
			domain.BeadsHookEvent, claudeFullName)), nil
	}

	// An empty matcher is how Claude Code says "every session", which is the only kind of session
	// that could want a stale view of the issue database.
	hooks[domain.BeadsHookEvent] = append(entries, map[string]any{
		matcherKey: "",
		hooksKey: []any{map[string]any{
			typeKey:    commandHookType,
			commandKey: domain.BeadsHookCommand,
		}},
	})
	document[hooksKey] = hooks

	data, err := encodeClaudeDocument(document)
	if err != nil {
		return domain.StepResult{}, err
	}

	if err := i.files.MkdirAll(filepath.Join(dir, claudeDir)); err != nil {
		return domain.StepResult{}, fmt.Errorf("create %s/: %w", claudeDir, err)
	}

	if err := i.files.WriteFile(claudePath(dir), data); err != nil {
		return domain.StepResult{}, fmt.Errorf("write %s: %w", claudeFullName, err)
	}

	return domain.HookStep.Done(fmt.Sprintf("added the bd prime %s hook to %s",
		domain.BeadsHookEvent, claudeFullName)), nil
}

// encodeClaudeDocument renders the harness's settings back into the bytes that go in the file.
//
// The encoder is built by hand rather than through json.Marshal because Marshal escapes <, > and &
// into \u sequences, which would silently rewrite a permission rule like "Bash(a && b)" in a file
// codefall does not own. Encode writes the trailing newline itself.
func encodeClaudeDocument(document map[string]any) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(document); err != nil {
		return nil, fmt.Errorf("encode %s: %w", claudeFullName, err)
	}

	return buffer.Bytes(), nil
}

// readClaudeDocument reads the harness's settings as the object they are. A file with nothing to say
// is an empty object — this step creates what it needs. A file that is not an object at all is an
// error, because the step is about to write over it.
func (i *Initialize) readClaudeDocument(dir string) (map[string]any, error) {
	body, err := i.readClaudeFile(dir)
	if err != nil {
		return nil, err
	}

	document := map[string]any{}

	data, said := body.Get()
	if !said {
		return document, nil
	}

	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode %s: %w", claudeFullName, err)
	}

	return document, nil
}

// objectAt returns the object stored under key, or an empty one when the file says nothing there.
// A key holding something that is not an object is an error rather than something to overwrite: the
// file is not shaped the way the harness expects, and init is not the thing that should decide what
// it meant.
func objectAt(document map[string]any, key string) (map[string]any, error) {
	value, present := document[key]
	if !present || value == nil {
		return map[string]any{}, nil
	}

	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: %s is not an object", claudeFullName, key)
	}

	return object, nil
}

// arrayAt is objectAt for a list, and refuses a wrong-typed key for the same reason. The parent's
// name is passed in so the complaint names the path a reader would go and look at, rather than
// whichever key this function happens to know the name of.
func arrayAt(object map[string]any, parent, key string) ([]any, error) {
	value, present := object[key]
	if !present || value == nil {
		return nil, nil
	}

	array, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s: %s.%s is not an array", claudeFullName, parent, key)
	}

	return array, nil
}

// hasBeadsHook reports whether one of the event's entries already runs the command codefall would
// add. Anything shaped unexpectedly further down is skipped rather than refused: this is a search
// for one command, and a hook init did not write is not init's to judge.
func hasBeadsHook(entries []any) bool {
	for _, entry := range entries {
		object, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		commands, ok := object[hooksKey].([]any)
		if !ok {
			continue
		}

		for _, command := range commands {
			hook, ok := command.(map[string]any)
			if !ok {
				continue
			}

			if name, ok := hook[commandKey].(string); ok && name == domain.BeadsHookCommand {
				return true
			}
		}
	}

	return false
}
