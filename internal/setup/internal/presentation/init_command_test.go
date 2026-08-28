package presentation

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/application"
	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

type fakeInitialize struct {
	report     domain.Report
	err        error
	exists     bool
	existsErr  error
	suggestion mo.Option[string]

	got       application.Request
	ran       bool
	suggested bool
}

func (f *fakeInitialize) Run(
	_ context.Context, request application.Request, observer application.Observer,
) (domain.Report, error) {
	f.got, f.ran = request, true

	for _, result := range f.report.Results() {
		observer.StepStarted(result.Step)
		observer.StepFinished(result)
	}

	return f.report, f.err
}

func (f *fakeInitialize) SettingsExist(string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeInitialize) SuggestGitHubRepo(context.Context, string) mo.Option[string] {
	f.suggested = true

	return f.suggestion
}

func newFakeInitialize() *fakeInitialize {
	return &fakeInitialize{
		report:     domain.NewReport(domain.SettingsStep.Done("wrote .codefall/settings.json (tracker: beads)")),
		suggestion: mo.None[string](),
	}
}

// run executes the command against a fake use case and returns everything it wrote.
//
// stdin is forced to answer "not a terminal", which is the path a script, CI, and an agent take —
// and the only one a test can take, because a form needs a terminal (ADR-002).
//
// CLICOLOR_FORCE and TTY_FORCE are cleared so a developer's forced-colour environment cannot leak
// ANSI into the buffer: colorprofile honours both regardless of whether the writer is a terminal.
func run(t *testing.T, initialize InitializeUseCase, args ...string) (string, error) {
	t.Helper()
	t.Setenv("CLICOLOR_FORCE", "0")
	t.Setenv("TTY_FORCE", "0")

	interactive(t, false)

	var out bytes.Buffer

	cmd := NewInitCommand(initialize)
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	// Fang sets both on the root in production, because it renders errors and usage itself.
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()

	return out.String(), err
}

// interactive decides what the command believes about stdin for the length of one test.
func interactive(t *testing.T, yes bool) {
	t.Helper()

	previous := stdinIsTerminal
	stdinIsTerminal = func() bool { return yes }

	t.Cleanup(func() { stdinIsTerminal = previous })
}

func TestInitCommandPassesTheFlagsToTheUseCase(t *testing.T) {
	initialize := newFakeInitialize()

	if _, err := run(t, initialize,
		"--tracker", "github",
		"--github-repo", "lividlabs/codefall-cli",
		"--github-project", "3",
		"--harness", "claude-code",
		"--force",
	); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	want := application.Request{
		Dir:           dir,
		Tracker:       domain.TrackerGitHub,
		GitHubRepo:    mo.Some("lividlabs/codefall-cli"),
		GitHubProject: mo.Some(3),
		Harness:       domain.HarnessClaudeCode,
		Force:         true,
	}

	if initialize.got != want {
		t.Errorf("request = %+v, want %+v", initialize.got, want)
	}

	// Every answer was on the command line, so there was nothing to ask gh either.
	if initialize.suggested {
		t.Error("asked gh for a repository, want it not asked when --github-repo was given")
	}
}

func TestInitCommandDefaultsTheHarnessAndLeavesTheOptionalValuesAbsent(t *testing.T) {
	initialize := newFakeInitialize()

	if _, err := run(t, initialize, "--tracker", "beads"); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if initialize.got.Harness != domain.HarnessClaudeCode {
		t.Errorf("Harness = %q, want %q", initialize.got.Harness, domain.HarnessClaudeCode)
	}

	if initialize.got.GitHubRepo.IsPresent() || initialize.got.GitHubProject.IsPresent() {
		t.Errorf("request = %+v, want the GitHub values absent", initialize.got)
	}

	if initialize.got.Force {
		t.Error("Force = true, want false without the flag")
	}

	// Beads needs no repository, so gh is not asked about one.
	if initialize.suggested {
		t.Error("asked gh for a repository, want it not asked when the tracker is beads")
	}
}

func TestInitCommandRejects(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "an unknown tracker",
			args: []string{"--tracker", "jira"},
			want: `unknown tracker "jira" (known trackers: beads, github)`,
		},
		{
			name: "a harness codefall cannot set up yet",
			args: []string{"--tracker", "beads", "--harness", "codex"},
			want: `harness "codex" is not supported yet (supported: claude-code)`,
		},
		{
			name: "a repository that is not owner/name",
			args: []string{"--tracker", "github", "--github-repo", "codefall-cli"},
			want: "owner/name",
		},
		{
			name: "a project number below one",
			args: []string{"--tracker", "github", "--github-repo", "owner/name", "--github-project", "0"},
			want: "the --github-project flag must be a positive integer, not 0",
		},
		{
			name: "a repository on a tracker that has no use for one",
			args: []string{"--tracker", "beads", "--github-repo", "owner/name"},
			want: "the --github-repo flag is only used with --tracker github",
		},
		{
			name: "a project number on a tracker that has no use for one",
			args: []string{"--tracker", "beads", "--github-project", "3"},
			want: "the --github-project flag is only used with --tracker github",
		},
		{
			name: "an argument",
			args: []string{"somewhere"},
			want: "unknown command",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initialize := newFakeInitialize()

			out, err := run(t, initialize, tc.args...)
			if err == nil {
				t.Fatalf("Execute = nil error, want one\n%s", out)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}

			if initialize.ran {
				t.Error("the use case ran, want the command to stop at the flag")
			}
		})
	}
}

// Without a terminal the command never waits for an answer that cannot arrive: it names the flag
// that would have supplied it (ADR-002).
func TestInitCommandWithoutATerminalNamesTheMissingFlag(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no tracker",
			args: nil,
			want: "missing --tracker (stdin is not a terminal)",
		},
		{
			name: "github without a repository",
			args: []string{"--tracker", "github"},
			want: "missing --github-repo (stdin is not a terminal)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initialize := newFakeInitialize()

			out, err := run(t, initialize, tc.args...)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("Execute error = %v, want %q\n%s", err, tc.want, out)
			}

			if initialize.ran {
				t.Error("the use case ran, want the command to stop before it")
			}
		})
	}
}

// A repository gh already knows is an answer, not a question: without a terminal it stands in for
// the flag rather than failing the run.
func TestInitCommandUsesTheRepositoryGHSuggests(t *testing.T) {
	initialize := newFakeInitialize()
	initialize.suggestion = mo.Some("lividlabs/codefall-cli")

	if _, err := run(t, initialize, "--tracker", "github"); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got, ok := initialize.got.GitHubRepo.Get(); !ok || got != "lividlabs/codefall-cli" {
		t.Errorf("GitHubRepo = %v, want Some(%q)", initialize.got.GitHubRepo, "lividlabs/codefall-cli")
	}
}

// Settings that are already there and are not being rewritten are not worth asking about: the step
// will skip them whatever the answers are, so the command does not fail on a missing flag either.
func TestInitCommandAsksNothingWhenSettingsAlreadyExist(t *testing.T) {
	initialize := newFakeInitialize()
	initialize.exists = true
	initialize.report = domain.NewReport(
		domain.SettingsStep.Skipped(".codefall/settings.json already exists (use --force to rewrite it)"))

	out, err := run(t, initialize)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if initialize.got.Tracker != "" {
		t.Errorf("Tracker = %q, want it left empty", initialize.got.Tracker)
	}

	if initialize.suggested {
		t.Error("asked gh for a repository, want nothing asked when the settings are already there")
	}

	want := "- .codefall/settings.json already exists (use --force to rewrite it)\n" + nextStep + "\n"
	if out != want {
		t.Errorf("output =\n%q\nwant\n%q", out, want)
	}
}

// --force is what turns the questions back on, because the file it would skip is going to be
// rewritten.
func TestInitCommandStillAsksWhenForcingOverExistingSettings(t *testing.T) {
	initialize := newFakeInitialize()
	initialize.exists = true

	out, err := run(t, initialize, "--force")
	if err == nil || err.Error() != "missing --tracker (stdin is not a terminal)" {
		t.Fatalf("Execute error = %v, want the missing flag\n%s", err, out)
	}
}

func TestInitCommandPrintsALineForEachFinishedStepAndWhatToRunNext(t *testing.T) {
	initialize := newFakeInitialize()
	initialize.report = domain.NewReport(
		domain.SettingsStep.Done("wrote .codefall/settings.json (tracker: beads)"),
		domain.PluginStep.Skipped(
			"the codefall marketplace is declared and codefall@codefall enabled in .claude/settings.json"),
	)

	out, err := run(t, initialize, "--tracker", "beads")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "✓ wrote .codefall/settings.json (tracker: beads)\n" +
		"- the codefall marketplace is declared and codefall@codefall enabled in .claude/settings.json\n" +
		nextStep + "\n"
	if out != want {
		t.Errorf("output =\n%q\nwant\n%q", out, want)
	}
}

func TestInitCommandWrapsAUseCaseError(t *testing.T) {
	failure := errors.New("settings: no space left on device")

	initialize := newFakeInitialize()
	initialize.report = domain.NewReport()
	initialize.err = failure

	out, err := run(t, initialize, "--tracker", "beads")
	if !errors.Is(err, failure) {
		t.Fatalf("Execute error = %v, want it to wrap %v", err, failure)
	}

	if !strings.HasPrefix(err.Error(), "init: ") {
		t.Errorf("error = %q, want it to say which command failed", err)
	}

	if out != "" {
		t.Errorf("output = %q, want nothing printed", out)
	}
}

