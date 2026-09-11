package infrastructure

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// The walk mirrors the whole source tree. The promise "everything that is in the source, nothing
// else" is checked with the names read from the destination, never from a second literal list.
func TestEmbeddedPluginFetcherCopiesTheTree(t *testing.T) {
	src := fstest.MapFS{
		"hooks/hooks.json":  &fstest.MapFile{Data: []byte(`{"event": "PreToolUse"}`)},
		"skills/x/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: x\n---\n")},
		"shared/preflight.sh": &fstest.MapFile{
			Data: []byte("#!/bin/sh\n"),
		},
		"README.md":       &fstest.MapFile{Data: []byte("# plugin\n")},
		"docs/ROADMAP.md": &fstest.MapFile{Data: []byte("road")},
	}
	fetcher := NewEmbeddedPluginFetcher(src)
	dest := t.TempDir()

	installed, err := fetcher.Fetch(context.Background(), dest)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	got := readAllFiles(t, dest)
	if len(got) != len(src) {
		t.Fatalf("dest holds %d files, want all %d source files copied (%q)", len(got), len(src), got)
	}

	if len(installed) != len(got) {
		t.Fatalf("installed = %d paths, want one per copied file (%v)", len(installed), installed)
	}
}

// readAllFiles returns the relative name of every file under dir; for the copy promise above.
func readAllFiles(t *testing.T, dir string) []string {
	t.Helper()

	var names []string
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, rerr := filepath.Rel(dir, path)
			if rerr != nil {
				return rerr
			}

			names = append(names, rel)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk dest: %v", err)
	}

	return names
}

// A cancelled context ends the walk rather than halting mid-directory.
func TestEmbeddedPluginFetcherHonoursTheContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewEmbeddedPluginFetcher(fstest.MapFS{"one.txt": &fstest.MapFile{}}).Fetch(ctx, t.TempDir())
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Fetch err = %v, want context.Canceled", err)
	}
}
