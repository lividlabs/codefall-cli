package domain

import (
	"slices"
	"testing"
)

func TestFindMissingScopes(t *testing.T) {
	for _, tc := range []struct {
		name            string
		granted         []string
		wantRequired    []string
		wantRecommended []string
	}{
		{
			// The scope set on a token that already works, taken from `gh auth status --json hosts`.
			name: "a working token",
			granted: []string{
				"admin:public_key", "gist", "project", "read:org", "read:packages", "repo", "workflow",
			},
		},
		{
			name:    "project alone satisfies read:project",
			granted: []string{"repo", "project"},
		},
		{
			name:            "read:project alone does not grant project",
			granted:         []string{"repo", "read:project"},
			wantRecommended: []string{"project"},
		},
		{
			name:         "repo missing",
			granted:      []string{"project"},
			wantRequired: []string{"repo"},
		},
		{
			name:            "neither project scope",
			granted:         []string{"repo"},
			wantRecommended: []string{"read:project", "project"},
		},
		{
			name:            "nothing granted",
			granted:         nil,
			wantRequired:    []string{"repo"},
			wantRecommended: []string{"read:project", "project"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			missing := FindMissingScopes(tc.granted)

			if !slices.Equal(missing.Required, tc.wantRequired) {
				t.Errorf("Required = %q, want %q", missing.Required, tc.wantRequired)
			}

			if !slices.Equal(missing.Recommended, tc.wantRecommended) {
				t.Errorf("Recommended = %q, want %q", missing.Recommended, tc.wantRecommended)
			}

			wantEmpty := len(tc.wantRequired) == 0 && len(tc.wantRecommended) == 0
			if got := missing.IsEmpty(); got != wantEmpty {
				t.Errorf("IsEmpty() = %v, want %v", got, wantEmpty)
			}
		})
	}
}
