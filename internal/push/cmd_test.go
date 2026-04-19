package push_test

import (
	"testing"

	"gitmulti/internal/push"
)

func TestPushRunInvalidBranch(t *testing.T) {
	c := push.Cmd()
	err := c.Run("", []string{}, []string{"bad!!"})
	if err == nil {
		t.Error("expected error for invalid branch name, got nil")
	}
}

func TestPushCompleteReturnsSlice(t *testing.T) {
	c := push.Cmd()
	got := c.Complete([]string{""})
	if got == nil {
		t.Error("Complete should return non-nil slice")
	}
}
