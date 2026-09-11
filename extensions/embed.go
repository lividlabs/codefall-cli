// Package extensions embeds the plugin tree into the binary, so `init` copies it straight from
// read-only data: no pin, no tarball, no network. It is the one deliverable the repo root owns.
package extensions

import (
	"embed"
	"io/fs"
)

//go:embed skills hooks shared README.md AGENTS.md docs
var tree embed.FS

// Files returns the embedded extension tree as a filesystem; callers walk it or copy it whole.
func Files() fs.FS {
	return tree
}
