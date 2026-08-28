// Package presentation builds setup's command. It is thin: it collects answers — from flags, from a
// survey, or from gh — hands them to the use case as a contract, and prints what each step did.
package presentation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/samber/mo"
	"github.com/spf13/cobra"

	"github.com/lividlabs/codefall-cli/internal/setup/internal/application"
	"github.com/lividlabs/codefall-cli/internal/setup/internal/domain"
)

// InitializeUseCase is what the command needs from the application layer, declared by its consumer.
// The two questions beside Run are what the command asks before it prompts: whether prompting is
// worth doing at all, and what to offer as the answer to the one question a tool can answer itself.
type InitializeUseCase interface {
	Run(ctx context.Context, request application.Request, observer application.Observer) (domain.Report, error)
	SettingsExist(dir string) (bool, error)
	SuggestGitHubRepo(ctx context.Context, dir string) mo.Option[string]
}

// The trackers the survey shows but does not accept. Huh has no disabled option, so they are offered
// with the state in their label and refused by the field's own validation — which is more use to a
// reader than leaving them out, because "not yet" is the answer they are looking for.
const (
	trackerJira   = "jira"
	trackerLinear = "linear"
)

// nextStep is the line that closes a successful run. Init writes what doctor checks, so doctor is
// what to run next.
const nextStep = "Next: codefall doctor"

// errCancelled is what both places a run can be interrupted report. Huh says the user aborted the
// form and Bubble Tea says the program was interrupted, but Ctrl-C during a question and Ctrl-C
// during the spinner are one event to the person who pressed it. It is returned as it is rather
// than wrapped, so Fang renders that sentence and not a prefix in front of it.
var errCancelled = errors.New("init cancelled")

// NewInitCommand builds `codefall init`.
func NewInitCommand(initialize InitializeUseCase) *cobra.Command {
	flags := &initFlags{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set this directory up for codefall",
		Long: "Creates .codefall/settings.json from your answers. Every question is also a flag, so " +
			"a scripted run passes them and is never prompted.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd, initialize, flags)
		},
	}

	flags.register(cmd)

	return cmd
}

// initFlags is every value the survey asks for, plus the two that change what a run does. Each
// prompted value is a flag, which is what keeps the command usable from a script (ADR-002).
type initFlags struct {
	tracker       string
	githubRepo    string
	githubProject int
	harness       string
	force         bool
}

func (f *initFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.tracker, "tracker", "",
		"issue tracker to use ("+strings.Join(domain.Trackers(), ", ")+")")
	cmd.Flags().StringVar(&f.githubRepo, "github-repo", "",
		"GitHub repository as owner/name (required when the tracker is "+domain.TrackerGitHub+")")
	cmd.Flags().IntVar(&f.githubProject, "github-project", 0,
		"GitHub Project number (optional)")
	cmd.Flags().StringVar(&f.harness, "harness", domain.HarnessClaudeCode,
		"coding harness to set up ("+strings.Join(domain.Harnesses(), ", ")+")")
	cmd.Flags().BoolVar(&f.force, "force", false,
		"rewrite .codefall/settings.json if it is already there")
}

// runInit is the command's body: collect the answers, run the steps, say what happened.
func runInit(cmd *cobra.Command, initialize InitializeUseCase, flags *initFlags) error {
	// The working directory is the one piece of environment the command reads; the use case
	// receives it as a value.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	// A bad flag or an unanswerable question is the user's own message; it is returned as it is so
	// Fang renders that sentence and not a prefix in front of it.
	request, err := buildRequest(cmd, initialize, flags, dir)
	if err != nil {
		return err
	}

	// One colorprofile writer for the whole run; lipgloss.Fprint* would build one per line.
	out := colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ())

	if _, err := runInitialize(cmd.Context(), initialize, request, out); err != nil {
		if errors.Is(err, errCancelled) {
			return err
		}

		return fmt.Errorf("init: %w", err)
	}

	return writeLine(out, faintStyle().Render(nextStep))
}

