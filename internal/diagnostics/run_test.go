package diagnostics

import "testing"

func TestValidTargetSecurity(t *testing.T) {
	cases := []struct {
		target string
		valid  bool
	}{
		{"1.1.1.1", true},
		{"2001:db8::1", true},
		{"example.com", true},
		{"sub.example.com.", true},
		{"-V", false},
		{"--help", false},
		{"-c5", false},
		{"foo..example.com", false},
		{"foo/bar", false},
		{"foo;id", false},
		{"foo example.com", false},
		{"foo_example.com", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := validTarget(tc.target); got != tc.valid {
			t.Errorf("validTarget(%q) = %v, want %v", tc.target, got, tc.valid)
		}
	}
}
