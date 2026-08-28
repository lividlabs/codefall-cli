package setup_test

import (
	"testing"

	"github.com/samber/do/v2"

	"github.com/lividlabs/codefall-cli/internal/setup"
)

// A provider whose static return type is the concrete type compiles and then fails at runtime with
// "could not find service". Resolving the whole graph through a real injector is the only thing that
// catches that, so this test does exactly what the composition root does.
func TestRegisterProvidesEverythingCommandNeeds(t *testing.T) {
	injector := do.New()
	t.Cleanup(func() { _ = injector.Shutdown() })

	setup.Register(injector)

	cmd := setup.Command(injector)
	if cmd == nil {
		t.Fatal("Command returned nil")
	}

	// The package is setup; the command is init.
	if got := cmd.Name(); got != "init" {
		t.Errorf("Command().Name() = %q, want %q", got, "init")
	}
}
