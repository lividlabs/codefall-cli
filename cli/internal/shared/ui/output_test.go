package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestWriteLine(t *testing.T) {
	var out bytes.Buffer

	if err := WriteLine(&out, "one"); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}

	if err := WriteLine(&out, "two"); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}

	if got := out.String(); got != "one\ntwo\n" {
		t.Errorf("output = %q, want %q", got, "one\ntwo\n")
	}
}

func TestWriteLineReportsAFailedWrite(t *testing.T) {
	failure := errors.New("broken pipe")

	err := WriteLine(failingWriter{err: failure}, "one")
	if !errors.Is(err, failure) {
		t.Fatalf("WriteLine = %v, want it to wrap %v", err, failure)
	}

	if !strings.HasPrefix(err.Error(), "write report: ") {
		t.Errorf("WriteLine = %q, want it to say what could not be written", err)
	}
}

// The colour-profile writer is what makes a styled line safe to send anywhere: on a non-terminal
// stdout, and under NO_COLOR, the escapes are stripped and the words survive.
func TestNewWriterStripsStylingForANonTerminal(t *testing.T) {
	t.Setenv("CLICOLOR_FORCE", "0")
	t.Setenv("TTY_FORCE", "0")

	var out bytes.Buffer

	if err := WriteLine(NewWriter(&out), Style(TonePrimary).Render("✓")+" done"); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}

	if got := out.String(); got != "✓ done\n" {
		t.Errorf("output = %q, want the plain %q", got, "✓ done\n")
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }
