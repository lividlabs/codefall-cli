// Command codefall is the CLI's entry point and the app's composition root: it builds the one
// injector, calls each component's registration function, mounts the commands those components
// expose, and hands the tree to Fang (ADR-GO-01, ADR-002).
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	"github.com/lividlabs/codefall-cli/internal/doctor"
)

func main() {
	// Fang has already rendered the error; only the exit status is left.
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}

// run holds the body so it is testable. Taking the arguments and both streams as parameters keeps
// main() the only part of the app bound to the real process.
func run(args []string, stdout, stderr io.Writer) error {
	// Cobra reads os.Args when its argument slice is nil, which under `go test` is the test binary's
	// own flags.
	if args == nil {
		args = []string{}
	}

	injector := do.New()
	defer func() {
		if report := injector.Shutdown(); !report.Succeed {
			_, _ = fmt.Fprintln(stderr, report.Error())
		}
	}()

	doctor.Register(injector)

	root := &cobra.Command{
		Use:   "codefall",
		Short: "Codefall's command-line tool",
	}
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.AddCommand(doctor.Command(injector))

	return fang.Execute(context.Background(), root)
}
