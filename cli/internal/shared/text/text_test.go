package text

import "testing"

func TestFirstLine(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", ""},
		{"one", "one"},
		{"  one  \n two \n", "one"},
		{"\nsecond", ""},
	} {
		if got := FirstLine(tc.in); got != tc.want {
			t.Errorf("FirstLine(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
