package push_test

import (
	"testing"

	"gitmulti/internal/push"
	"gitmulti/internal/testutil"
)

func TestPush_WithUpstream(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	testutil.WriteFile(t, src, "new.txt", "hello")
	testutil.GitMustRun(t, src, "add", "new.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "add new.txt")

	push.Push(src, "")

	out, err := testutil.GitOutput(t, bare, "log", "--oneline", "-1")
	if err != nil || out == "" {
		t.Error("expected commit in bare repo after push")
	}
}

func TestPush_NoUpstream_SetsTracking(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "origin", "main")
	testutil.GitMustRun(t, src, "checkout", "-b", "feature")
	testutil.WriteFile(t, src, "feat.txt", "feature")
	testutil.GitMustRun(t, src, "add", "feat.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "feature commit")

	push.Push(src, "feature")

	out, err := testutil.GitOutput(t, src, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		t.Errorf("expected tracking branch after push, got error: %v", err)
	}
	if out != "origin/feature" {
		t.Errorf("expected origin/feature, got %s", out)
	}
}

// TestPush_Rejected: two clones diverge; the second push should be rejected.
func TestPush_Rejected(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	clone1 := testutil.CloneRepo(t, bare)
	clone2 := testutil.CloneRepo(t, bare)

	// clone1 pushes first.
	testutil.WriteFile(t, clone1, "c1.txt", "clone1")
	testutil.GitMustRun(t, clone1, "add", "c1.txt")
	testutil.GitMustRun(t, clone1, "commit", "-m", "clone1 commit")
	testutil.GitMustRun(t, clone1, "push")

	// clone2 also commits locally (diverged) then tries to push.
	testutil.WriteFile(t, clone2, "c2.txt", "clone2")
	testutil.GitMustRun(t, clone2, "add", "c2.txt")
	testutil.GitMustRun(t, clone2, "commit", "-m", "clone2 commit")

	// push.Push should handle the rejection gracefully (no panic).
	push.Push(clone2, "")

	// Verify bare still has clone1's commit, not clone2's.
	out, _ := testutil.GitOutput(t, bare, "show", "HEAD:c1.txt")
	if out != "clone1" {
		t.Errorf("expected clone1's commit to be in bare, got %q", out)
	}
	_, err := testutil.GitOutput(t, bare, "show", "HEAD:c2.txt")
	if err == nil {
		t.Error("expected clone2's commit NOT to be in bare after rejected push")
	}
}