// buildRequest turns the flags into the use case's contract, asking for whatever they left out.
func buildRequest(
	cmd *cobra.Command, initialize InitializeUseCase, flags *initFlags, dir string,
) (application.Request, error) {
	harness, err := domain.ParseHarness(flags.harness)
	if err != nil {
		return application.Request{}, err
	}

	request := application.Request{Dir: dir, Harness: harness, Force: flags.force}

	if flags.tracker != "" {
		tracker, err := domain.ParseTracker(flags.tracker)
		if err != nil {
			return application.Request{}, err
		}

		request.Tracker = tracker
	}

	if flags.githubRepo != "" {
		if err := domain.ValidateRepo(flags.githubRepo); err != nil {
			return application.Request{}, err
		}

		request.GitHubRepo = mo.Some(flags.githubRepo)
	}

	// A project number is optional, so "not given" and "given as zero" are different answers and
	// only the flag's own record of being set can tell them apart.
	if cmd.Flags().Changed("github-project") {
		if flags.githubProject < 1 {
			return application.Request{}, fmt.Errorf(
				"--github-project %d must be a positive integer", flags.githubProject)
		}

		request.GitHubProject = mo.Some(flags.githubProject)
	}

	// Settings that are already there and are not being rewritten are not worth surveying for: the
	// step will skip whatever the answers are.
	settled, err := initialize.SettingsExist(dir)
	if err != nil {
		return application.Request{}, fmt.Errorf("init: %w", err)
	}

	if !settled || request.Force {
		request, err = collect(cmd.Context(), initialize, request)
		if err != nil {
			return application.Request{}, err
		}
	}

	// The tracker is settled by now, whether a flag or the survey chose it, so this is the last
	// place the two answers can be held against each other.
	if err := rejectGitHubFlags(cmd, request.Tracker); err != nil {
		return application.Request{}, err
	}

	return request, nil
}

// rejectGitHubFlags refuses the GitHub flags on a tracker that does not use them, in the terms the
// person typed. The domain refuses the same combination as an invariant, but it does so two layers
// away and in terms of settings, in a sentence that names neither flag; this is the sentence that
// does.
func rejectGitHubFlags(cmd *cobra.Command, tracker string) error {
	if tracker == "" || tracker == domain.TrackerGitHub {
		return nil
	}

	for _, name := range []string{"github-repo", "github-project"} {
		if cmd.Flags().Changed(name) {
			return fmt.Errorf("--%s is only used with --tracker %s", name, domain.TrackerGitHub)
		}
	}

	return nil
}

// collect fills in the answers the flags did not supply, by asking a person when there is one and by
// failing with the flag's name when there is not (ADR-002).
func collect(
	ctx context.Context, initialize InitializeUseCase, request application.Request,
) (application.Request, error) {
	if !needsAnswers(request) {
		return request, nil
	}

	// gh knows what repository this directory belongs to, which makes it an answer rather than a
	// question: it pre-fills the field when someone is answering, and stands in for the flag when
	// nobody is.
	suggestion := mo.None[string]()
	if mightUseGitHub(request) {
		suggestion = initialize.SuggestGitHubRepo(ctx, request.Dir)
	}

	if !stdinIsTerminal() {
		return withoutPrompting(request, suggestion)
	}

	return survey(ctx, request, suggestion)
}

// needsAnswers reports whether anything the settings cannot be built without is still missing. The
// project number is not one of those, so it is never on its own a reason to prompt.
func needsAnswers(request application.Request) bool {
	return request.Tracker == "" ||
		(request.Tracker == domain.TrackerGitHub && request.GitHubRepo.IsAbsent())
}

// mightUseGitHub reports whether a repository could still be wanted — either because the tracker is
// GitHub, or because it has not been chosen yet and might be.
func mightUseGitHub(request application.Request) bool {
	return request.GitHubRepo.IsAbsent() &&
		(request.Tracker == "" || request.Tracker == domain.TrackerGitHub)
}

