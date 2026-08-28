// Package presentation builds doctor's command. It is thin: it reads the working directory, calls
// the use case, and prints what comes back.
package presentation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/lividlabs/codefall-cli/internal/doctor/internal/domain"
)

// heading is the report's first line, drawn as a chip the way Fang draws its ERROR header: the same
// foreground-on-background pair, padding, and weight, in the colour Fang gives a title. The
// colorprofile writer strips the styling on a non-terminal stdout and under NO_COLOR, leaving the
// word itself.
const heading = "DOCTOR SUMMARY"

// errCancelled is what an interrupted run reports. It is returned as it is rather than wrapped, so
// Fang renders that sentence and not a prefix in front of it.
var errCancelled = errors.New("doctor cancelled")

var headingStyle = sync.OnceValue(func() lipgloss.Style {
    c := lipgloss.LightDark(hasDarkBackground())
    return lipgloss.NewStyle().
        Bold(true).
        Foreground(c(lipgloss.Color("#C8DADA"), lipgloss.Color("#C8DADA"))).
        Background(c(lipgloss.Color("#143337"), lipgloss.Color("#3A7680"))).
        Padding(0, 1)
})

// DiagnoseUseCase is what the command needs from the application layer, declared by its consumer.
type DiagnoseUseCase interface {
	Run(ctx context.Context, dir string) (domain.Report, error)
}

// Only the status mark is styled: the rest of a line is a path, a version, or a command to type, and
// colour there would be decoration rather than information.
//
// The palette is the teal trail from the brand image — desaturated teals for pass, muted amber for
// warn, muted coral for fail — picked light/dark-aware the way Fang's own colour scheme picks its.
// Fail and warn are warm hues so they do not collapse into the teal.
//
// The scheme is built once, on first use, because deciding it means asking the terminal a question.
var statusStyles = sync.OnceValue(func() map[domain.Status]lipgloss.Style {
    c := lipgloss.LightDark(hasDarkBackground())
    return map[domain.Status]lipgloss.Style{
        domain.StatusPass: lipgloss.NewStyle().Bold(true).Foreground(c(lipgloss.Color("#2F6B6B"), lipgloss.Color("#6AB3B3"))),
        domain.StatusWarn: lipgloss.NewStyle().Bold(true).Foreground(c(lipgloss.Color("#7E6217"), lipgloss.Color("#D9B44A"))),
        domain.StatusFail: lipgloss.NewStyle().Bold(true).Foreground(c(lipgloss.Color("#9E3A3A"), lipgloss.Color("#E08878"))),
    }
})

// hasDarkBackground asks the terminal for its background colour, the way Fang does before building
// its styles. When stdout is not a terminal there is nothing to ask and nothing to see — colorprofile
// strips the colour on its way out — so the answer is the dark variant, which is the one lipgloss
// itself falls back to.
func hasDarkBackground() bool {
	if !stdoutIsTerminal() {
		return true
	}

	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}

// stdoutIsTerminal decides both questions that depend on who is reading: whether the terminal can be
// asked about its background, and whether a spinner has anywhere to run. A pipe, a file, CI, and the
// tests all answer no.
func stdoutIsTerminal() bool {
	return term.IsTerminal(os.Stdout.Fd())
}

