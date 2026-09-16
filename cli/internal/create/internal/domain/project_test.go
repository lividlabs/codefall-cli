package domain

import (
	"testing"

	"github.com/samber/mo"
)

func TestReadme(t *testing.T) {
	for _, tc := range []struct {
		name        string
		description mo.Option[string]
		want        string
	}{
		{"title only", mo.None[string](), "# app\n"},
		{"with a description", mo.Some("A tool for things."), "# app\n\nA tool for things.\n"},
		{"a blank description is none", mo.Some("   "), "# app\n"},
		{"the description is trimmed", mo.Some("  A tool.\n"), "# app\n\nA tool.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Readme("app", tc.description); got != tc.want {
				t.Errorf("Readme = %q, want %q", got, tc.want)
			}
		})
	}
}