// withoutPrompting is the path a script, CI, or an agent takes: what is known is used, and what is
// missing is an error naming the flag that would have supplied it. Nothing waits on input that
// cannot arrive.
func withoutPrompting(
	request application.Request, suggestion mo.Option[string],
) (application.Request, error) {
	if request.Tracker == "" {
		return application.Request{}, missingFlag("--tracker")
	}

	if request.Tracker != domain.TrackerGitHub || request.GitHubRepo.IsPresent() {
		return request, nil
	}

	repo, ok := suggestion.Get()
	if !ok {
		return application.Request{}, missingFlag("--github-repo")
	}

	request.GitHubRepo = mo.Some(repo)

	return request, nil
}

func missingFlag(flag string) error {
	return fmt.Errorf("missing %s (stdin is not a terminal)", flag)
}

// survey asks for the answers that are still missing and nothing else: a question whose flag was
// given is not asked again.
func survey(
	ctx context.Context, request application.Request, suggestion mo.Option[string],
) (application.Request, error) {
	// GitHub Issues is the first option and the starting value, so the highlighted answer is the
	// one most projects want.
	tracker := domain.TrackerGitHub
	if request.Tracker != "" {
		tracker = request.Tracker
	}

	repo := request.GitHubRepo.OrElse(suggestion.OrEmpty())

	project := ""
	if number, ok := request.GitHubProject.Get(); ok {
		project = strconv.Itoa(number)
	}

	var groups []*huh.Group

	if request.Tracker == "" {
		groups = append(groups, huh.NewGroup(trackerField(&tracker)))
	}

	// The GitHub questions are a group of their own so they can disappear: which tracker is chosen
	// may only be known once the form is running.
	if fields := gitHubFields(request, &repo, &project); len(fields) > 0 {
		groups = append(groups, huh.NewGroup(fields...).
			WithHideFunc(func() bool { return tracker != domain.TrackerGitHub }))
	}

	if err := runForm(ctx, groups); err != nil {
		return application.Request{}, err
	}

	return answered(request, tracker, repo, project)
}

func gitHubFields(request application.Request, repo, project *string) []huh.Field {
	var fields []huh.Field

	if request.GitHubRepo.IsAbsent() {
		fields = append(fields, repoField(repo))
	}

	if request.GitHubProject.IsAbsent() {
		fields = append(fields, projectField(project))
	}

	return fields
}

// runForm runs the survey. A form is a terminal program, so a cancelled command cancels it
// (ADR-002) and ACCESSIBLE chooses the screen-reader mode, as Huh's own examples do.
//
// There is always at least one group: survey is reached only when an answer is missing, and each
// answer that could be missing adds a group of its own.
func runForm(ctx context.Context, groups []*huh.Group) error {
	err := huh.NewForm(groups...).
		WithAccessible(os.Getenv("ACCESSIBLE") != "").
		RunWithContext(ctx)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, huh.ErrUserAborted):
		return errCancelled
	default:
		return fmt.Errorf("init: %w", err)
	}
}

// answered folds the survey's strings back into the contract. Everything here has already been
// validated by the field that collected it.
func answered(
	request application.Request, tracker, repo, project string,
) (application.Request, error) {
	request.Tracker = tracker

	if tracker != domain.TrackerGitHub {
		return request, nil
	}

	request.GitHubRepo = mo.Some(strings.TrimSpace(repo))

	trimmed := strings.TrimSpace(project)
	if trimmed == "" {
		return request, nil
	}

	number, err := strconv.Atoi(trimmed)
	if err != nil {
		return application.Request{}, fmt.Errorf("project number %q: %w", project, err)
	}

	request.GitHubProject = mo.Some(number)

	return request, nil
}

func trackerField(tracker *string) huh.Field {
	return huh.NewSelect[string]().
		Title("Which issue tracker should codefall use?").
		Options(
			huh.NewOption("GitHub Issues", domain.TrackerGitHub),
			huh.NewOption("Beads", domain.TrackerBeads),
			huh.NewOption("Jira (not yet available)", trackerJira),
			huh.NewOption("Linear (not yet available)", trackerLinear),
		).
		Value(tracker).
		Validate(availableTracker)
}