func TestInitCommandReportsSettingsItCannotRead(t *testing.T) {
	initialize := newFakeInitialize()
	initialize.existsErr = errors.New("permission denied")

	if _, err := run(t, initialize, "--tracker", "beads"); err == nil ||
		!strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Execute error = %v, want it to carry the read failure", err)
	}
}

// The marks are the only styled part of a line, and they degrade to plain text when nothing is
// reading them as a terminal.
func TestStepLine(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result domain.StepResult
		want   string
	}{
		{
			name:   "a step that did its work",
			result: domain.SettingsStep.Done("wrote .codefall/settings.json (tracker: beads)"),
			want:   "✓ wrote .codefall/settings.json (tracker: beads)",
		},
		{
			name:   "a step that found its work done",
			result: domain.SettingsStep.Skipped(".codefall/settings.json already exists"),
			want:   "- .codefall/settings.json already exists",
		},
		{
			name:   "an outcome from nowhere",
			result: domain.StepResult{Outcome: domain.Outcome(42), Detail: "nothing to say"},
			want:   "? nothing to say",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripANSI(stepLine(tc.result)); got != tc.want {
				t.Errorf("stepLine() = %q, want %q", got, tc.want)
			}

			if want := outcomeStyle(tc.result.Outcome).Render(glyph(tc.result.Outcome)); !strings.HasPrefix(
				stepLine(tc.result), want+" ") {
				t.Errorf("stepLine() = %q, want it to open with the styled mark %q", stepLine(tc.result), want)
			}
		})
	}
}

// Dimming is invisible to the command tests, because the colorprofile writer strips it before the
// buffer sees it. Assert it against the styled string.
func TestTheClosingLineIsFaint(t *testing.T) {
	const faintAttr = "\x1b[2"

	line := faintStyle().Render(nextStep)
	if !strings.Contains(line, faintAttr) || !strings.Contains(line, nextStep) {
		t.Errorf("the closing line = %q, want %q dimmed", line, nextStep)
	}
}

func TestValidateProject(t *testing.T) {
	for _, value := range []string{"", "   ", "1", "42"} {
		if err := validateProject(value); err != nil {
			t.Errorf("validateProject(%q) = %v, want nil", value, err)
		}
	}

	for _, value := range []string{"0", "-1", "three", "1.5"} {
		if err := validateProject(value); err == nil {
			t.Errorf("validateProject(%q) = nil, want an error", value)
		}
	}
}

// Huh v2 has no disabled option, so the two trackers that are only announced are refused here.
func TestAvailableTracker(t *testing.T) {
	for _, tracker := range domain.Trackers() {
		if err := availableTracker(tracker); err != nil {
			t.Errorf("availableTracker(%q) = %v, want nil", tracker, err)
		}
	}

	for _, tc := range []struct{ tracker, want string }{
		{trackerJira, "codefall cannot use Jira yet"},
		{trackerLinear, "codefall cannot use Linear yet"},
		{"nonsense", "unknown tracker"},
	} {
		err := availableTracker(tc.tracker)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("availableTracker(%q) = %v, want it to mention %q", tc.tracker, err, tc.want)
		}
	}
}

