// Package presentation builds doctor's command. It is thin: it reads the working directory, calls
// the use case, and prints what comes back.
package presentation

import (
	"context"
	"fmt"
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/cobra"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/domain"
)

// DiagnoseUseCase is what the command needs from the application layer, declared by its consumer.
type DiagnoseUseCase interface {
	Run(ctx context.Context, dir string) (domain.Report, error)
}

// Only the status word is styled: the rest of a line is a path, a version, or a command to type, and
// colour there would be decoration rather than information.
var statusStyles = map[domain.Status]lipgloss.Style{
	domain.StatusPass: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Green),
	domain.StatusWarn: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Yellow),
	domain.StatusFail: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Red),
}

// NewDoctorCommand builds `codefall doctor`.
func NewDoctorCommand(diagnose DiagnoseUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that this directory and your tools are ready for codefall",
		Long:  "Reports on .codefall/settings.json, Beads, and gh. It never changes anything.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// The working directory is the one piece of environment the command reads; the use case
			// receives it as a value.
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("doctor: %w", err)
			}

			report, err := diagnose.Run(cmd.Context(), dir)
			if err != nil {
				return fmt.Errorf("doctor: %w", err)
			}

			// One colorprofile writer for the whole report; lipgloss.Fprint* would build one per line.
			if err := renderReport(colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ()), report); err != nil {
				return err
			}

			// Fang renders this; nothing here calls os.Exit.
			if failed := report.Failed(); failed > 0 {
				return fmt.Errorf("doctor: %d %s failed", failed, plural(failed, "check", "checks"))
			}

			return nil
		},
	}
}

func renderReport(w io.Writer, report domain.Report) error {
	for _, result := range report.Results() {
		if _, err := fmt.Fprintln(w, formatLine(result)); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}

	return nil
}

func formatLine(result domain.Result) string {
	status := result.Status.String()
	if style, ok := statusStyles[result.Status]; ok {
		status = style.Render(status)
	}

	line := status + "  " + result.Check.Title

	if detail, ok := result.Detail.Get(); ok {
		line += " — " + detail
	}

	if remedy, ok := result.Remedy.Get(); ok {
		line += " (fix: " + remedy + ")"
	}

	return line
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}

	return many
}
