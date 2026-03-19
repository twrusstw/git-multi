package pull_test

import (
	"testing"

	"gitmulti/internal/pull"
	"gitmulti/internal/repo"
	"gitmulti/internal/testutil"
)

func TestPull_UpToDate(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)
	pull.Pull(local, "main")

	if repo.CurrentBranch(local) != "main" {
		t.Errorf("expected main, got %s", repo.CurrentBranch(local))
	}
}

func TestPull_WithRemoteCommit(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Push a new commit to bare via src.
	testutil.WriteFile(t, src, "new.txt", "hello")
	testutil.GitMustRun(t, src, "add", "new.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "add new.txt")
	testutil.GitMustRun(t, src, "push")

	pull.Pull(local, "main")

	out, err := testutil.GitOutput(t, local, "log", "--oneline", "-2")
	if err != nil || len(out) == 0 {
		t.Error("expected log entries after pull")
	}
}

func TestPullForce_ResetsToRemote(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Local divergence.
	testutil.WriteFile(t, local, "local-only.txt", "diverged")
	testutil.GitMustRun(t, local, "add", "local-only.txt")
	testutil.GitMustRun(t, local, "commit", "-m", "local diverge")

	pull.PullForce(local, "main")

	out, _ := testutil.GitOutput(t, local, "show", "HEAD:local-only.txt")
	if out != "" {
		t.Error("expected local-only.txt to be gone after force pull")
	}
}
