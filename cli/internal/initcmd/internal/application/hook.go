package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
)

// commandKey is the one key every harness's hook format shares: the shell command to run. The merge
// dedupes on it, so a rerun does not repeat a registration.
const commandKey = "command"

// How a hook definition lands. The JSON harnesses get one recursive merge each — what differs
// between them is the destination, not the algorithm — and OpenCode's plugin is a copied file.
const (
	formatMerge hookFormat = iota
	formatCopy
)

type hookFormat int

// hookSpec is one harness's registration: which file in the embedded tree, where it goes in the
// project, and how to treat it.
type hookSpec struct {
	source string
	dest   string
	format hookFormat
}

// hookSpecs is the one table of what init registers where. A harness with no entry is skipped
// rather than errored: the run reports what it did not need to do. Adding a harness that reads one
// of these formats is one row; adding one that needs a new treatment is a third value of hookFormat.
var hookSpecs = map[string]hookSpec{
	domain.HarnessClaudeCode:  {source: "hooks/claude/hooks.json", dest: claudeFullName, format: formatMerge},
	domain.HarnessCodex:       {source: "hooks/codex/hooks.json", dest: ".codex/hooks.json", format: formatMerge},
	domain.HarnessAntigravity: {source: "hooks/antigravity/hooks.json", dest: ".agents/hooks.json", format: formatMerge},
	domain.HarnessOpenCode:    {source: "hooks/opencode/codefall.js", dest: ".opencode/plugins/codefall.js", format: formatCopy},
}

// hookSourceDirs is every directory of per-harness hook definitions, derived from hookSpecs so a
// new harness lands in exactly one place: the extension step copies the shared scripts and leaves
// these for the hook step, which reads them straight from the embedded tree.
var hookSourceDirs = func() []string {
	dirs := make([]string, 0, len(hookSpecs))
	for _, spec := range hookSpecs {
		dir := path.Dir(spec.source)
		if !slices.Contains(dirs, dir) {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}()

// hook is the fourth step of a run: it takes the harness's entry in hookSpecs and lands it. The
// extension step has already copied the shared scripts the hooks point at, so the destination of
// any script path in a definition exists by the time this runs.
func (i *Initialize) hook(_ context.Context, request Request) (domain.StepResult, error) {
	if _, known := extensionDestDirs[request.Harness]; !known {
		return domain.StepResult{}, fmt.Errorf("harness %q has no extension mechanism", request.Harness)
	}

	spec, known := hookSpecs[request.Harness]
	if !known {
		return domain.HookStep.Skipped(fmt.Sprintf("codefall has no hooks for %s", request.Harness)), nil
	}

	if spec.format == formatCopy {
		return i.copyHook(request.Dir, spec)
	}

	return i.mergeHook(request.Dir, spec)
}

// mergeHook merges the harness's hook definitions into the file it reads them from. The
// destination is decoded as a plain object so everything else it says — permissions, plugin
// declarations, other hooks — survives the round trip; encoding/json sorts keys on the way out,
// which is a one-time diff on hand-ordered files.
func (i *Initialize) mergeHook(dir string, spec hookSpec) (domain.StepResult, error) {
	source, err := i.source.Read(spec.source)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("read %s: %w", spec.source, err)
	}

	var addition map[string]any
	if err := json.Unmarshal(source, &addition); err != nil {
		return domain.StepResult{}, fmt.Errorf("decode %s: %w", spec.source, err)
	}

	body, err := i.readDestination(filepath.Join(dir, spec.dest), spec.dest)
	if err != nil {
		return domain.StepResult{}, err
	}

	document := map[string]any{}
	if data, said := body.Get(); said {
		if err := json.Unmarshal(data, &document); err != nil {
			return domain.StepResult{}, fmt.Errorf("decode %s: %w", spec.dest, err)
		}
	}

	changed, err := mergeInto(document, addition, "", spec.dest)
	if err != nil {
		return domain.StepResult{}, err
	}

	if !changed {
		return domain.HookStep.Skipped(fmt.Sprintf("codefall's hooks are already in %s", spec.dest)), nil
	}

	data, err := encodeDocument(document, spec.dest)
	if err != nil {
		return domain.StepResult{}, err
	}

	if err := i.files.MkdirAll(filepath.Join(dir, filepath.Dir(spec.dest))); err != nil {
		return domain.StepResult{}, fmt.Errorf("create %s/: %w", filepath.Dir(spec.dest), err)
	}

	if err := i.files.WriteFile(filepath.Join(dir, spec.dest), data); err != nil {
		return domain.StepResult{}, fmt.Errorf("write %s: %w", spec.dest, err)
	}

	return domain.HookStep.Done(fmt.Sprintf("merged codefall's hooks into %s", spec.dest)), nil
}

// copyHook installs the harness's plugin by copying one file over. A plugin whose bytes match is a
// skip: what would be written is what is there.
func (i *Initialize) copyHook(dir string, spec hookSpec) (domain.StepResult, error) {
	source, err := i.source.Read(spec.source)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("read %s: %w", spec.source, err)
	}

	path := filepath.Join(dir, spec.dest)

	current, err := i.files.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return domain.StepResult{}, fmt.Errorf("read %s: %w", spec.dest, err)
	}
	if err == nil && bytes.Equal(current, source) {
		return domain.HookStep.Skipped(fmt.Sprintf("codefall's plugin at %s is already installed", spec.dest)), nil
	}

	if err := i.files.MkdirAll(filepath.Join(dir, filepath.Dir(spec.dest))); err != nil {
		return domain.StepResult{}, fmt.Errorf("create %s/: %w", filepath.Dir(spec.dest), err)
	}

	if err := i.files.WriteFile(path, source); err != nil {
		return domain.StepResult{}, fmt.Errorf("write %s: %w", spec.dest, err)
	}

	return domain.HookStep.Done(fmt.Sprintf("installed codefall's plugin at %s", spec.dest)), nil
}

