// Package initcmd is the facade for `codefall init`, the command that makes a directory ready for
// codefall: it writes the .codefall/settings.json that doctor checks and installs the codefall
// plugin for the harness, and the step that initialises the tracker is added to the same run.
//
// The package is named initcmd rather than init because a package called init cannot be imported
// without an alias — `import ".../internal/init"` does not compile, since init must be a func.
// initcmd keeps the directory named after the command it holds. The command it exports is still
// `init`.
//
// Exported identifiers here are the component's whole public API. Its layers live under this
// package's own internal/, where the compiler keeps them (ADR-BASE-02).
package initcmd

import (
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	"github.com/lividlabs/codefall-cli/internal/initcmd/internal/application"
	"github.com/lividlabs/codefall-cli/internal/initcmd/internal/infrastructure"
	"github.com/lividlabs/codefall-cli/internal/initcmd/internal/presentation"
)

// Register wires initcmd's object graph into the app's injector. This is the only place the
// component's concrete types are named: main cannot import them (ADR-GO-01 as amended 2026-08-19).
//
// Every provider's static return type is the interface, never the concrete type — a provider that
// returns the concrete type compiles and then fails at runtime with "could not find service".
func Register(injector do.Injector) {
	do.Provide(injector, func(do.Injector) (application.FileSystem, error) {
		return infrastructure.NewOSFileSystem(), nil
	})

	do.Provide(injector, func(do.Injector) (application.CommandRunner, error) {
		return infrastructure.NewExecCommandRunner(), nil
	})

	do.Provide(injector, func(i do.Injector) (presentation.InitializeUseCase, error) {
		return application.NewInitialize(
			do.MustInvoke[application.FileSystem](i),
			do.MustInvoke[application.CommandRunner](i),
		), nil
	})
}

// Command returns initcmd's command for the composition root to mount. A *cobra.Command is a delivery
// type, not a domain entity, so ADR-001 does not apply to it.
func Command(injector do.Injector) *cobra.Command {
	return presentation.NewInitCommand(do.MustInvoke[presentation.InitializeUseCase](injector))
}
