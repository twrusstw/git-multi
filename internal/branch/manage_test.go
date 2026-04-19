package branch_test

import (
	"os"
	"strings"
	"testing"

	"gitmulti/internal/branch"
	"gitmulti/internal/repo"
	"gitmulti/internal/testutil"
	"gitmulti/internal/ui"
)

// ── branch --find ─────────────────────────────────────────────────────────────

func TestFind_ExactMatch(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "feature/login")
	testutil.GitMustRun(t, dir, "checkout", "main")

	// Should not panic; no way to capture stdout in unit test, just verify no crash.
	branch.Find(dir, "feature/login")
}

func TestFind_NoMatch(t *testing.T) {
	dir := testutil.InitRepo(t)
	branch.Find(dir, "nonexistent-branch")
}

func TestFind_PartialMatch(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "feature/auth")
	testutil.GitMustRun(t, dir, "checkout", "main")
	branch.Find(dir, "auth")
}

// ── branch -d ─────────────────────────────────────────────────────────────────

func TestDelete_BranchExists(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "to-delete")
	testutil.GitMustRun(t, dir, "checkout", "main")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\n") // yes to all
	defer func() { ui.StdinReader = orig }()

	branch.Delete([]string{dir}, "to-delete", false)

	if repo.BranchExistsLocal(dir, "to-delete") {
		t.Error("expected branch to-delete to be deleted")
	}
}

func TestDelete_BranchNotFound(t *testing.T) {
	dir := testutil.InitRepo(t)
	branch.Delete([]string{dir}, "nonexistent", false)
}

func TestDelete_ConfirmEach_Yes(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "to-delete")
	testutil.GitMustRun(t, dir, "checkout", "main")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("2\ny\n") // confirm each, then yes
	defer func() { ui.StdinReader = orig }()

	branch.Delete([]string{dir}, "to-delete", false)

	if repo.BranchExistsLocal(dir, "to-delete") {
		t.Error("expected branch to be deleted after confirm-each yes")
	}
}

func TestDelete_ConfirmEach_No(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "keep-me")
	testutil.GitMustRun(t, dir, "checkout", "main")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("2\nn\n") // confirm each, then no
	defer func() { ui.StdinReader = orig }()

	branch.Delete([]string{dir}, "keep-me", false)

	if !repo.BranchExistsLocal(dir, "keep-me") {
		t.Error("expected branch to survive after confirm-each no")
	}
}

func TestDelete_Cancel(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "keep-me")
	testutil.GitMustRun(t, dir, "checkout", "main")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("3\n") // cancel
	defer func() { ui.StdinReader = orig }()

	branch.Delete([]string{dir}, "keep-me", false)

	if !repo.BranchExistsLocal(dir, "keep-me") {
		t.Error("expected branch to survive after cancel")
	}
}

func TestDelete_WithRemote(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Create and push the branch to be deleted.
	testutil.GitMustRun(t, local, "checkout", "-b", "to-delete")
	testutil.GitMustRun(t, local, "push", "-u", "origin", "to-delete")
	testutil.GitMustRun(t, local, "checkout", "main")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\n") // yes to all
	defer func() { ui.StdinReader = orig }()

	branch.Delete([]string{local}, "to-delete", true /* deleteRemote */)

	if repo.BranchExistsLocal(local, "to-delete") {
		t.Error("expected local branch to-delete to be deleted")
	}
	if repo.BranchExistsRemoteLocal(local, "to-delete") {
		t.Error("expected remote branch origin/to-delete to be deleted")
	}
}

// ── branch -D ─────────────────────────────────────────────────────────────────

func TestForceDelete_UnmergedWarn(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "unmerged")
	testutil.WriteFile(t, dir, "new.txt", "content")
	testutil.GitMustRun(t, dir, "add", "new.txt")
	testutil.GitMustRun(t, dir, "commit", "-m", "unmerged commit")
	testutil.GitMustRun(t, dir, "checkout", "main")

	// "1" = yes to all (menu), "y" = confirm force-delete unmerged warning.
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\ny\n")
	defer func() { ui.StdinReader = orig }()

	branch.ForceDelete([]string{dir}, "unmerged", false)

	if repo.BranchExistsLocal(dir, "unmerged") {
		t.Error("expected unmerged branch to be force-deleted")
	}
}

