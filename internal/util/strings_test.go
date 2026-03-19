package util_test

import (
	"testing"

	"gitmulti/internal/util"
)

func TestNonEmpty(t *testing.T) {
	got := util.NonEmpty([]string{"a", "", "  ", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestNormaliseBranchName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"main", "main"},
		{"* main", "main"},
		{"  feature/x  ", "feature/x"},
		{"remotes/origin/main", "main"},
		{"remotes/origin/feature/login", "feature/login"},
		{"remotes/origin/HEAD", ""},
	}
	for _, tc := range cases {
		got := util.NormaliseBranchName(tc.input)
		if got != tc.want {
			t.Errorf("NormaliseBranchName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
