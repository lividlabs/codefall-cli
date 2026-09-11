package initcmd_test

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/samber/do/v2"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd"
	"github.com/lividlabs/codefall-cli/extensions"
)

// A provider whose static return type is the concrete type compiles and then fails at runtime with
// "could not find service". Resolving the whole graph through a real injector is the only thing that
// catches that, so this test does exactly what the composition root does.
func TestRegisterProvidesEverythingCommandNeeds(t *testing.T) {
	injector := do.New()
	t.Cleanup(func() { _ = injector.Shutdown() })

	initcmd.Register(injector)

	cmd := initcmd.Command(injector)
	if cmd == nil {
		t.Fatal("Command returned nil")
	}

	// The package is initcmd; the command is init.
	if got := cmd.Name(); got != "init" {
		t.Errorf("Command().Name() = %q, want %q", got, "init")
	}
}

// The hook step's table and the extension step's destination map agree with the embedded tree by
// three hand-maintained consistencies: every definition exists, every script a definition's
// commands point at exists under hooks/shared/, and the path they name sits under the directory
// the extension step copies that directory into. The application layer's tests run over a fake
// copy of the definitions, so nothing there would fail on a rename — this test runs over the
// real tree and does. It states the table again the way the schema test states the format: a
// pin, not a second source of truth.
func TestHookDefinitionsPointAtScriptsThatLand(t *testing.T) {
	tree := extensions.Files()

	for _, tc := range []struct {
		harness string
		source  string
		destDir string
	}{
		{"claude-code", "hooks/claude/hooks.json", ".claude"},
		{"codex", "hooks/codex/hooks.json", ".agents"},
		{"antigravity", "hooks/antigravity/hooks.json", ".agents"},
	} {
		t.Run(tc.harness, func(t *testing.T) {
			body, err := fs.ReadFile(tree, tc.source)
			if err != nil {
				t.Fatalf("read %s from the embedded tree: %v", tc.source, err)
			}

			var document any
			if err := json.Unmarshal(body, &document); err != nil {
				t.Fatalf("decode %s: %v", tc.source, err)
			}

			commands := commandsIn(document)
			if len(commands) == 0 {
				t.Fatalf("%s declares no commands", tc.source)
			}

			for _, command := range commands {
				at := strings.Index(command, "hooks/shared/")
				if at < 0 {
					continue
				}

				if prefix := command[:at]; !strings.HasSuffix(prefix, tc.destDir+"/") {
					t.Errorf("command %q points outside %s/, where the extension step copies the shared scripts", command, tc.destDir)
				}

				name := scriptName(command[at:])
				if _, err := fs.Stat(tree, "hooks/shared/"+name); err != nil {
					t.Errorf("hooks/shared/%s: %v — a renamed script would leave the %s guard pointing at nothing", name, err, tc.harness)
				}
			}
		})
	}
}

// The OpenCode plugin is the one definition that is not JSON, so its pin is textual: the guard it
// delegates to is the same shared script the others name.
func TestOpenCodePluginPointsAtTheSharedScript(t *testing.T) {
	tree := extensions.Files()

	body, err := fs.ReadFile(tree, "hooks/opencode/codefall.js")
	if err != nil {
		t.Fatalf("read hooks/opencode/codefall.js from the embedded tree: %v", err)
	}

	const referenced = ".agents/hooks/shared/codefall-block-merge-to-main.sh"
	if !strings.Contains(string(body), referenced) {
		t.Errorf("the plugin does not name %s — a renamed script would leave its guard pointing at nothing", referenced)
	}

	if _, err := fs.Stat(tree, "hooks/shared/codefall-block-merge-to-main.sh"); err != nil {
		t.Errorf("hooks/shared/codefall-block-merge-to-main.sh: %v", err)
	}
}

// commandsIn collects every command string in a decoded hook document, whatever depth the
// harness's format keeps it at.
func commandsIn(value any) []string {
	switch shaped := value.(type) {
	case map[string]any:
		var commands []string
		if command, ok := shaped["command"].(string); ok {
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

// scriptName takes the file name off the front of what follows "hooks/shared/", stopping at the
// first space or closing quote.
func scriptName(rest string) string {
	if cut := strings.IndexAny(rest, " \""); cut >= 0 {
		return rest[len("hooks/shared/"):cut]
	}
	return strings.TrimSuffix(rest[len("hooks/shared/"):], "\n")
}
