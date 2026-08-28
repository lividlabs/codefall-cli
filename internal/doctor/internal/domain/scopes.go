package domain

// RequiredScopes are the token scopes codefall cannot work without.
var RequiredScopes = []string{"repo"}

// RecommendedScopes are the token scopes codefall wants but can do without for now.
var RecommendedScopes = []string{"read:project", "project"}

// impliedBy maps a scope to the scopes that already grant it. GitHub's "project" is read/write, so
// holding it satisfies "read:project" and asking for both would be noise.
var impliedBy = map[string][]string{
	"read:project": {"project"},
}

// MissingScopes separates the scopes that block codefall from the ones that only limit it.
type MissingScopes struct {
	Required    []string
	Recommended []string
}

// IsEmpty reports whether every scope codefall looks for is granted.
func (m MissingScopes) IsEmpty() bool {
	return len(m.Required) == 0 && len(m.Recommended) == 0
}

// FindMissingScopes compares the scopes a token carries against the ones codefall looks for.
func FindMissingScopes(granted []string) MissingScopes {
	held := make(map[string]bool, len(granted))
	for _, scope := range granted {
		held[scope] = true
	}

	var missing MissingScopes

	for _, scope := range RequiredScopes {
		if !satisfies(held, scope) {
			missing.Required = append(missing.Required, scope)
		}
	}

	for _, scope := range RecommendedScopes {
		if !satisfies(held, scope) {
			missing.Recommended = append(missing.Recommended, scope)
		}
	}

	return missing
}

func satisfies(held map[string]bool, scope string) bool {
	if held[scope] {
		return true
	}

	for _, parent := range impliedBy[scope] {
		if held[parent] {
			return true
		}
	}

	return false
}
