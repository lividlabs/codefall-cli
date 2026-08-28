package application

import (
	"context"
	"strings"

	"github.com/samber/mo"
)

// SuggestGitHubRepo asks gh which repository dir belongs to, so the survey can offer an answer
// rather than an empty field and a non-interactive run has one less flag to pass.
//
// Every way of not knowing means the same thing to the caller — there is nothing to suggest — so a
// missing gh, a directory that is not a GitHub repository, and gh failing for its own reasons all
// return None rather than an error (ADR-GO-03).
func (i *Initialize) SuggestGitHubRepo(ctx context.Context, dir string) mo.Option[string] {
	if i.runner.LookPath("gh").IsAbsent() {
		return mo.None[string]()
	}

	result, err := i.runner.Run(ctx, dir, "gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	if err != nil || result.ExitCode != 0 {
		return mo.None[string]()
	}

	repo := strings.TrimSpace(result.Stdout)
	if repo == "" {
		return mo.None[string]()
	}

	return mo.Some(repo)
}
