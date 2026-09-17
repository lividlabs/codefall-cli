// Package manifest is the one definition of the .codefall/manifest.json format: what finished init
// runs recorded, one entry per harness. Both components need it — initcmd writes the file and doctor
// reports on what it says — so it lives here rather than in either one's domain (ADR-003).
//
// It is a pure shared module: it imports the standard library and nothing else, which is what lets
// domain/ and application/ name it. Adding a dependency here breaks that permission and fails the
// pure-shared-modules rule in .golangci.yml.
//
// What belongs here is the format. What a run installs, and what to do about an entry the settings no
// longer name, belong to the component that decides it.
package manifest

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Name is where the record lives, in .codefall/ beside settings.json. It is written once every step
// of a run has succeeded, so what it says is an install that finished.
const Name = ".codefall/manifest.json"

// Document is a decoded manifest: one entry per harness a finished run installed for.
type Document struct {
	Harnesses map[string]Install `json:"harnesses"`
}

// Install is one harness's entry: the binary that wrote its files, and the files, relative to the
// directory init installed in.
type Install struct {
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

// Decode reads a manifest. A file written before harnesses were recorded per install names none of
// them and decodes cleanly to an empty record rather than an error: a reader then knows nothing about
// what is installed, which is the same position no file at all leaves it in.
func Decode(data []byte) (Document, error) {
	var document Document

	if err := json.Unmarshal(data, &document); err != nil {
		return Document{}, fmt.Errorf("decode %s: %w", Name, err)
	}

	return document, nil
}

// Encode renders a manifest as the bytes that go in the file, at the indent a person would have
// written by hand so a diff reads a line at a time.
func Encode(document Document) ([]byte, error) {
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", Name, err)
	}

	return data, nil
}

// Recorded returns the harnesses the manifest names, sorted. An entry carrying no version records
// nothing a reader can act on — which is what a hand-edited file looks like — so it is left out.
func (d Document) Recorded() []string {
	recorded := make([]string, 0, len(d.Harnesses))

	for name, install := range d.Harnesses {
		if install.Version != "" {
			recorded = append(recorded, name)
		}
	}

	slices.Sort(recorded)

	return recorded
}

// Versions is the version each harness was installed at, for a caller holding them against a binary's
// own tag. An entry carrying no version is left out, for the reason Recorded leaves it out.
func (d Document) Versions() map[string]string {
	versions := make(map[string]string, len(d.Harnesses))

	for name, install := range d.Harnesses {
		if install.Version != "" {
			versions[name] = install.Version
		}
	}

	return versions
}