// mergeInto folds the addition into the document and reports whether anything came in. Objects
// merge recursively and arrays are appended; a scalar the document already says is kept, because
// only the lists the file holds are codefall's to grow — an Antigravity hook the project disabled
// stays disabled. A path where the two disagree about type is an error naming it: the file is not
// shaped the way the harness expects, and init is not the thing that should decide what it meant.
func mergeInto(document, addition map[string]any, path string, dest string) (bool, error) {
	changed := false

	for key, incoming := range addition {
		existing, declared := document[key]
		// A key holding null says nothing, the same as a key that is absent: a hand-edited file
		// that nulls out a section it does not use is not a file shaped against the harness.
		if !declared || existing == nil {
			document[key] = incoming
			changed = true
			continue
		}

		merged, grew, err := mergeValue(existing, incoming, join(path, key), dest)
		if err != nil {
			return false, err
		}
		if grew {
			document[key] = merged
			changed = true
		}
	}

	return changed, nil
}

// mergeValue resolves one key's worth of source against the same key in the destination. Scalars
// and source entries that turn up nothing new never change the destination.
func mergeValue(existing, incoming any, path, dest string) (any, bool, error) {
	switch src := incoming.(type) {
	case map[string]any:
		dst, ok := existing.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("%s: %s is not an object", dest, path)
		}
		grew, err := mergeInto(dst, src, path, dest)
		if err != nil {
			return nil, false, err
		}
		return dst, grew, nil

	case []any:
		dst, ok := existing.([]any)
		if !ok {
			return nil, false, fmt.Errorf("%s: %s is not an array", dest, path)
		}
		merged, grew := appendEntries(dst, src)
		return merged, grew, nil
	}

	// A scalar the destination already set is the destination's own business.
	return nil, false, nil
}

// appendEntries adds each source entry the destination array does not already run, and reports
// whether the array grew. An entry is known by its matcher together with the commands it runs: the
// same command registered under another matcher still leaves this event unguarded, so it joins
// rather than being taken for present. The command is wherever the harness's format keeps it —
// inside a nested handler list, as in Claude's and Codex's, or on the entry itself, as in
// Antigravity's — so the search is recursive. An entry that declares no command is joined by
// content instead: the same entry twice is one entry.
func appendEntries(dst, src []any) ([]any, bool) {
	known := map[string]bool{}

	for _, entry := range dst {
		for _, command := range commandsIn(entry) {
			known[matcherOf(entry)+"\x00"+command] = true
		}
	}

	grew := false

	for _, entry := range src {
		commands := commandsIn(entry)

		if len(commands) == 0 {
			duplicate := false
			for _, existing := range dst {
				if reflect.DeepEqual(existing, entry) {
					duplicate = true
					break
				}
			}
			if !duplicate {
				dst = append(dst, entry)
				grew = true
			}
			continue
		}

		matcher := matcherOf(entry)

		duplicate := true
		for _, command := range commands {
			if !known[matcher+"\x00"+command] {
				duplicate = false
				break
			}
		}
		if duplicate {
			continue
		}

		dst = append(dst, entry)
		grew = true
		for _, command := range commands {
			known[matcher+"\x00"+command] = true
		}
	}

	return dst, grew
}

// matcherOf is the matcher's half of an entry's identity: the filter the harness applies before the
// entry's command ever runs. An entry that declares none matches everything, which is the empty
// string — the same nothing a missing matcher means to the harness.
func matcherOf(entry any) string {
	object, ok := entry.(map[string]any)
	if !ok {
		return ""
	}

	matcher, _ := object["matcher"].(string)
	return matcher
}

// join names the path a reader would go and look at, rather than whichever key a helper happens to
// know the name of.
func join(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// commandsIn collects every command string an entry runs, whatever shape the entry has. An entry
// that holds none contributes nothing to the dedupe set.
func commandsIn(value any) []string {
	switch shaped := value.(type) {
	case map[string]any:
		var commands []string
		if command, ok := shaped[commandKey].(string); ok {
			commands = append(commands, command)
		}
		for _, nested := range shaped {
			commands = append(commands, commandsIn(nested)...)
		}
		return commands

	case []any:
		var commands []string
		for _, nested := range shaped {
			commands = append(commands, commandsIn(nested)...)
		}
		return commands
	}

	return nil
}

// readDestination is the one read of a file a merge writes back: three files say the same nothing —
// absent, empty or whitespace, and the JSON literal null — and they are one answer here rather than
// three behaviours further down, because json.Unmarshal accepts null into anything while rejecting
// the other two (ADR-GO-03). A read failure names the destination the way a person reads it in a
// report, the same display form every other error in the step uses.
func (i *Initialize) readDestination(path, display string) (mo.Option[[]byte], error) {
	data, err := i.files.ReadFile(path)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return mo.None[[]byte](), nil
	case err != nil:
		return mo.None[[]byte](), fmt.Errorf("read %s: %w", display, err)
	}

	if trimmed := strings.TrimSpace(string(data)); trimmed == "" || trimmed == "null" {
		return mo.None[[]byte](), nil
	}

	return mo.Some(data), nil
}

// encodeDocument renders the merged document back into the bytes that go in the file. The encoder
// is built by hand rather than through json.Marshal because Marshal escapes <, > and & into \u
// sequences, which would silently rewrite a permission rule like "Bash(a && b)" in a file codefall
// does not own. Encode writes the trailing newline itself.
func encodeDocument(document map[string]any, dest string) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(document); err != nil {
		return nil, fmt.Errorf("encode %s: %w", dest, err)
	}

	return buffer.Bytes(), nil
}
