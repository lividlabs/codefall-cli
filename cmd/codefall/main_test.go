package main

import (
	"bytes"
	"testing"
)

func TestRunPrintsWelcome(t *testing.T) {
	var out bytes.Buffer

	if err := run(nil, &out); err != nil {
		t.Fatalf("run: %v", err)
	}

	if got, want := out.String(), "Welcome to Codefall\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
