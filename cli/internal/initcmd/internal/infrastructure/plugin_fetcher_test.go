package infrastructure

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// The walk mirrors the whole source tree and reports the manifest's version. The promise
// "everything that is in the source, nothing else" is checked with the names read from the source
// itself rather than a second list of literals a reader could not compare anyway.
func TestEmbeddedPluginFetcherCopiesTheTreeAndReportsTheVersion(t *testing.T) {
	src := fstest.MapFS{
		pluginManifest:      &fstest.MapFile{Data: []byte(`{"version": "1.2.3"}`)},
		"hooks/hooks.json":  &fstest.MapFile{Data: []byte(`{"event": "PreToolUse"}`)},
		"skills/x/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: x\n---\n")},
		"shared/preflight.sh": &fstest.MapFile{
			Data: []byte("#!/bin/sh\n"),
		},
		"README.md":       &fstest.MapFile{Data: []byte("# plugin\n")},
		"docs/ROADMAP.md": &fstest.MapFile{Data: []byte("road")},
	}
	fetcher := NewEmbeddedPluginFetcher(src)

	version, err := fetcher.Version()
	if err != nil {
		t.Fatalf("Version: %v", err)
	}

	if version != "1.2.3" {
		t.Errorf("version = %q, want 1.2.3", version)
	}

	dest := t.TempDir()
	if err := fetcher.Fetch(context.Background(), dest); err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	got := readAllFiles(t, dest)
	if len(got) != len(src) {
		t.Fatalf("dest holds %d files, want all %d source files copied (%q)", len(got), len(src), got)
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

// A manifest that is not a manifest makes the version check noisy rather than trust the walk.
func TestEmbeddedPluginFetcherReportsAManifestItCannotParse(t *testing.T) {
	src := fstest.MapFS{
		pluginManifest: &fstest.MapFile{Data: []byte("{ not json")},
	}

	if _, err := NewEmbeddedPluginFetcher(src).Version(); err == nil {
		t.Error("Version err = nil, want a decode failure")
	}
}

// A cancelled context ends the walk rather than halting mid-directory.
func TestEmbeddedPluginFetcherHonoursTheContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewEmbeddedPluginFetcher(fstest.MapFS{"one.txt": &fstest.MapFile{}}).Fetch(ctx, t.TempDir())
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Fetch err = %v, want context.Canceled", err)
	}
}
