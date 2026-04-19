package stash_test

import (
	"strings"
	"testing"

	"gitmulti/internal/stash"
	"gitmulti/internal/testutil"
)

func TestStash_DirtyRepo(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "dirty.txt", "changes")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")

	stash.Stash([]string{dir})

	out, _ := testutil.GitOutput(t, dir, "status", "--porcelain")
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected clean after stash, got: %s", out)
	}
	list, _ := testutil.GitOutput(t, dir, "stash", "list")
	if strings.TrimSpace(list) == "" {
		t.Error("expected stash entry after stash")
	}
}

func TestStash_CleanRepo_NoOp(t *testing.T) {
	dir := testutil.InitRepo(t)
	stash.Stash([]string{dir})

	list, _ := testutil.GitOutput(t, dir, "stash", "list")
	if strings.TrimSpace(list) != "" {
		t.Error("expected no stash on clean repo")
	}
}

func TestPop_RestoresChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "staged.txt", "data")
	testutil.GitMustRun(t, dir, "add", "staged.txt")
	testutil.GitMustRun(t, dir, "stash")

	stash.Pop([]string{dir})

	out, _ := testutil.GitOutput(t, dir, "status", "--porcelain")
	if !strings.Contains(out, "staged.txt") {
		t.Error("expected staged.txt to be restored after stash pop")
	}
	list, _ := testutil.GitOutput(t, dir, "stash", "list")
	if strings.TrimSpace(list) != "" {
		t.Error("expected stash to be empty after pop")
	}
}

func TestApply_KeepsStash(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "staged.txt", "data")
	testutil.GitMustRun(t, dir, "add", "staged.txt")
	testutil.GitMustRun(t, dir, "stash")

	stash.Apply([]string{dir})

	out, _ := testutil.GitOutput(t, dir, "status", "--porcelain")
	if !strings.Contains(out, "staged.txt") {
		t.Error("expected staged.txt to be restored after stash apply")
	}
	list, _ := testutil.GitOutput(t, dir, "stash", "list")
	if strings.TrimSpace(list) == "" {
		t.Error("expected stash to remain after apply")
	}
}

func TestList_ShowsEntries(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "f.txt", "x")
	testutil.GitMustRun(t, dir, "add", "f.txt")
	testutil.GitMustRun(t, dir, "stash")

	stash.List([]string{dir})
}

func TestList_MultipleRepos_EmptyStash(t *testing.T) {
	dir1 := testutil.InitRepo(t)
	dir2 := testutil.InitRepo(t)

	// Should print "No stash entries" and not panic.
	stash.List([]string{dir1, dir2})
}

// TestPop_Conflict: stash contains changes to a file that was also modified
// and committed on the branch → stash pop creates a conflict.
// Verifies that Pop handles the conflict gracefully (no panic, prints hint).
func TestPop_Conflict(t *testing.T) {
	dir := testutil.InitRepo(t)

	// Stash a change to file.txt.
	testutil.WriteFile(t, dir, "file.txt", "stashed content\n")
	testutil.GitMustRun(t, dir, "add", "file.txt")
	testutil.GitMustRun(t, dir, "stash")

	// Now commit a conflicting change to file.txt.
	testutil.WriteFile(t, dir, "file.txt", "committed content\n")
	testutil.GitMustRun(t, dir, "add", "file.txt")
	testutil.GitMustRun(t, dir, "commit", "-m", "conflicting commit")

	// Pop should encounter a conflict and handle it gracefully.
	stash.Pop([]string{dir})

	// After a failed stash pop, the repo should still be in a usable state.
	// git status should not be empty (conflict markers present).
	out, _ := testutil.GitOutput(t, dir, "status", "--porcelain")
	if strings.TrimSpace(out) == "" {
		t.Error("expected conflict state after stash pop conflict")
	}
}

// TestApply_NoStash: Apply on a repo with no stash should be a no-op.
func TestApply_NoStash(t *testing.T) {
	dir := testutil.InitRepo(t)
	stash.Apply([]string{dir})
}

// TestPop_NoStash: Pop on a repo with no stash should be a no-op.
func TestPop_NoStash(t *testing.T) {
	dir := testutil.InitRepo(t)
	stash.Pop([]string{dir})
}
