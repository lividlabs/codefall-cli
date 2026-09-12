package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/mo"
)

// rootExpression is how the Claude and Codex definitions name the repository root: both harnesses
// run a hook command through a shell, so the root is worked out when the hook runs rather than
// written down when init does. A subdirectory install puts its path relative to the root after it.
const rootExpression = "$(git rev-parse --show-toplevel)/"

// shellSpecial is every character that keeps its meaning inside the double quotes a definition
// wraps a script path in. A subdirectory whose name holds one would change what the command runs, so
// it is refused rather than escaped: the escape would differ per harness, and such a name is rare
// enough that asking for another directory costs less than getting the quoting wrong.
const shellSpecial = "\"$`\\"

// RepositoryRoot reports the root of the git work tree dir sits inside, when dir is below it. At the
// root itself there is nothing to choose between, and outside a work tree there is nothing to
// choose at all — preflight says why — so both are None (ADR-GO-03).
//
// One git call answers both halves: --show-toplevel prints the root and --show-prefix prints dir's
// path relative to it, an empty line at the root. The prefix is what decides, rather than comparing
// dir with the root, because the two can name one directory differently: /tmp and /private/tmp on
// macOS, or any path through a symlink.
func (i *Initialize) RepositoryRoot(ctx context.Context, dir string) mo.Option[string] {
	if i.runner.LookPath(gitCommand).IsAbsent() {
		return mo.None[string]()
	}

	result, err := i.runner.Run(ctx, dir, gitCommand, "rev-parse", "--show-toplevel", "--show-prefix")
	if err != nil || result.ExitCode != 0 {
		return mo.None[string]()
	}

	root, prefix, _ := strings.Cut(strings.TrimRight(result.Stdout, "\n"), "\n")
	if root == "" || prefix == "" {
		return mo.None[string]()
	}

	return mo.Some(root)
}

// repositoryPrefix is dir's path relative to the root of its work tree, with a trailing slash, or
// empty at the root. Preflight has already established that dir is inside a work tree, so a git that
// cannot answer is an error rather than a None.
func (i *Initialize) repositoryPrefix(ctx context.Context, dir string) (string, error) {
	result, err := i.probe(ctx, dir, gitCommand, "rev-parse", "--show-prefix")
	if err != nil {
		return "", err
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("git rev-parse --show-prefix exited %d: %s",
			result.ExitCode, strings.TrimSpace(result.Stderr))
	}

	prefix := strings.TrimSpace(result.Stdout)
	if strings.ContainsAny(prefix, shellSpecial) {
		return "", fmt.Errorf("%q holds a character a hook command cannot quote (%s); "+
			"install at the repository root or from another directory", prefix, shellSpecial)
	}

	return prefix, nil
}

// placeUnder rewrites every command in a decoded definition that names the repository root so it
// names prefix below it instead: the shared script landed in the install directory, not the root.
// An empty prefix changes nothing, so an install at the root registers exactly the command the
// definition ships and a rerun still recognises it.
func placeUnder(value any, prefix string) {
	if prefix == "" {
		return
	}

	switch shaped := value.(type) {
	case map[string]any:
		for key, child := range shaped {
			if command, ok := child.(string); ok && key == commandKey {
				shaped[key] = strings.ReplaceAll(command, rootExpression, rootExpression+prefix)
				continue
			}

			placeUnder(child, prefix)
		}
	case []any:
		for _, child := range shaped {
			placeUnder(child, prefix)
		}
	}
}
