// Package text holds the string helpers more than one component needs. They are the kind of thing
// every component would have written the same way, which is what makes them shared rather than
// anybody's (ADR-003).
//
// It is a pure shared module: it imports the standard library and nothing else, which is what lets
// domain/ and application/ name it. Adding a dependency here breaks that permission and fails the
// pure-shared-modules rule in .golangci.yml.
package text

import "strings"

// FirstLine is how a tool's chatty output becomes one line of a report or an error: the first line,
// trimmed. Output that starts with a blank line yields an empty string, which callers read as
// "nothing worth printing" and replace with something of their own.
func FirstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")

	return strings.TrimSpace(line)
}
