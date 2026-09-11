package domain

import (
	"fmt"
	"slices"
	"strings"
)

// The coding harnesses codefall can set up.
//
// This is initcmd's own, unlike the settings format: nothing else in the project has an opinion
// about which harnesses can be set up, because setting them up is what init does.
const (
	HarnessAntigravity = "antigravity"
	HarnessClaudeCode  = "claude-code"
	HarnessCodex       = "codex"
	HarnessMuse        = "muse"
	HarnessOpenCode    = "opencode"
)

// The list is sorted once, here.
var harnesses = slices.Sorted(slices.Values([]string{
	HarnessAntigravity,
	HarnessClaudeCode,
	HarnessCodex,
	HarnessMuse,
	HarnessOpenCode,
}))

// Harnesses returns the supported harness names, sorted.
func Harnesses() []string {
	return slices.Clone(harnesses)
}

// ParseHarness returns the harness name when codefall can set it up, and an error naming the ones it
// can when it cannot. A harness codefall does not support yet is not a typo, so the message says so
// rather than calling the value unknown.
func ParseHarness(name string) (string, error) {
	if slices.Contains(harnesses, name) {
		return name, nil
	}

	return "", fmt.Errorf("harness %q is not supported yet (supported: %s)", name, strings.Join(harnesses, ", "))
}
