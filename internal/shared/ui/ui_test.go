package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestMark(t *testing.T) {
	for _, tc := range []struct {
		name string
		tone Tone
		want string
	}{
		{name: "primary", tone: TonePrimary, want: "✓"},
		{name: "warn", tone: ToneWarn, want: "!"},
		{name: "fail", tone: ToneFail, want: "✗"},
		{name: "faint has no mark of its own", tone: ToneFaint, want: "?"},
		{name: "a tone from nowhere", tone: Tone(42), want: "?"},
		{name: "the zero value", tone: ToneNone, want: "?"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Mark(tc.tone); got != tc.want {
				t.Errorf("Mark(%v) = %q, want %q", tc.tone, got, tc.want)
			}
		})
	}
}

// Every tone that means something renders differently from the others, so a reader can tell them
// apart; the ones that mean nothing render nothing.
func TestStyle(t *testing.T) {
	rendered := map[Tone]string{}

	for _, tone := range []Tone{TonePrimary, ToneWarn, ToneFail, ToneFaint} {
		got := Style(tone).Render("x")
		if !strings.ContainsRune(got, 0x1b) {
			t.Errorf("Style(%v).Render(%q) = %q, want it styled", tone, "x", got)
		}

		for other, seen := range rendered {
			if seen == got {
				t.Errorf("Style(%v) renders as Style(%v) does (%q), want them distinguishable",
					tone, other, got)
			}
		}

		rendered[tone] = got
	}

	for _, tone := range []Tone{ToneNone, Tone(42)} {
		if got, want := Style(tone).Render("x"), lipgloss.NewStyle().Render("x"); got != want {
			t.Errorf("Style(%v).Render(%q) = %q, want the unstyled %q", tone, "x", got, want)
		}
	}
}

// Secondary text is dimmed. The exact escape is faint (2) plus a 24-bit foreground, e.g.
// "\x1b[2;38;2;138;163;168m" on dark; check for the faint attribute rather than an exact byte
// sequence so the test does not pin the colour.
func TestFaintIsDimmed(t *testing.T) {
	const faintAttr = "\x1b[2"

	if got := Style(ToneFaint).Render("versions"); !strings.Contains(got, faintAttr) ||
		!strings.Contains(got, "versions") {
		t.Errorf("Style(ToneFaint).Render(%q) = %q, want it dimmed", "versions", got)
	}

	if got := Style(TonePrimary).Render("done"); strings.Contains(got, faintAttr) {
		t.Errorf("Style(TonePrimary).Render(%q) = %q, want nothing dimmed", "done", got)
	}
}

// The heading is the one style that paints a background, and its padding is what makes it read as a
// chip rather than a word.
func TestHeadingStyleIsAPaddedChip(t *testing.T) {
	got := HeadingStyle().Render("DOCTOR SUMMARY")

	// The padding is what a non-terminal stdout is left with once colorprofile has stripped the
	// colour, so it is the part worth pinning: the word, a space on either side, and nothing else.
	if plain := stripANSI(got); plain != " DOCTOR SUMMARY " {
		t.Errorf("HeadingStyle().Render() = %q, plain %q, want %q", got, plain, " DOCTOR SUMMARY ")
	}

	if !strings.ContainsRune(got, 0x1b) {
		t.Errorf("HeadingStyle().Render() = %q, want it styled", got)
	}
}

// stripANSI removes the escape sequences lipgloss renders, which the colorprofile writer would
// normally downsample away on its way to a non-terminal.
func stripANSI(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			for i < len(s) && s[i] != 'm' {
				i++
			}

			continue
		}

		b.WriteByte(s[i])
	}

	return b.String()
}

// Under `go test` stdout is not a terminal, which is what puts every command test on the plain path:
// no spinner, and nobody to ask about the background. Dark is the answer to the second, which is
// what lipgloss falls back to as well.
func TestWithoutATerminalNothingSpinsAndTheBackgroundIsDark(t *testing.T) {
	if StdoutIsTerminal() {
		t.Error("StdoutIsTerminal() = true, want false under go test")
	}

	if !HasDarkBackground() {
		t.Error("HasDarkBackground() = false, want true when stdout is not a terminal")
	}
}
