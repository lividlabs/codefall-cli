package presentation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/lividlabs/codefall-cli/internal/initcmd/internal/application"
	"github.com/lividlabs/codefall-cli/internal/initcmd/internal/domain"
)

// openingLabel is what the spinner says before the first step reports itself.
const openingLabel = "Setting up…"

// runInitialize runs the use case, spinning while each step works when there is a terminal to spin
// on. Anywhere else — a pipe, a file, CI, a test — it calls the use case directly and each finished
// step prints as a plain line.
func runInitialize(
	ctx context.Context, initialize InitializeUseCase, request application.Request, out io.Writer,
) (domain.Report, error) {
	if !stdoutIsTerminal() {
		observer := &writingObserver{w: out}

		report, err := initialize.Run(ctx, request, observer)
		if err != nil {
			// A line that could not be written is dropped here on purpose: the run's own error is
			// what the reader needs, and a failed write is only how they would have been told about
			// a step that did finish.
			return domain.Report{}, err
		}

		return report, observer.err
	}

	// The observer is built before the program so it can send into it, and used only from the
	// command the program itself starts — so the program is running by the time anything is sent.
	observer := &programObserver{}

	// WithContext is what makes a cancelled command abort: the program stops with an error rather
	// than spinning until the use case notices.
	program := tea.NewProgram(newInitSpinner(ctx, initialize, request, observer), tea.WithContext(ctx))
	observer.program = program

	final, err := program.Run()
	if err != nil {
		return domain.Report{}, spinnerError(err)
	}

	model, ok := final.(initSpinner)
	if !ok {
		return domain.Report{}, fmt.Errorf("spinner: finished as %T, want %T", final, initSpinner{})
	}

	// The program's own frames are gone: the renderer clears what it drew when it stops, which is why
	// doctor prints its report after the spinner rather than from inside it. The lines the model
	// collected are written again here, where they stay — including the ones from before a step that
	// failed, which the report the use case returned does not carry.
	for _, line := range model.lines {
		if err := writeLine(out, line); err != nil {
			return domain.Report{}, err
		}
	}

	return model.report, model.err
}

// spinnerError is what a terminal program's failure means to the command. Bubble Tea reports Ctrl-C
// as a program that was interrupted; that is the same event the survey reports as an abort, so it
// gets the same sentence rather than a second phrasing of one keystroke.
func spinnerError(err error) error {
	if errors.Is(err, tea.ErrInterrupted) {
		return errCancelled
	}

	return fmt.Errorf("spinner: %w", err)
}

// writingObserver is what watches a run with no terminal to draw on: each finished step is a line,
// written as it finishes. A write that fails is kept rather than dropped, because the caller decides
// what a failed write means.
type writingObserver struct {
	w   io.Writer
	err error
}

func (o *writingObserver) StepStarted(domain.Step) {}

func (o *writingObserver) StepFinished(result domain.StepResult) {
	if o.err != nil {
		return
	}

	o.err = writeLine(o.w, stepLine(result))
}

// programObserver carries the run's progress into the terminal program. Send is safe from the
// goroutine the use case runs on and is a no-op once the program has stopped, so a cancelled run
// cannot leave a step waiting to be heard.
type programObserver struct {
	program *tea.Program
}

func (o *programObserver) StepStarted(step domain.Step) {
	o.program.Send(stepStartedMsg{step: step})
}

func (o *programObserver) StepFinished(result domain.StepResult) {
	o.program.Send(stepFinishedMsg{result: result})
}

// The three things the program hears: a step starting, a step finishing, and the run being over.
type (
	stepStartedMsg  struct{ step domain.Step }
	stepFinishedMsg struct{ result domain.StepResult }

	// initializedMsg holds the error rather than failing the program with it, because the command
	// decides what an error means, not the spinner.
	initializedMsg struct {
		report domain.Report
		err    error
	}
)

// initSpinner is the whole terminal program: the steps that have finished, the label of the step
// running now, and one command running the use case.
//
// The finished lines are part of the view, so a run with several steps shows what it has done while
// it is still working. Nothing the program draws survives it, though — the renderer clears its own
// output when it stops — so the lines are also kept in the model for the caller to write again once
// the program is gone.
type initSpinner struct {
	spinner spinner.Model
	run     tea.Cmd
	label   string
	lines   []string
	report  domain.Report
	err     error
	done    bool
}

func newInitSpinner(
	ctx context.Context, initialize InitializeUseCase, request application.Request, observer application.Observer,
) initSpinner {
	return initSpinner{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.MiniDot),
			spinner.WithStyle(outcomeStyle(domain.OutcomeDone)),
		),
		label: openingLabel,
		run: func() tea.Msg {
			report, err := initialize.Run(ctx, request, observer)

			return initializedMsg{report: report, err: err}
		},
	}
}

func (m initSpinner) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.run)
}

func (m initSpinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stepStartedMsg:
		m.label = msg.step.Title + "…"

		return m, nil
	case stepFinishedMsg:
		m.lines = append(m.lines, stepLine(msg.result))

		return m, nil
	case initializedMsg:
		m.report, m.err, m.done = msg.report, msg.err, true

		return m, tea.Quit
	}

	var cmd tea.Cmd

	m.spinner, cmd = m.spinner.Update(msg)

	return m, cmd
}

// View is the finished steps, and under them the spinner and what it is waiting on. The last frame
// is empty, as doctor's is, so what the caller writes afterwards starts on a clean line.
func (m initSpinner) View() tea.View {
	if m.done {
		return tea.NewView("")
	}

	report := strings.Join(m.lines, "\n")
	if report != "" {
		report += "\n"
	}

	return tea.NewView(report + m.spinner.View() + " " + faintStyle().Render(m.label))
}
