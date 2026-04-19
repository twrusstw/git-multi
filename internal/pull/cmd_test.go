package pull_test

import (
	"testing"

	"gitmulti/internal/pull"
)

func TestPullComplete(t *testing.T) {
	c := pull.Cmd()

	got := c.Complete([]string{"--reb"})
	found := false
	for _, s := range got {
		if s == "--rebase" {
			found = true
		}
	}
	if !found {
		t.Errorf("Complete([\"--reb\"]) missing \"--rebase\", got %v", got)
	}

	got2 := c.Complete([]string{"--rebase", "--"})
	for _, s := range got2 {
		if s == "--rebase" {
			t.Errorf("Complete after --rebase should not suggest --rebase again, got %v", got2)
		}
	}
}

func TestPullRunInvalidBranch(t *testing.T) {
	c := pull.Cmd()
	err := c.Run("", []string{}, []string{"bad branch!!"})
	if err == nil {
		t.Error("expected error for invalid branch name, got nil")
	}
}
