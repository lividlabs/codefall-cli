package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Progress is what a job may say about itself while it runs: what it is doing now, and what it has
// finished. A job with nothing to report ignores it and still gets the spinner.
type Progress interface {
	// Label replaces what the spinner says it is waiting on.
	Label(label string)
	// Line adds a finished line, shown under the spinner and written out when the run is over.
	Line(line string)
}

// RunWithSpinner runs work, spinning while it works when there is a terminal to spin on. Anywhere
// else — a pipe, a file, CI, a test — it calls work directly and writes exactly the bytes it would
// have written before the spinner existed.
//
// Lines reported through Progress reach w either way. Under the spinner they are drawn as they
// arrive and written again once the program has stopped, because Bubble Tea's renderer clears what
// it drew; on the plain path they are written as they arrive and not again. They are written even
// when work fails, because a step that did finish is news whatever happened after it.
//
// Ctrl-C is a person stopping the run rather than something going wrong, so it is reported as the
// caller's own cancelled error, returned as it is for Fang to render without a prefix in front of
// it.
func RunWithSpinner[T any](
	ctx context.Context,
	w io.Writer,
	label string,
	cancelled error,
	work func(context.Context, Progress) (T, error),
) (T, error) {
	if !StdoutIsTerminal() {
		return runPlain(ctx, w, work)
	}

	return runSpinning(ctx, w, label, cancelled, work)
}

// runPlain is the path with no terminal to draw on: each finished line is written as it finishes.
func runPlain[T any](
	ctx context.Context, w io.Writer, work func(context.Context, Progress) (T, error),
) (T, error) {
	progress := &writingProgress{w: w}

	value, err := work(ctx, progress)
	if err != nil {
		// A line that could not be written is dropped here on purpose: the run's own error is what
		// the reader needs, and a failed write is only how they would have been told about a step
		// that did finish.
		var zero T

		return zero, err
	}

	return value, progress.err
}

// runSpinning is the terminal path: one Bubble Tea program whose only command runs the job.
func runSpinning[T any](
	ctx context.Context,
	w io.Writer,
	label string,
	cancelled error,
	work func(context.Context, Progress) (T, error),
) (T, error) {
	var zero T

	// The progress is built before the program so it can send into it, and used only from the
	// command the program itself starts — so the program is running by the time anything is sent.
	progress := &programProgress{}

	// WithContext is what makes a cancelled command abort: the program stops with an error rather
	// than spinning until the job notices.
	program := tea.NewProgram(newSpinnerModel(ctx, label, progress, work), tea.WithContext(ctx))
	progress.program = program

	final, err := program.Run()
	if err != nil {
		return zero, spinnerError(err, cancelled)
	}

	model, ok := final.(spinnerModel[T])
	if !ok {
		return zero, fmt.Errorf("spinner: finished as %T, want %T", final, spinnerModel[T]{})
	}

	// The program's own frames are gone: the renderer clears what it drew when it stops. The lines
	// the model collected are written again here, where they stay — including the ones from before a
	// step that failed, which the value the job returned does not carry.
	for _, line := range model.lines {
		if err := WriteLine(w, line); err != nil {
			return zero, err
		}
	}

	return model.value, model.err
}

// spinnerError is what a terminal program's failure means to the caller. Bubble Tea reports Ctrl-C
// as a program that was interrupted, which is a person stopping the run rather than anything going
// wrong, so it is said in the caller's own terms.
func spinnerError(err, cancelled error) error {
	if errors.Is(err, tea.ErrInterrupted) {
		return cancelled
	}

	return fmt.Errorf("spinner: %w", err)
}

// writingProgress watches a run with no terminal to draw on. A write that fails is kept rather than
// dropped, because the caller decides what a failed write means.
type writingProgress struct {
	w   io.Writer
	err error
}

func (p *writingProgress) Label(string) {}

func (p *writingProgress) Line(line string) {
	if p.err != nil {
		return
	}

	p.err = WriteLine(p.w, line)
}

// programProgress carries the run's progress into the terminal program. Send is safe from the
// goroutine the job runs on and is a no-op once the program has stopped, so a cancelled run cannot
// leave a step waiting to be heard.
type programProgress struct {
	program *tea.Program
}

func (p *programProgress) Label(label string) { p.program.Send(labelMsg(label)) }

func (p *programProgress) Line(line string) { p.program.Send(lineMsg(line)) }

// The three things the program hears: the label changing, a line finishing, and the job being over.
type (
	labelMsg string
	lineMsg  string

	// doneMsg holds the error rather than failing the program with it, because the caller decides
	// what an error means, not the spinner.
	doneMsg[T any] struct {
		value T
		err   error
	}
)

// spinnerModel is the whole terminal program: the lines that have finished, the label of the work
// running now, and one command running the job.
//
// The finished lines are part of the view, so a run with several steps shows what it has done while
// it is still working. Nothing the program draws survives it, though, so the lines are also kept
// here for the caller to write again once the program is gone.
type spinnerModel[T any] struct {
	spinner spinner.Model
	run     tea.Cmd
	label   string
	lines   []string
	value   T
	err     error
	done    bool
}

func newSpinnerModel[T any](
	ctx context.Context, label string, progress Progress, work func(context.Context, Progress) (T, error),
) spinnerModel[T] {
	return spinnerModel[T]{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.MiniDot),
			spinner.WithStyle(Style(TonePrimary)),
		),
		label: label,
		run: func() tea.Msg {
			value, err := work(ctx, progress)

			return doneMsg[T]{value: value, err: err}
		},
	}
}

func (m spinnerModel[T]) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.run)
}

func (m spinnerModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case labelMsg:
		m.label = string(msg)

		return m, nil
	case lineMsg:
		m.lines = append(m.lines, string(msg))

		return m, nil
	case doneMsg[T]:
		m.value, m.err, m.done = msg.value, msg.err, true

		return m, tea.Quit
	}

	var cmd tea.Cmd

	m.spinner, cmd = m.spinner.Update(msg)

	return m, cmd
}

// View is the finished lines, and under them the spinner and what it is waiting on. The last frame
// is empty, so whatever the caller writes afterwards starts on a clean line.
func (m spinnerModel[T]) View() tea.View {
	if m.done {
		return tea.NewView("")
	}

	finished := strings.Join(m.lines, "\n")
	if finished != "" {
		finished += "\n"
	}

	return tea.NewView(finished + m.spinner.View() + " " + Style(ToneFaint).Render(m.label))
}