// Secondary text is dimmed and tinted teal-grey: the versions and accounts in a header, and the
// remedy under a problem. What is left at full strength is the sentence that says what is wrong,
// which is the only thing a reader has to take in.
var faintStyle = sync.OnceValue(func() lipgloss.Style {
	c := lipgloss.LightDark(hasDarkBackground())

	return lipgloss.NewStyle().Faint(true).Foreground(c(lipgloss.Color("#526D71"), lipgloss.Color("#8AA3A8")))
})

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

			report, err := runDiagnose(cmd.Context(), diagnose, dir)
			if err != nil {
				if errors.Is(err, errCancelled) {
					return err
				}

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

// spinnerLabel is what the spinner says while the checks run. The checks shell out to bd and gh, so
// a run is slow enough to be worth acknowledging.
const spinnerLabel = "Running checks…"

// runDiagnose runs the use case, spinning while it works when there is a terminal to spin on.
// Anywhere else — a pipe, a file, CI, a test — it calls the use case directly and the command writes
// exactly the bytes it wrote before the spinner existed.
func runDiagnose(ctx context.Context, diagnose DiagnoseUseCase, dir string) (domain.Report, error) {
	if !stdoutIsTerminal() {
		return diagnose.Run(ctx, dir)
	}

	// WithContext is what makes a cancelled command abort: the program stops with an error rather
	// than spinning until the use case notices.
	final, err := tea.NewProgram(newDiagnoseSpinner(ctx, diagnose, dir), tea.WithContext(ctx)).Run()
	if err != nil {
		return domain.Report{}, spinnerError(err)
	}

	model, ok := final.(diagnoseSpinner)
	if !ok {
		return domain.Report{}, fmt.Errorf("spinner: finished as %T, want %T", final, diagnoseSpinner{})
	}

	return model.report, model.err
}

// spinnerError is what a terminal program's failure means to the command. Bubble Tea reports Ctrl-C
// as a program that was interrupted, which is a person stopping the run rather than anything going
// wrong, so it is said in those terms.
func spinnerError(err error) error {
	if errors.Is(err, tea.ErrInterrupted) {
		return errCancelled
	}

	return fmt.Errorf("spinner: %w", err)
}

// diagnosedMsg carries the use case's outcome back into the program. It holds the error rather than
// failing the program with it, because the command decides what an error means, not the spinner.
type diagnosedMsg struct {
	report domain.Report
	err    error
}

// diagnoseSpinner is the whole terminal program: a frame, a label, and one command running the use
// case. It draws nothing once the answer arrives, so the report prints on a clean line.
type diagnoseSpinner struct {
	spinner spinner.Model
	run     tea.Cmd
	report  domain.Report
	err     error
	done    bool
}

func newDiagnoseSpinner(ctx context.Context, diagnose DiagnoseUseCase, dir string) diagnoseSpinner {
	return diagnoseSpinner{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.MiniDot),
			spinner.WithStyle(statusStyle(domain.StatusPass)),
		),
		run: func() tea.Msg {
			report, err := diagnose.Run(ctx, dir)

			return diagnosedMsg{report: report, err: err}
		},
	}
}

func (m diagnoseSpinner) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.run)
}

func (m diagnoseSpinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if done, ok := msg.(diagnosedMsg); ok {
		m.report, m.err, m.done = done.report, done.err, true

		return m, tea.Quit
	}

	var cmd tea.Cmd

	m.spinner, cmd = m.spinner.Update(msg)

	return m, cmd
}

func (m diagnoseSpinner) View() tea.View {
	if m.done {
		return tea.NewView("")
	}

	return tea.NewView(m.spinner.View() + " " + faintStyle().Render(spinnerLabel))
}

// renderReport prints the heading, then one header line per category and, under it, one line per
// problem. A check that passed says all it has to say in its section's header.
func renderReport(w io.Writer, report domain.Report) error {
	// The trailing newline is the blank line between the heading and the first section.
	if err := writeLine(w, headingStyle().Render(heading)+"\n"); err != nil {
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
	guide := faintStyle().PaddingRight(1)

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
			problem.Child(faintStyle().Render("fix: " + remedy))
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
	header := bracketedMark(section.Status()) + " " + section.Category.Title

	var details []string

	for _, result := range section.Results {
		if detail, ok := result.Detail.Get(); ok && result.Status == domain.StatusPass {
			details = append(details, detail)
		}
	}

	if len(details) > 0 {
		header += " " + faintStyle().Render("("+strings.Join(details, ", ")+")")
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

// mark is the bare glyph a problem line hangs off, in its status's colour.
func mark(status domain.Status) string {
	return statusStyle(status).Render(glyph(status))
}

// bracketedMark is the header's form of the same thing. The brackets are rendered inside the style
// rather than around it, so the three characters read as one token instead of a coloured glyph
// between two uncoloured ones.
func bracketedMark(status domain.Status) string {
	return statusStyle(status).Render("[" + glyph(status) + "]")
}

func glyph(status domain.Status) string {
	if g, ok := statusMarks[status]; ok {
		return g
	}

	return "?"
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
