package ui

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Under `go test` stdout is not a terminal, so RunWithSpinner takes the plain path: no program, and
// every line written as it is reported.
func TestRunWithSpinnerWithoutATerminalWritesEachLineAsItArrives(t *testing.T) {
	var out bytes.Buffer

	value, err := RunWithSpinner(t.Context(), &out, "Working…", errCancelledForTest,
		func(_ context.Context, progress Progress) (int, error) {
			progress.Label("first…")
			progress.Line("✓ first")

			if got := out.String(); got != "✓ first\n" {
				t.Errorf("after the first line the output was %q, want %q", got, "✓ first\n")
			}

			progress.Line("- second")

			return 7, nil
		})
	if err != nil {
		t.Fatalf("RunWithSpinner: %v", err)
	}

	if value != 7 {
		t.Errorf("value = %d, want 7", value)
	}

	if got := out.String(); got != "✓ first\n- second\n" {
		t.Errorf("output = %q, want %q", got, "✓ first\n- second\n")
	}
}

// A job that fails keeps the lines it already reported and hands back its own error with the zero
// value beside it.
func TestRunWithSpinnerWithoutATerminalKeepsTheLinesOfAFailedJob(t *testing.T) {
	failure := errors.New("no space left on device")

	var out bytes.Buffer

	value, err := RunWithSpinner(t.Context(), &out, "Working…", errCancelledForTest,
		func(_ context.Context, progress Progress) (string, error) {
			progress.Line("✓ first")

			return "unfinished", failure
		})
	if !errors.Is(err, failure) {
		t.Fatalf("RunWithSpinner = %v, want %v", err, failure)
	}

	if value != "" {
		t.Errorf("value = %q, want the zero value", value)
	}

	if got := out.String(); got != "✓ first\n" {
		t.Errorf("output = %q, want the line that did finish", got)
	}
}

// A write that fails is the run's error, not a line quietly dropped — and only the first one, since
// the rest are failing for the same reason.
func TestRunWithSpinnerWithoutATerminalReportsAFailedWrite(t *testing.T) {
	failure := errors.New("broken pipe")

	_, err := RunWithSpinner(t.Context(), failingWriter{err: failure}, "Working…", errCancelledForTest,
		func(_ context.Context, progress Progress) (int, error) {
			progress.Line("✓ first")
			progress.Line("- second")

			return 1, nil
		})
	if !errors.Is(err, failure) {
		t.Errorf("RunWithSpinner = %v, want it to wrap %v", err, failure)
	}
}

// Ctrl-C during the spinner is a person stopping the run, not something going wrong, so it becomes
// the caller's own sentence, returned as it is for Fang to render without a prefix.
func TestSpinnerError(t *testing.T) {
	if got := spinnerError(tea.ErrInterrupted, errCancelledForTest); !errors.Is(got, errCancelledForTest) {
		t.Errorf("spinnerError(tea.ErrInterrupted) = %v, want %v", got, errCancelledForTest)
	}

	failure := errors.New("no terminal")
	if got := spinnerError(failure, errCancelledForTest); !errors.Is(got, failure) ||
		!strings.HasPrefix(got.Error(), "spinner: ") {
		t.Errorf("spinnerError(%v) = %v, want it wrapped as a spinner failure", failure, got)
	}
}

// The spinner needs a terminal, so what is testable here is the model around it: that the job runs
// as its one command, that a label and a line reach the screen, that the answer quits the program,
// and that the last frame is empty so what the caller writes next starts on a clean line.
func TestSpinnerModelFollowsTheJobAndQuitsWithItsAnswer(t *testing.T) {
	failure := errors.New("no space left on device")

	model := newSpinnerModel(context.Background(), "Working…", &writingProgress{w: new(bytes.Buffer)},
		func(context.Context, Progress) (int, error) { return 7, failure })

	if !strings.Contains(model.View().Content, "Working…") {
		t.Errorf("View() = %q, want it to show the opening label", model.View().Content)
	}

	msg, ok := model.run().(doneMsg[int])
	if !ok {
		t.Fatalf("the spinner's command returned %T, want a doneMsg[int]", model.run())
	}

	// The error travels in the message rather than failing the program: the caller decides what an
	// error means, not the spinner.
	if !errors.Is(msg.err, failure) || msg.value != 7 {
		t.Errorf("message = %+v, want the job's own value and error", msg)
	}

	relabelled, _ := model.Update(labelMsg("Installing…"))

	running, ok := relabelled.(spinnerModel[int])
	if !ok {
		t.Fatalf("Update returned %T, want a spinnerModel[int]", relabelled)
	}

	if running.label != "Installing…" {
		t.Errorf("label = %q, want %q", running.label, "Installing…")
	}

	reported, _ := running.Update(lineMsg("✓ first"))

	running, ok = reported.(spinnerModel[int])
	if !ok {
		t.Fatalf("Update returned %T, want a spinnerModel[int]", reported)
	}

	if len(running.lines) != 1 || running.lines[0] != "✓ first" {
		t.Errorf("lines = %q, want %q", running.lines, []string{"✓ first"})
	}

	if !strings.Contains(running.View().Content, "✓ first") {
		t.Errorf("View() = %q, want the finished line in it", running.View().Content)
	}

	updated, cmd := running.Update(msg)

	finished, ok := updated.(spinnerModel[int])
	if !ok {
		t.Fatalf("Update returned %T, want a spinnerModel[int]", updated)
	}

	if cmd == nil {
		t.Fatal("Update returned no command, want tea.Quit")
	}

	if _, quitting := cmd().(tea.QuitMsg); !quitting {
		t.Errorf("Update's command produced %T, want tea.QuitMsg", cmd())
	}

	if !errors.Is(finished.err, failure) || finished.value != 7 {
		t.Errorf("the model kept %d and %v, want the job's own", finished.value, finished.err)
	}

	if content := finished.View().Content; content != "" {
		t.Errorf("the last frame = %q, want nothing left on screen", content)
	}
}

// A tick is the other message the model sees, and it must keep the program running.
func TestSpinnerModelKeepsSpinningOnATick(t *testing.T) {
	model := newSpinnerModel(context.Background(), "Working…", &writingProgress{w: new(bytes.Buffer)},
		func(context.Context, Progress) (int, error) { return 0, nil })

	updated, cmd := model.Update(model.spinner.Tick())
	if cmd == nil {
		t.Fatal("Update on a tick returned no command, want the next frame")
	}

	if _, quitting := cmd().(tea.QuitMsg); quitting {
		t.Error("Update on a tick quit the program, want it still spinning")
	}

	if spinning, ok := updated.(spinnerModel[int]); !ok || spinning.done {
		t.Errorf("Update on a tick = %#v, want a spinner that is not done", updated)
	}
}

// errCancelledForTest stands in for a command's own cancelled error, which is the caller's to name.
var errCancelledForTest = errors.New("cancelled")
