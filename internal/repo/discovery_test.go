package repo_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gitmulti/internal/repo"
	"gitmulti/internal/testutil"
)

func TestFindGitRepos(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"alpha", "beta"} {
		sub := filepath.Join(root, name)
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "init", "-b", "main")
		cmd.Dir = sub
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init: %v\n%s", err, out)
		}
	}
	_ = os.MkdirAll(filepath.Join(root, "docs"), 0o755)

	repos := repo.FindGitRepos(root)
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d: %v", len(repos), repos)
	}
}

func TestFindGitRepos_Empty(t *testing.T) {
	root := t.TempDir()
	repos := repo.FindGitRepos(root)
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestCurrentBranch(t *testing.T) {
	dir := testutil.InitRepo(t)
	if b := repo.CurrentBranch(dir); b != "main" {
		t.Errorf("expected main, got %q", b)
	}
}

func TestHasUncommittedChanges_Clean(t *testing.T) {
	dir := testutil.InitRepo(t)
	if repo.HasUncommittedChanges(dir) {
		t.Error("expected no uncommitted changes on clean repo")
	}
}

func TestHasUncommittedChanges_Dirty(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "new.txt", "dirty")
	testutil.GitMustRun(t, dir, "add", "new.txt")
	if !repo.HasUncommittedChanges(dir) {
		t.Error("expected uncommitted changes after staging a file")
	}
}

func TestBranchExistsLocal(t *testing.T) {
	dir := testutil.InitRepo(t)
	if !repo.BranchExistsLocal(dir, "main") {
		t.Error("expected main to exist locally")
	}
	if repo.BranchExistsLocal(dir, "nonexistent") {
		t.Error("expected nonexistent branch to not exist")
	}
}

func TestIsGitRepo(t *testing.T) {
	dir := testutil.InitRepo(t)
	if !repo.IsGitRepo(dir) {
		t.Error("expected dir to be a git repo")
	}
	if repo.IsGitRepo(t.TempDir()) {
		t.Error("expected plain dir to not be a git repo")
	}
}