// TestForceDelete_MergedBranch: already merged, no extra warning prompt.
func TestForceDelete_MergedBranch(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "merged-feature")
	testutil.WriteFile(t, dir, "feat.txt", "content")
	testutil.GitMustRun(t, dir, "add", "feat.txt")
	testutil.GitMustRun(t, dir, "commit", "-m", "feature commit")
	testutil.GitMustRun(t, dir, "checkout", "main")
	testutil.GitMustRun(t, dir, "merge", "--no-ff", "merged-feature", "-m", "merge")

	// "1" = yes to all (no unmerged warning expected).
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\n")
	defer func() { ui.StdinReader = orig }()

	branch.ForceDelete([]string{dir}, "merged-feature", false)

	if repo.BranchExistsLocal(dir, "merged-feature") {
		t.Error("expected merged branch to be force-deleted")
	}
}

// TestForceDelete_UnmergedWarn_DeclineSkips: user declines the force-delete warning.
func TestForceDelete_UnmergedWarn_DeclineSkips(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "unmerged")
	testutil.WriteFile(t, dir, "new.txt", "x")
	testutil.GitMustRun(t, dir, "add", "new.txt")
	testutil.GitMustRun(t, dir, "commit", "-m", "unmerged")
	testutil.GitMustRun(t, dir, "checkout", "main")

	// "1" = yes to all (menu), "n" = decline force-delete warning.
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\nn\n")
	defer func() { ui.StdinReader = orig }()

	branch.ForceDelete([]string{dir}, "unmerged", false)

	if !repo.BranchExistsLocal(dir, "unmerged") {
		t.Error("expected branch to survive when user declines force-delete warning")
	}
}

// ── branch -m ─────────────────────────────────────────────────────────────────

func TestRename_BranchExists(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "old-name")
	testutil.GitMustRun(t, dir, "checkout", "main")

	// "y" = confirm rename, "3" = skip remote sync.
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("y\n3\n")
	defer func() { ui.StdinReader = orig }()

	branch.Rename([]string{dir}, "old-name", "new-name")

	if repo.BranchExistsLocal(dir, "old-name") {
		t.Error("expected old-name to be gone after rename")
	}
	if !repo.BranchExistsLocal(dir, "new-name") {
		t.Error("expected new-name to exist after rename")
	}
}

func TestRename_BranchNotFound(t *testing.T) {
	dir := testutil.InitRepo(t)
	branch.Rename([]string{dir}, "ghost", "new-ghost")
	if repo.BranchExistsLocal(dir, "new-ghost") {
		t.Error("unexpected branch new-ghost")
	}
}

func TestRename_Cancel(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "stay")
	testutil.GitMustRun(t, dir, "checkout", "main")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("n\n") // decline confirmation
	defer func() { ui.StdinReader = orig }()

	branch.Rename([]string{dir}, "stay", "renamed")

	if !repo.BranchExistsLocal(dir, "stay") {
		t.Error("expected stay to survive after cancelling rename")
	}
}

func TestRename_WithRemoteSync(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Create and push old-name to remote.
	testutil.GitMustRun(t, local, "checkout", "-b", "old-name")
	testutil.GitMustRun(t, local, "push", "-u", "origin", "old-name")
	testutil.GitMustRun(t, local, "checkout", "main")

	// "y" = confirm rename, "1" = yes to all for remote sync.
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("y\n1\n")
	defer func() { ui.StdinReader = orig }()

	branch.Rename([]string{local}, "old-name", "new-name")

	if repo.BranchExistsLocal(local, "old-name") {
		t.Error("expected old-name to be gone locally after rename")
	}
	if !repo.BranchExistsLocal(local, "new-name") {
		t.Error("expected new-name to exist locally after rename")
	}
	// Verify remote sync: old-name gone, new-name present.
	if repo.BranchExistsRemoteLocal(local, "old-name") {
		t.Error("expected origin/old-name to be deleted after remote sync")
	}

	// After pushing new-name, fetch to confirm.
	testutil.GitMustRun(t, local, "fetch", "origin")
	if !repo.BranchExistsRemoteLocal(local, "new-name") {
		t.Error("expected origin/new-name to exist after remote sync")
	}
}

