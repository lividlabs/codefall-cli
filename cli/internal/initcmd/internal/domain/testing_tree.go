package domain

import _ "embed"

// TestingAgents is the skeleton AGENTS.md init writes at the testing root: the headings for the
// facts only the project can state — its runners and their commands, the setup and state-forcing
// commands, how a run cleans up what it created, and what the environment needs — each with the one
// line that says what belongs under it.
//
// It is a skeleton and not a template: written once when the file is missing, and never afterwards.
// From the moment it exists it is the project's document, and `codefall-equip` is what fills the
// runners in (ADR-007).
//
//go:embed testing_agents.md
var TestingAgents string

// TestingReadme is the skeleton README.md init writes beside it: what the tree is, in a few lines,
// and the empty case index a project fills as it writes cases. Written under the same rule.
//
//go:embed testing_readme.md
var TestingReadme string
