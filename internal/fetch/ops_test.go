package fetch_test

import (
	"testing"

	"gitmulti/internal/fetch"
	"gitmulti/internal/testutil"
)

func TestFetchAll_NoPanic(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Should not panic.
	fetch.FetchAll([]string{local})
}

// TestFetchAll_UpdatesRemoteTrackingRef verifies that after FetchAll, the remote
// tracking ref reflects the new commit pushed to the remote.
func TestFetchAll_UpdatesRemoteTrackingRef(t *testing.T) {
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

	fetch.FetchAll([]string{local})

	remoteHeadAfter, _ := testutil.GitOutput(t, local, "rev-parse", "origin/main")
	if remoteHeadBefore == remoteHeadAfter {
		t.Error("expected origin/main to be updated after FetchAll")
	}
}

// TestFetchAll_NoRemote verifies that FetchAll doesn't panic on a repo with no remote.
func TestFetchAll_NoRemote(t *testing.T) {
	dir := testutil.InitRepo(t)
	fetch.FetchAll([]string{dir})
}

// TestFetchAll_MultipleRepos verifies multi-repo mode doesn't panic.
func TestFetchAll_MultipleRepos(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local1 := testutil.CloneRepo(t, bare)
	local2 := testutil.CloneRepo(t, bare)

	fetch.FetchAll([]string{local1, local2})
}
