package application

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"

	"github.com/lividlabs/codefall-cli/cli/internal/initcmd/internal/domain"
	"github.com/lividlabs/codefall-cli/cli/internal/shared/harness"
)

// extension is the second step of a run: it mirrors the embedded extension tree wherever the
// harnesses the project uses read skills, so it never touches a network or a foreign CLI. It also
// hands back every path it wrote, per harness, which the run records in the manifest once the last
// step has succeeded.
//
// Harnesses share skills directories — four of the five read `.agents/` — so the copy happens once
// per directory, and every harness that reads that directory records the same files. Which directory
// each one is belongs to the shared harness module rather than to this step: doctor reports on the
// same directories, so neither component can be the one that decides where they are (ADR-003).
func (i *Initialize) extension(
	ctx context.Context, request Request,
) (domain.StepResult, map[string][]string, error) {
	dests, err := skillsDirs(request)
	if err != nil {
		return domain.StepResult{}, nil, err
	}

	// What each directory received, so a directory two harnesses share is fetched once, and what
	// each harness installed, which is what the manifest records.
	copied := map[string][]string{}
	installed := map[string][]string{}

	for _, name := range chosen(request) {
		dest := dests[name]

		if _, done := copied[dest]; !done {
			files, err := i.fetchSkills(ctx, request.Dir, dest)
			if err != nil {
				return domain.StepResult{}, nil, err
			}

			copied[dest] = files
		}

		installed[name] = projectPaths(dest, copied[dest])
	}

	return domain.ExtensionStep.Done(fmt.Sprintf("installed codefall's skills into %s",
		directoryList(slices.Sorted(maps.Keys(copied))))), installed, nil
}

// skillsDirs is where each chosen harness reads skills, or the error naming a harness codefall
// cannot set up. Presentation refuses an unknown harness before the use case sees it, so that error
// is a path a run should never take.
func skillsDirs(request Request) (map[string]string, error) {
	dirs := map[string]string{}

	for _, name := range chosen(request) {
		dir, known := harness.SkillsDir(name).Get()
		if !known {
			return nil, fmt.Errorf("harness %q has no extension mechanism", name)
		}

		dirs[name] = dir
	}

	return dirs, nil
}

// projectPaths puts the skills directory back in front of what Fetch wrote. A recorded path has to
// say which directory it is in: four harnesses share one directory, and a bare "skills/…" would not
// say whether it landed under `.claude/` or `.agents/`.
func projectPaths(dest string, files []string) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, path.Join(dest, file))
	}

	return paths
}

// directoryList names the directories the step wrote to, as the end of the sentence it reports: one
// directory on its own, and several joined the way a person reads a list.
func directoryList(dirs []string) string {
	slashed := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		slashed = append(slashed, dir+"/")
	}

	return sentenceList(slashed)
}
