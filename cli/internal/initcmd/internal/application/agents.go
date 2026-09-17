package application

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/samber/mo"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/harness"
)

// The two files this step knows about, named as a person reads them. AGENTS.md holds the rules;
// CLAUDE.md is a pointer to it, because Claude Code reads CLAUDE.md and the rules are not kept in
// two places. The pointer is the whole file, one line and a newline.
const (
	agentsName       = "AGENTS.md"
	claudeMemoryName = "CLAUDE.md"
	claudePointer    = "See [AGENTS.md](AGENTS.md) — the rules for this repo live there, and there only.\n"
)

// agentsChange is what the step did to AGENTS.md, which is most of what it has to report.
type agentsChange int

const (
	agentsCurrent agentsChange = iota
	agentsCreated
	agentsAppended
	agentsReplaced
)

// agents is the last step of a run: it writes codefall's own Beads section into AGENTS.md, and gives
// Claude Code the CLAUDE.md pointer to it when the project has no CLAUDE.md at all.
//
// bd would have written a section of its own here, and init runs `bd init --skip-agents` because
// that text describes a setup codefall does not create. So the section is codefall's, and this is
// what writes it.
//
// It runs after the Beads step on purpose. bd init stages AGENTS.md and CLAUDE.md when they exist
// and commits what it staged under its own message, so an edit made before it ran would land in
// bd's commit rather than the author's.
func (i *Initialize) agents(_ context.Context, request Request) (domain.StepResult, error) {
	change, err := i.writeBeadsSection(request.Dir)
	if err != nil {
		return domain.StepResult{}, err
	}

	pointed, err := i.writeClaudePointer(request)
	if err != nil {
		return domain.StepResult{}, err
	}

	detail := agentsDetail(change, pointed)
	if detail == "" {
		return domain.AgentsStep.Skipped(
			fmt.Sprintf("the Beads section in %s is current", agentsName)), nil
	}

	return domain.AgentsStep.Done(detail), nil
}

// agentsDetail is what the step reports it did. An empty string is the skip: nothing was written,
// because the section was already what codefall would have written and CLAUDE.md was not this run's
// to make.
func agentsDetail(change agentsChange, pointed bool) string {
	var detail string

	switch change {
	case agentsCreated:
		detail = fmt.Sprintf("created %s with the Beads section", agentsName)
	case agentsAppended:
		detail = fmt.Sprintf("added the Beads section to %s", agentsName)
	case agentsReplaced:
		detail = fmt.Sprintf("updated the Beads section in %s", agentsName)
	case agentsCurrent:
	}

	if !pointed {
		return detail
	}

	created := "created " + claudeMemoryName

	if detail == "" {
		return created
	}

	return detail + " and " + created
}

// writeBeadsSection puts codefall's section in AGENTS.md and reports what that took.
//
// The file belongs to the project, so the step changes as little of it as it can: a section already
// marked is replaced between its markers and everything on either side survives byte for byte, and
// a file with no markers keeps what it says and gains the section at the end. A file that is already
// what would be written is not written at all, which is what makes a second run a skip.
func (i *Initialize) writeBeadsSection(dir string) (agentsChange, error) {
	body, err := i.readProjectFile(dir, agentsName)
	if err != nil {
		return agentsCurrent, err
	}

	existing, present := body.Get()

	next, change, err := beadsSectionIn(existing, present)
	if err != nil {
		return agentsCurrent, err
	}

	if change == agentsCurrent {
		return agentsCurrent, nil
	}

	if err := i.files.WriteFile(filepath.Join(dir, agentsName), []byte(next)); err != nil {
		return agentsCurrent, fmt.Errorf("write %s: %w", agentsName, err)
	}

	return change, nil
}

// beadsSectionIn is the string surgery: what AGENTS.md should say, given what it says now. The
// encoding of the section is here rather than in the domain, which holds the words and the markers
// and nothing about how they are spliced into somebody's file.
func beadsSectionIn(existing string, present bool) (string, agentsChange, error) {
	// The embedded file ends with a newline after its closing marker. What is spliced into a file
	// does not carry it, because where it lands decides what follows.
	section := strings.TrimSuffix(domain.BeadsSection, "\n")

	begin := strings.Index(existing, domain.BeadsSectionBegin)
	if begin >= 0 {
		// A file with an opening marker and no closing one cannot be edited without guessing where
		// codefall's words stop and the project's resume, and a guess here silently eats somebody's
		// prose.
		offset := strings.Index(existing[begin:], domain.BeadsSectionEnd)
		if offset < 0 {
			return "", agentsCurrent, fmt.Errorf("%s has %s with no %s after it",
				agentsName, domain.BeadsSectionBegin, domain.BeadsSectionEnd)
		}

		end := begin + offset + len(domain.BeadsSectionEnd)

		next := existing[:begin] + section + existing[end:]
		if next == existing {
			return existing, agentsCurrent, nil
		}

		return next, agentsReplaced, nil
	}

	// One blank line between what the file said and the section, however the file ended, and a
	// single newline at the end of it.
	next := section + "\n"
	if trimmed := strings.TrimRight(existing, "\n"); trimmed != "" {
		next = trimmed + "\n\n" + next
	}

	if !present {
		return next, agentsCreated, nil
	}

	return next, agentsAppended, nil
}

// writeClaudePointer creates CLAUDE.md for the harness that reads it, and reports whether it did.
//
// Only when the file is missing. A CLAUDE.md the project already has says whatever its author meant
// it to say, and replacing that with a pointer would throw the rules away rather than point at them.
func (i *Initialize) writeClaudePointer(request Request) (bool, error) {
	if !slices.Contains(chosen(request), harness.ClaudeCode) {
		return false, nil
	}

	body, err := i.readProjectFile(request.Dir, claudeMemoryName)
	if err != nil {
		return false, err
	}

	if body.IsPresent() {
		return false, nil
	}

	if err := i.files.WriteFile(
		filepath.Join(request.Dir, claudeMemoryName), []byte(claudePointer),
	); err != nil {
		return false, fmt.Errorf("write %s: %w", claudeMemoryName, err)
	}

	return true, nil
}

// readProjectFile reads one of the project's markdown files. A file that is not there is None rather
// than an error, because this step creates what it needs (ADR-GO-03); a file that cannot be read is
// an error naming it, because the step was about to write it.
//
// Unlike the harness's settings, an empty file is present and empty: markdown has no shape to be
// wrong, and a file somebody has made is a file the step appends to rather than creates.
func (i *Initialize) readProjectFile(dir, name string) (mo.Option[string], error) {
	data, err := i.files.ReadFile(filepath.Join(dir, name))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return mo.None[string](), nil
	case err != nil:
		return mo.None[string](), fmt.Errorf("read %s: %w", name, err)
	}

	return mo.Some(string(data)), nil
}