// TestRename_MultipleRepos: branches exist in some repos but not others.
func TestRename_MultipleRepos(t *testing.T) {
	dir1 := testutil.InitRepo(t)
	dir2 := testutil.InitRepo(t)

	testutil.GitMustRun(t, dir1, "checkout", "-b", "old-name")
	testutil.GitMustRun(t, dir1, "checkout", "main")
	// dir2 does NOT have old-name.

	orig := ui.StdinReader
	// "y" = confirm, "3" = skip remote sync.
	ui.StdinReader = testutil.NewStringReader("y\n3\n")
	defer func() { ui.StdinReader = orig }()

	branch.Rename([]string{dir1, dir2}, "old-name", "new-name")

	if repo.BranchExistsLocal(dir1, "old-name") {
		t.Error("expected old-name to be gone in dir1")
	}
	if !repo.BranchExistsLocal(dir1, "new-name") {
		t.Error("expected new-name in dir1")
	}
	// dir2 should be untouched.
	if repo.BranchExistsLocal(dir2, "new-name") {
		t.Error("expected dir2 to be skipped (no old-name)")
	}
}

// ── stash pop conflict ────────────────────────────────────────────────────────
// (Tested via pull cascade tests, but also worth a direct smoke test.)

func TestDelete_MultipleRepos_SomeHaveBranch(t *testing.T) {
	dir1 := testutil.InitRepo(t)
	dir2 := testutil.InitRepo(t)

	testutil.GitMustRun(t, dir1, "checkout", "-b", "to-delete")
	testutil.GitMustRun(t, dir1, "checkout", "main")
	// dir2 does NOT have to-delete.

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\n") // yes to all
	defer func() { ui.StdinReader = orig }()

	branch.Delete([]string{dir1, dir2}, "to-delete", false)

	if repo.BranchExistsLocal(dir1, "to-delete") {
		t.Error("expected to-delete to be deleted from dir1")
	}
}

// TestFind_ListAllNames_WithKeyword verifies ListAllNames filters by keyword.
// ListAllNames expects a root directory containing git repo subdirectories.
func TestFind_ListAllNames_WithKeyword(t *testing.T) {
	root := t.TempDir()
	dir := root + "/repo"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.GitMustRun(t, dir, "init", "-b", "main")
	testutil.GitMustRun(t, dir, "config", "user.email", "t@t.com")
	testutil.GitMustRun(t, dir, "config", "user.name", "T")
	testutil.WriteFile(t, dir, "f.txt", "x")
	testutil.GitMustRun(t, dir, "add", ".")
	testutil.GitMustRun(t, dir, "commit", "-m", "init")

	testutil.GitMustRun(t, dir, "checkout", "-b", "feature/auth")
	testutil.GitMustRun(t, dir, "checkout", "main")
	testutil.GitMustRun(t, dir, "checkout", "-b", "feature/login")
	testutil.GitMustRun(t, dir, "checkout", "main")
	testutil.GitMustRun(t, dir, "checkout", "-b", "hotfix/bug")
	testutil.GitMustRun(t, dir, "checkout", "main")

	names := branch.ListAllNames(root, "feature")
	for _, n := range names {
		if !strings.Contains(n, "feature") {
			t.Errorf("got branch %q that doesn't match 'feature' keyword", n)
		}
	}
	found := false
	for _, n := range names {
		if n == "feature/auth" || n == "feature/login" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected feature/* branches in filtered result, got %v", names)
	}
}
