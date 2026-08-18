// Command codefall is the CLI's entry point. It becomes the app's composition root — the only place
// naming concrete implementations (ADR-GO-01) — once there is an object graph to wire. Today there
// isn't one, so this is a banner and nothing else.
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "codefall:", err)
		os.Exit(1)
	}
}

// run holds the body so it is testable. Taking the output stream as a parameter keeps main() the
// only part of the app bound to the real process.
func run(_ []string, out io.Writer) error {
	if _, err := fmt.Fprintln(out, "Welcome to Codefall"); err != nil {
		return fmt.Errorf("write banner: %w", err)
	}

	return nil
}
