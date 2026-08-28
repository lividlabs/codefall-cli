package ui

import (
	"fmt"
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

// StdoutIsTerminal decides both questions that depend on who is reading: whether the terminal can be
// asked about its background, and whether a spinner has anywhere to run. A pipe, a file, CI, and the
// tests all answer no.
func StdoutIsTerminal() bool {
	return term.IsTerminal(os.Stdout.Fd())
}

// HasDarkBackground asks the terminal for its background colour, the way Fang does before building
// its styles. When stdout is not a terminal there is nothing to ask and nothing to see — colorprofile
// strips the colour on its way out — so the answer is the dark variant, which is the one lipgloss
// itself falls back to.
func HasDarkBackground() bool {
	if !StdoutIsTerminal() {
		return true
	}

	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}

// NewWriter wraps a command's output stream in the colour-profile writer every styled line goes
// through, so a pipe, a file, and NO_COLOR all get plain text (ADR-002). One writer serves a whole
// run; lipgloss.Fprint* would build one per line.
func NewWriter(w io.Writer) io.Writer {
	return colorprofile.NewWriter(w, os.Environ())
}

// WriteLine writes one rendered line and says which job the failure belongs to.
func WriteLine(w io.Writer, line string) error {
	if _, err := fmt.Fprintln(w, line); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	return nil
}
