package stash_test

import (
	"testing"

	"gitmulti/internal/stash"
)

func TestStashRunUnknownSubcommand(t *testing.T) {
	c := stash.Cmd()
	err := c.Run("", []string{}, []string{"badsubcmd"})
	if err == nil {
		t.Error("expected error for unknown stash subcommand")
	}
}

func TestStashComplete(t *testing.T) {
	c := stash.Cmd()
	got := c.Complete([]string{"p"})
	found := false
	for _, s := range got {
		if s == "pop" {
			found = true
		}
	}
	if !found {
		t.Errorf("Complete([\"p\"]) should include \"pop\", got %v", got)
	}
}
