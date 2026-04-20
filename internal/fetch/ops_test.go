package fetch_test

import (
	"testing"

	"gitmulti/internal/fetch"
	"gitmulti/internal/testutil"
)

func TestAll_NoPanic(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Should not panic.
	fetch.All([]string{local})
}

// TestAll_UpdatesRemoteTrackingRef verifies that after FetchAll, the remote
// tracking ref reflects the new commit pushed to the remote.
func TestAll_UpdatesRemoteTrackingRef(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Push a new commit to remote.
	testutil.WriteFile(t, src, "new.txt", "content")
	testutil.GitMustRun(t, src, "add", "new.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "remote advance")
	testutil.GitMustRun(t, src, "push")

	remoteHeadBefore, _ := testutil.GitOutput(t, local, "rev-parse", "origin/main")

	fetch.All([]string{local})

	remoteHeadAfter, _ := testutil.GitOutput(t, local, "rev-parse", "origin/main")
	if remoteHeadBefore == remoteHeadAfter {
		t.Error("expected origin/main to be updated after All")
	}
}

// TestAll_NoRemote verifies that FetchAll doesn't panic on a repo with no remote.
func TestAll_NoRemote(t *testing.T) {
	dir := testutil.InitRepo(t)
	fetch.All([]string{dir})
}

// TestAll_MultipleRepos verifies multi-repo mode doesn't panic.
func TestAll_MultipleRepos(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local1 := testutil.CloneRepo(t, bare)
	local2 := testutil.CloneRepo(t, bare)

	fetch.All([]string{local1, local2})
}