// availableTracker is what makes the last two options unchoosable. Huh v2 has no disabled option, so
// the refusal happens on submit, and its message says what to choose instead.
//
// The name is not the first word of the message because an error that opens with a capital fails the
// linter, and "Jira" is a name rather than a sentence starting in the wrong case.
func availableTracker(tracker string) error {
	switch tracker {
	case trackerJira:
		return notAvailable("Jira")
	case trackerLinear:
		return notAvailable("Linear")
	default:
		_, err := domain.ParseTracker(tracker)

		return err
	}
}

func notAvailable(tracker string) error {
	return fmt.Errorf("codefall cannot use %s yet — choose GitHub Issues or Beads", tracker)
}

func repoField(repo *string) huh.Field {
	return huh.NewInput().
		Title("Which GitHub repository holds the issues?").
		Placeholder("owner/name").
		Value(repo).
		Validate(func(value string) error { return domain.ValidateRepo(strings.TrimSpace(value)) })
}

func projectField(project *string) huh.Field {
	return huh.NewInput().
		Title("Which GitHub Project number? Leave it blank for none.").
		Placeholder("3").
		Value(project).
		Validate(validateProject)
}

// validateProject accepts a blank answer, because a project is optional and blank is how a form says
// none.
func validateProject(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	number, err := strconv.Atoi(trimmed)
	if err != nil || number < 1 {
		return fmt.Errorf("a project number is a positive integer, not %q", value)
	}

	return nil
}

// stdinIsTerminal reports whether there is someone to answer a question. It is a variable so the
// tests can take the path a script takes; nothing else reassigns it.
var stdinIsTerminal = func() bool {
	return term.IsTerminal(os.Stdin.Fd())
}

// stdoutIsTerminal decides both questions that depend on who is reading: whether the terminal can be
// asked about its background, and whether a spinner has anywhere to run. A pipe, a file, CI, and the
// tests all answer no.
func stdoutIsTerminal() bool {
	return term.IsTerminal(os.Stdout.Fd())
}

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

// The palette is doctor's: the same teal for something that was done and the same amber for
// something that was not, so the two commands read as one tool. Only the mark is coloured — the rest
// of a line is a path or a reason, where colour would be decoration rather than information.
//
// The scheme is built once, on first use, because deciding it means asking the terminal a question.
var outcomeStyles = sync.OnceValue(func() map[domain.Outcome]lipgloss.Style {
	c := lipgloss.LightDark(hasDarkBackground())

	return map[domain.Outcome]lipgloss.Style{
		domain.OutcomeDone: lipgloss.NewStyle().Bold(true).
			Foreground(c(lipgloss.Color("#2F6B6B"), lipgloss.Color("#6AB3B3"))),
		domain.OutcomeSkipped: lipgloss.NewStyle().Bold(true).
			Foreground(c(lipgloss.Color("#7E6217"), lipgloss.Color("#D9B44A"))),
	}
})

// Secondary text is dimmed and tinted teal-grey, as it is in doctor: here it is the closing line
// that says what to run next, and the label under the spinner.
var faintStyle = sync.OnceValue(func() lipgloss.Style {
	c := lipgloss.LightDark(hasDarkBackground())

	return lipgloss.NewStyle().Faint(true).Foreground(c(lipgloss.Color("#526D71"), lipgloss.Color("#8AA3A8")))
})

// The mark each outcome prints. A skipped step is a dash rather than doctor's warning glyph: nothing
// is wrong, the work was simply already done.
var outcomeMarks = map[domain.Outcome]string{
	domain.OutcomeDone:    "✓",
	domain.OutcomeSkipped: "-",
}

// stepLine is one finished step: its mark, then the sentence the step wrote about itself.
func stepLine(result domain.StepResult) string {
	return outcomeStyle(result.Outcome).Render(glyph(result.Outcome)) + " " + result.Detail
}

func glyph(outcome domain.Outcome) string {
	if mark, ok := outcomeMarks[outcome]; ok {
		return mark
	}

	return "?"
}

// outcomeStyle is the one place an outcome becomes a colour. An unknown outcome is left unstyled.
func outcomeStyle(outcome domain.Outcome) lipgloss.Style {
	if style, ok := outcomeStyles()[outcome]; ok {
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
