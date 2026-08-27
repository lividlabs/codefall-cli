// Package presentation builds doctor's command. It is thin: it reads the working directory, calls
// the use case, and prints what comes back.
package presentation

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/cobra"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/domain"
)

// DiagnoseUseCase is what the command needs from the application layer, declared by its consumer.
type DiagnoseUseCase interface {
	Run(ctx context.Context, dir string) (domain.Report, error)
}

// Only the status mark is styled: the rest of a line is a path, a version, or a command to type, and
// colour there would be decoration rather than information.
var statusStyles = map[domain.Status]lipgloss.Style{
	domain.StatusPass: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Green),
	domain.StatusWarn: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Yellow),
	domain.StatusFail: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Red),
}

// The mark each status prints, in the header's brackets and in front of a problem line.
var statusMarks = map[domain.Status]string{
	domain.StatusPass: "✓",
	domain.StatusWarn: "!",
	domain.StatusFail: "✗",
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
			out := colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ())

			if err := renderReport(out, report); err != nil {
				return err
			}

			issues := report.SectionsWithIssues()

			// Fang renders this; nothing here calls os.Exit. A failing run prints no summary line
			// because the error is the summary.
			if report.Failed() > 0 {
				return fmt.Errorf("doctor found issues in %d %s",
					issues, plural(issues, "category", "categories"))
			}

			return writeLine(out, summaryLine(issues))
		},
	}
}

// renderReport prints one header line per category and, under it, one line per problem. A check that
// passed says all it has to say in its section's header.
func renderReport(w io.Writer, report domain.Report) error {
	for _, section := range report.Sections() {
		if err := writeLine(w, sectionHeader(section)); err != nil {
			return err
		}

		for _, result := range section.Results {
			if result.Status == domain.StatusPass {
				continue
			}

			if err := writeLine(w, problemLine(result)); err != nil {
				return err
			}

			if remedy, ok := result.Remedy.Get(); ok {
				if err := writeLine(w, "      fix: "+remedy); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// sectionHeader is the category's status, its title, and whatever its passing checks reported about
// themselves — the versions and the account, the things worth knowing when nothing is wrong.
func sectionHeader(section domain.Section) string {
	header := "[" + mark(section.Status()) + "] " + section.Category.Title

	var details []string

	for _, result := range section.Results {
		if detail, ok := result.Detail.Get(); ok && result.Status == domain.StatusPass {
			details = append(details, detail)
		}
	}

	if len(details) > 0 {
		header += " (" + strings.Join(details, ", ") + ")"
	}

	return header
}

// problemLine is a warning or a failure, indented under its header. The check's title is not printed:
// the detail is written as a sentence that stands on its own.
func problemLine(result domain.Result) string {
	line := "    " + mark(result.Status)

	if detail, ok := result.Detail.Get(); ok {
		line += " " + detail
	}

	return line
}

func summaryLine(issues int) string {
	if issues == 0 {
		return "• No issues found."
	}

	return fmt.Sprintf("• Doctor found issues in %d %s.", issues, plural(issues, "category", "categories"))
}

func mark(status domain.Status) string {
	glyph, ok := statusMarks[status]
	if !ok {
		glyph = "?"
	}

	if style, ok := statusStyles[status]; ok {
		return style.Render(glyph)
	}

	return glyph
}

func writeLine(w io.Writer, line string) error {
	if _, err := fmt.Fprintln(w, line); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	return nil
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}

	return many
}
