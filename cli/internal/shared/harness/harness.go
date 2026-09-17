// Package harness is the one definition of the coding harnesses codefall can set up: their names,
// and where each one reads the skills the extension installs. Both components need it — initcmd
// installs for the harnesses a project chose, and doctor reports on what it finds there — so it
// lives here rather than in either one's domain (ADR-003).
//
// It is a pure shared module: it imports the standard library and samber/mo, which is what lets
// domain/ and application/ name it. Adding a dependency here breaks that permission and fails the
// pure-shared-modules rule in .golangci.yml.
//
// What belongs here is what a harness is. How a hook definition reaches one — which file in the
// embedded tree it comes from, and whether that file is merged or copied — stays initcmd's, because
// registering it is what init does.
package harness

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/samber/mo"
)

// The harnesses codefall can set up.
const (
	Antigravity = "antigravity"
	ClaudeCode  = "claude-code"
	Codex       = "codex"
	Muse        = "muse"
	OpenCode    = "opencode"
)

// The two skills directories, relative to the directory init installs in.
const (
	claudeDir = ".claude"
	agentsDir = ".agents"
)

// SharedHooksPath is where the extension tree keeps the scripts every harness's hooks run, relative
// to the tree's own root. It is copied into each harness's skills directory, which is what makes the
// copy evidence that codefall's extension is installed there: nothing but codefall writes it, where
// the skills directory itself exists in plenty of projects that have never run codefall.
const SharedHooksPath = "hooks/shared"

// skillsDirs is where each harness reads the skills the extension step installs. Claude Code reads
// its own .claude/; a harness that follows the .agents/skills convention takes .agents/, and adding
// one of those is one row.
//
// The map is also the roster: a harness codefall can set up is one that has somewhere to install, so
// All and Parse read their answer from these keys rather than from a second list that could disagree
// with them.
var skillsDirs = map[string]string{
	Antigravity: agentsDir,
	ClaudeCode:  claudeDir,
	Codex:       agentsDir,
	Muse:        agentsDir,
	OpenCode:    agentsDir,
}

// All returns the harness names codefall can set up, sorted. Each call builds its own slice, so a
// caller that sorts or truncates what it gets back changes nothing for the next one.
func All() []string {
	return slices.Sorted(maps.Keys(skillsDirs))
}

// Parse returns the harness name when codefall can set it up, and an error naming the ones it can
// when it cannot. A harness codefall does not support yet is not a typo, so the message says so
// rather than calling the value unknown.
func Parse(name string) (string, error) {
	if _, known := skillsDirs[name]; known {
		return name, nil
	}

	return "", fmt.Errorf("harness %q is not supported yet (supported: %s)", name, strings.Join(All(), ", "))
}

// SkillsDir returns where the named harness reads skills, relative to the directory init installs
// in, or None when codefall cannot set that harness up. There is nothing useful to say about a name
// Parse has already refused, so this is an Option and not an error (ADR-GO-03).
func SkillsDir(name string) mo.Option[string] {
	dir, known := skillsDirs[name]
	if !known {
		return mo.None[string]()
	}

	return mo.Some(dir)
}

// SharedHooksDir returns where the named harness reads the hook scripts codefall installs, relative
// to the directory init installed in, or None when codefall cannot set that harness up. It is what a
// reader stats to tell whether codefall's extension is there.
func SharedHooksDir(name string) mo.Option[string] {
	dir, known := skillsDirs[name]
	if !known {
		return mo.None[string]()
	}

	return mo.Some(path.Join(dir, SharedHooksPath))
}
