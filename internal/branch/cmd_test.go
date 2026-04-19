package branch_test

import (
	"testing"

	"gitmulti/internal/branch"
)

func TestSwitchRunValidation(t *testing.T) {
	c := branch.SwitchCmd()

	if err := c.Run("", []string{}, []string{}); err == nil {
		t.Error("expected error for missing branch name")
	}
	if err := c.Run("", []string{}, []string{"bad!branch"}); err == nil {
		t.Error("expected error for invalid branch name")
	}
	if err := c.Run("", []string{}, []string{"-f"}); err == nil {
		t.Error("expected error: -f requires branch name")
	}
	if err := c.Run("", []string{}, []string{"-c"}); err == nil {
		t.Error("expected error: -c requires branch name")
	}
}

func TestSwitchComplete(t *testing.T) {
	c := branch.SwitchCmd()
	got := c.Complete([]string{"-"})
	flags := map[string]bool{}
	for _, s := range got {
		flags[s] = true
	}
	if !flags["-f"] || !flags["-c"] {
		t.Errorf("SwitchCmd.Complete([\"-\"]) should include -f and -c, got %v", got)
	}
}

func TestBranchRunValidation(t *testing.T) {
	c := branch.BranchCmd()
	if err := c.Run("", []string{}, []string{"--unknown"}); err == nil {
		t.Error("expected error for unknown branch flag")
	}
	if err := c.Run("", []string{}, []string{"-d"}); err == nil {
		t.Error("expected error: -d requires branch name")
	}
}

func TestBranchComplete(t *testing.T) {
	c := branch.BranchCmd()
	got := c.Complete([]string{"-"})
	want := map[string]bool{"-a": true, "--find": true, "-d": true, "-D": true, "-m": true}
	for _, s := range got {
		delete(want, s)
	}
	if len(want) > 0 {
		t.Errorf("BranchCmd.Complete([\"-\"]) missing flags: %v", want)
	}
}
