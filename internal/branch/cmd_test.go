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
	if err := c.Run("", []string{}, []string{"-s"}); err == nil {
		t.Error("expected error: -s requires branch name")
	}
	if err := c.Run("", []string{}, []string{"-d"}); err == nil {
		t.Error("expected error: -d requires branch name")
	}
}

func TestSwitchComplete(t *testing.T) {
	c := branch.SwitchCmd()
	got := c.Complete([]string{"-"})
	flags := map[string]bool{}
	for _, s := range got {
		flags[s] = true
	}
	if !flags["-s"] || !flags["-d"] {
		t.Errorf("SwitchCmd.Complete([\"-\"]) should include -s and -d, got %v", got)
	}
}

func TestBranchRunValidation(t *testing.T) {
	c := branch.Cmd()
	if err := c.Run("", []string{}, []string{"--unknown"}); err == nil {
		t.Error("expected error for unknown branch flag")
	}
	if err := c.Run("", []string{}, []string{"-d"}); err == nil {
		t.Error("expected error: -d requires branch name")
	}
	if err := c.Run("", []string{}, []string{"-n"}); err == nil {
		t.Error("expected error: -n requires branch name")
	}
}

func TestBranchComplete(t *testing.T) {
	c := branch.Cmd()
	got := c.Complete([]string{"-"})
	want := map[string]bool{"-a": true, "--find": true, "-d": true, "-D": true, "-m": true, "-n": true}
	for _, s := range got {
		delete(want, s)
	}
	if len(want) > 0 {
		t.Errorf("Cmd.Complete([\"-\"]) missing flags: %v", want)
	}
}
