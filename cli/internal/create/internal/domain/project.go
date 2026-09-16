// Package domain holds what a new project starts with: the files create writes, the names it gives
// the remote and the first commit, and the steps of a run. It does no IO.
package domain

import (
	"strings"

	"github.com/samber/mo"
)

// The files a new project starts with, the remote create adds, and the message of the commit that
// holds them.
const (
	ReadmeName           = "README.md"
	GitIgnoreName        = ".gitignore"
	RemoteName           = "origin"
	InitialCommitMessage = "Initial commit"
)

// GitIgnore is the .gitignore a new project starts with. It names no language or framework — create
// does not know what the project will be written in — only what every checkout collects regardless:
// operating-system litter, editor state, local environment files, and logs.
const GitIgnore = `# Operating systems
.DS_Store
Thumbs.db
Desktop.ini

# Editors
.idea/
.vscode/
*.swp
*~

# Local environment
.env
.env.*
!.env.example

# Logs
*.log
`

// Readme is the README.md a new project starts with: the title, and the description under it when
// there is one. A description that is only whitespace is no description.
func Readme(title string, description mo.Option[string]) string {
	body := "# " + title + "\n"

	if text := strings.TrimSpace(description.OrEmpty()); text != "" {
		body += "\n" + text + "\n"
	}

	return body
}
