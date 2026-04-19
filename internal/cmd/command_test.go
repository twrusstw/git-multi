package cmd_test

import (
	"testing"

	"gitmulti/internal/cmd"
)

func TestArgOrEmpty(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"main"}, "main"},
		{[]string{"--rebase", "main"}, "main"},
		{[]string{"--rebase"}, ""},
		{[]string{"-f", "feat/x"}, "feat/x"},
	}
	for _, tc := range cases {
		if got := cmd.ArgOrEmpty(tc.args); got != tc.want {
			t.Errorf("ArgOrEmpty(%v) = %q, want %q", tc.args, got, tc.want)
		}
	}
}