// The spinner needs a terminal, so what is testable here is the model around it: that the use case
// runs as its one command, that each step's news reaches the screen, and that the last frame is
// empty so the closing line starts on a clean line.
func TestInitSpinnerFollowsTheRunAndQuitsWithItsAnswer(t *testing.T) {
	initialize := newFakeInitialize()
	initialize.err = errors.New("settings: no space left on device")

	model := newInitSpinner(context.Background(), initialize, application.Request{Dir: "/somewhere"},
		&writingObserver{w: new(bytes.Buffer)})

	if !strings.Contains(model.View().Content, openingLabel) {
		t.Errorf("View() = %q, want it to show %q", model.View().Content, openingLabel)
	}

	msg, ok := model.run().(initializedMsg)
	if !ok {
		t.Fatalf("the spinner's command returned %T, want an initializedMsg", model.run())
	}

	if initialize.got.Dir != "/somewhere" {
		t.Errorf("ran in %q, want %q", initialize.got.Dir, "/somewhere")
	}

	// A step that starts renames the label; a step that finishes becomes a line the model keeps, so
	// the caller can write it again once the program has cleared its own frames.
	started, _ := model.Update(stepStartedMsg{step: domain.SettingsStep})

	running, ok := started.(initSpinner)
	if !ok {
		t.Fatalf("Update returned %T, want an initSpinner", started)
	}

	if want := domain.SettingsStep.Title + "…"; running.label != want {
		t.Errorf("label = %q, want %q", running.label, want)
	}

	result := domain.SettingsStep.Done("wrote .codefall/settings.json (tracker: beads)")

	reported, _ := running.Update(stepFinishedMsg{result: result})

	running, ok = reported.(initSpinner)
	if !ok {
		t.Fatalf("Update returned %T, want an initSpinner", reported)
	}

	if want := []string{stepLine(result)}; len(running.lines) != 1 || running.lines[0] != want[0] {
		t.Errorf("lines = %q, want %q", running.lines, want)
	}

	if !strings.Contains(stripANSI(running.View().Content), stripANSI(stepLine(result))) {
		t.Errorf("View() = %q, want the finished step in it", running.View().Content)
	}

	// The error travels in the message rather than failing the program: the command decides what an
	// error means, not the spinner.
	if !errors.Is(msg.err, initialize.err) {
		t.Errorf("message error = %v, want it to wrap %v", msg.err, initialize.err)
	}

	updated, cmd := running.Update(msg)

	finished, ok := updated.(initSpinner)
	if !ok {
		t.Fatalf("Update returned %T, want an initSpinner", updated)
	}

	if cmd == nil {
		t.Fatal("Update returned no command, want tea.Quit")
	}

	if _, quitting := cmd().(tea.QuitMsg); !quitting {
		t.Errorf("Update's command produced %T, want tea.QuitMsg", cmd())
	}

	if !errors.Is(finished.err, initialize.err) {
		t.Errorf("the model kept error %v, want the use case's own", finished.err)
	}

	if content := finished.View().Content; content != "" {
		t.Errorf("the last frame = %q, want nothing left on screen", content)
	}
}

// A tick is the other message the model sees, and it must keep the program running.
func TestInitSpinnerKeepsSpinningOnATick(t *testing.T) {
	model := newInitSpinner(context.Background(), newFakeInitialize(), application.Request{},
		&writingObserver{w: new(bytes.Buffer)})

	updated, cmd := model.Update(model.spinner.Tick())
	if cmd == nil {
		t.Fatal("Update on a tick returned no command, want the next frame")
	}

	if _, quitting := cmd().(tea.QuitMsg); quitting {
		t.Error("Update on a tick quit the program, want it still spinning")
	}

	if spinning, ok := updated.(initSpinner); !ok || spinning.done {
		t.Errorf("Update on a tick = %#v, want a spinner that is not done", updated)
	}
}

// Ctrl-C during the spinner and Ctrl-C during a question are one event to the person who pressed it,
// so they get one sentence — and it is returned as it is, for Fang to render without a prefix.
func TestSpinnerErrorSaysCancelledWhenTheProgramWasInterrupted(t *testing.T) {
	if got := spinnerError(tea.ErrInterrupted); !errors.Is(got, errCancelled) {
		t.Errorf("spinnerError(tea.ErrInterrupted) = %v, want %v", got, errCancelled)
	}

	failure := errors.New("no terminal")
	if got := spinnerError(failure); !errors.Is(got, failure) ||
		!strings.HasPrefix(got.Error(), "spinner: ") {
		t.Errorf("spinnerError(%v) = %v, want it wrapped as a spinner failure", failure, got)
	}
}

// A write that fails is the run's error, not a line quietly dropped.
func TestWritingObserverKeepsTheFirstWriteFailure(t *testing.T) {
	failure := errors.New("broken pipe")
	observer := &writingObserver{w: failingWriter{err: failure}}

	observer.StepFinished(domain.SettingsStep.Done("wrote it"))
	observer.StepFinished(domain.SettingsStep.Done("wrote it again"))

	if !errors.Is(observer.err, failure) {
		t.Errorf("err = %v, want it to wrap %v", observer.err, failure)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

// Under `go test` stdout is not a terminal, which is what puts every test above on the plain path:
// no spinner, and nobody to ask about the background. Dark is the answer to the second, which is
// what lipgloss falls back to as well.
func TestWithoutATerminalNothingSpinsAndTheBackgroundIsDark(t *testing.T) {
	if stdoutIsTerminal() {
		t.Error("stdoutIsTerminal() = true, want false under go test")
	}

	if !hasDarkBackground() {
		t.Error("hasDarkBackground() = false, want true when stdout is not a terminal")
	}
}

// stripANSI removes the escape sequences lipgloss renders, which the colorprofile writer would
// normally downsample away on its way to a non-terminal.
func stripANSI(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			for i < len(s) && s[i] != 'm' {
				i++
			}

			continue
		}

		b.WriteByte(s[i])
	}

	return b.String()
}
