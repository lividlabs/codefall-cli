package manifest

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

const twoHarnesses = `{
  "harnesses": {
    "codex": {"version": "v1.1.0", "files": [".agents/skills/design/SKILL.md"]},
    "claude-code": {"version": "v1.2.3", "files": [".claude/skills/design/SKILL.md"]}
  }
}`

func TestDecodeReadsEveryEntry(t *testing.T) {
	document, err := Decode([]byte(twoHarnesses))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got, want := document.Recorded(), []string{"claude-code", "codex"}; !slices.Equal(got, want) {
		t.Errorf("Recorded() = %q, want %q", got, want)
	}

	want := map[string]string{"claude-code": "v1.2.3", "codex": "v1.1.0"}
	if got := document.Versions(); !maps.Equal(got, want) {
		t.Errorf("Versions() = %v, want %v", got, want)
	}

	if got, want := document.Harnesses["codex"].Files,
		[]string{".agents/skills/design/SKILL.md"}; !slices.Equal(got, want) {
		t.Errorf("codex files = %q, want %q", got, want)
	}
}

// A file written before harnesses were recorded per install names none of them. It decodes rather
// than failing, and records nothing — the same position no file at all leaves a reader in.
func TestDecodeReadsAnOlderRecordAsEmpty(t *testing.T) {
	document, err := Decode([]byte(`{"harness": "codex", "version": "v1.2.3"}`))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got := document.Recorded(); len(got) != 0 {
		t.Errorf("Recorded() = %q, want none", got)
	}

	if got := document.Versions(); len(got) != 0 {
		t.Errorf("Versions() = %v, want none", got)
	}
}

// An entry with no version says an install was recorded without saying what installed it, which is
// nothing a reader can hold against a binary's tag or report as present.
func TestAnEntryWithNoVersionIsLeftOut(t *testing.T) {
	document, err := Decode([]byte(
		`{"harnesses": {"codex": {"files": ["x"]}, "muse": {"version": "v1.0.0"}}}`))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got, want := document.Recorded(), []string{"muse"}; !slices.Equal(got, want) {
		t.Errorf("Recorded() = %q, want %q", got, want)
	}

	if _, left := document.Versions()["codex"]; left {
		t.Errorf("Versions() = %v, want codex left out", document.Versions())
	}
}

func TestDecodeNamesTheFileItCouldNotRead(t *testing.T) {
	_, err := Decode([]byte("{ not json"))
	if err == nil || !strings.Contains(err.Error(), Name) {
		t.Errorf("Decode error = %v, want it to name %s", err, Name)
	}
}

// What Encode writes, Decode reads back: the two halves of the format are one definition, and a
// change to either that the other cannot follow fails here.
func TestEncodeAndDecodeAgree(t *testing.T) {
	document := Document{Harnesses: map[string]Install{
		"claude-code": {Version: "v1.2.3", Files: []string{".claude/skills/design/SKILL.md"}},
	}}

	data, err := Encode(document)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	read, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got, want := read.Versions()["claude-code"], "v1.2.3"; got != want {
		t.Errorf("version after a round trip = %q, want %q", got, want)
	}

	if !strings.Contains(string(data), "\n  \"harnesses\"") {
		t.Errorf("Encode wrote %s, want it indented the way a person would write it", data)
	}
}
