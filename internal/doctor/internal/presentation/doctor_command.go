// Package presentation builds doctor's command. It is thin: it reads the working directory, calls
// the use case, and prints what comes back.
package presentation

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/domain"
)

// heading is the report's first line, drawn as a chip the way Fang draws its ERROR header: the same
// foreground-on-background pair, padding, and weight, in the colour Fang gives a title. The
// colorprofile writer strips the styling on a non-terminal stdout and under NO_COLOR, leaving the
// word itself.
const heading = "DOCTOR SUMMARY"

var headingStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(charmtone.Butter).
	Background(charmtone.Charple).
	Padding(0, 1)

// DiagnoseUseCase is what the command needs from the application layer, declared by its consumer.
type DiagnoseUseCase interface {
	Run(ctx context.Context, dir string) (domain.Report, error)
}

// Only the status mark is styled: the rest of a line is a path, a version, or a command to type, and
// colour there would be decoration rather than information.
//
// The colours come from the Charm palette, picked light/dark-aware the way Fang's own colour scheme
// picks its. Failure is Cherry in both, because it is the one mark that has to read as alarming
// whatever the terminal looks like — and it is the colour Fang paints its error header.
//
// The scheme is built once, on first use, because deciding it means asking the terminal a question.
var statusStyles = sync.OnceValue(func() map[domain.Status]lipgloss.Style {
	c := lipgloss.LightDark(hasDarkBackground())

	return map[domain.Status]lipgloss.Style{
		domain.StatusPass: lipgloss.NewStyle().Bold(true).Foreground(c(charmtone.Guac, charmtone.Julep)),
		domain.StatusWarn: lipgloss.NewStyle().Bold(true).Foreground(c(charmtone.Mustard, charmtone.Citron)),
		domain.StatusFail: lipgloss.NewStyle().Bold(true).Foreground(charmtone.Cherry),
	}
})

// hasDarkBackground asks the terminal for its background colour, the way Fang does before building
// its styles. When stdout is not a terminal there is nothing to ask and nothing to see — colorprofile
// strips the colour on its way out — so the answer is the dark variant, which is the one lipgloss
// itself falls back to.
func hasDarkBackground() bool {
	if !term.IsTerminal(os.Stdout.Fd()) {
		return true
	}

	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}

// Secondary text is dimmed rather than coloured: the versions and accounts in a header, and the
// remedy under a problem. What is left at full strength is the sentence that says what is wrong,
// which is the only thing a reader has to take in.
var faintStyle = lipgloss.NewStyle().Faint(true)

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

// renderReport prints the heading, then one header line per category and, under it, one line per
// problem. A check that passed says all it has to say in its section's header.
func renderReport(w io.Writer, report domain.Report) error {
	// The trailing newline is the blank line between the heading and the first section.
	if err := writeLine(w, headingStyle.Render(heading)+"\n"); err != nil {
		return err
	}

	for _, section := range report.Sections() {
		if err := writeLine(w, sectionTree(section)); err != nil {
			return err
		}
	}

	return nil
}

// sectionTree draws one category: the header is the root, each problem is a child of it, and a
// remedy is a child of the problem it repairs. A section whose checks all passed is a root on its
// own, because its header already says everything there is to say.
func sectionTree(section domain.Section) string {
	guide := faintStyle.PaddingRight(1)

	root := tree.Root(sectionHeader(section)).
		Enumerator(enumerator).
		Indenter(indenter).
		EnumeratorStyle(guide).
		IndenterStyle(guide)

	for _, result := range section.Results {
		if result.Status == domain.StatusPass {
			continue
		}

		problem := tree.Root(problemLine(result))

		if remedy, ok := result.Remedy.Get(); ok {
			problem.Child(faintStyle.Render("fix: " + remedy))
		}

		root.Child(problem)
	}

	return root.String()
}

// enumerator draws the elbow in front of a child, rounded on the last one. Two characters rather
// than the tree package's three, which keeps a remedy's text at the column its problem's mark sits
// in.
func enumerator(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return "╰─"
	}

	return "├─"
}

// indenter carries the vertical down the left of a problem's remedy while the section still has
// problems to come, and drops it once there are none.
func indenter(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return "  "
	}

	return "│ "
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
		header += " " + faintStyle.Render("("+strings.Join(details, ", ")+")")
	}

	return header
}

// problemLine is a warning or a failure, hung off its header by the tree. The check's title is not
// printed: the detail is written as a sentence that stands on its own.
func problemLine(result domain.Result) string {
	line := mark(result.Status)

	if detail, ok := result.Detail.Get(); ok {
		line += " " + detail
	}

	return line
}

// summaryLine closes a run that had nothing to fail on, so it carries the status it is reporting:
// the pass colour when the report is clean, the warning colour when it is not. A failing run never
// reaches here — its error is its summary, and Fang renders that.
func summaryLine(issues int) string {
	if issues == 0 {
		return statusStyle(domain.StatusPass).Render("• No issues found.")
	}

	return statusStyle(domain.StatusWarn).Render(
		fmt.Sprintf("• Doctor found issues in %d %s.", issues, plural(issues, "category", "categories")))
}

func mark(status domain.Status) string {
	glyph, ok := statusMarks[status]
	if !ok {
		glyph = "?"
	}

	return statusStyle(status).Render(glyph)
}

// statusStyle is the one place a status becomes a colour, so the marks and the summary line cannot
// drift apart. An unknown status is left unstyled.
func statusStyle(status domain.Status) lipgloss.Style {
	if style, ok := statusStyles()[status]; ok {
		return style
	}

	return lipgloss.NewStyle()
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
