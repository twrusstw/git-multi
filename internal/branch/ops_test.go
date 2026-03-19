package branch_test

import (
	"os"
	"path/filepath"
	"testing"

	"gitmulti/internal/branch"
	"gitmulti/internal/repo"
	"gitmulti/internal/testutil"
	"gitmulti/internal/ui"
)

// ── Switch ────────────────────────────────────────────────────────────────────

func TestSwitch_AlreadyOnBranch(t *testing.T) {
	dir := testutil.InitRepo(t)
	branch.Switch(dir, "main")
	if repo.CurrentBranch(dir) != "main" {
		t.Error("should still be on main")
	}
}

func TestSwitch_ToExistingLocal(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "feature")
	testutil.GitMustRun(t, dir, "checkout", "main")

	branch.Switch(dir, "feature")
	if repo.CurrentBranch(dir) != "feature" {
		t.Errorf("expected feature, got %s", repo.CurrentBranch(dir))
	}
}

func TestSwitch_NonExistent(t *testing.T) {
	dir := testutil.InitRepo(t)
	branch.Switch(dir, "ghost")
	if repo.CurrentBranch(dir) != "main" {
		t.Errorf("expected to stay on main, got %s", repo.CurrentBranch(dir))
	}
}

func TestSwitchForce_WithStagedChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "feature")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "change")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")

	branch.SwitchForce(dir, "feature")
	if repo.CurrentBranch(dir) != "feature" {
		t.Errorf("expected feature, got %s", repo.CurrentBranch(dir))
	}
}

// ── CreateIfModified ──────────────────────────────────────────────────────────

func TestCreateIfModified_NoChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	branch.CreateIfModified(dir, "new-branch")
	if repo.BranchExistsLocal(dir, "new-branch") {
		t.Error("branch should not be created when repo is clean")
	}
}

func TestCreateIfModified_WithRealChange(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "app.go", "package main")
	testutil.GitMustRun(t, dir, "add", "app.go")

	branch.CreateIfModified(dir, "feature-x")
	if repo.CurrentBranch(dir) != "feature-x" {
		t.Errorf("expected feature-x, got %s", repo.CurrentBranch(dir))
	}
}

func TestCreateIfModified_OnlySkipFiles(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "go.mod", "module foo")
	testutil.WriteFile(t, dir, "go.sum", "hash")
	testutil.GitMustRun(t, dir, "add", "go.mod", "go.sum")

	branch.CreateIfModified(dir, "should-not-create")
	if repo.BranchExistsLocal(dir, "should-not-create") {
		t.Error("branch should not be created when only skip files changed")
	}
	if repo.CurrentBranch(dir) != "main" {
		t.Errorf("expected to stay on main, got %s", repo.CurrentBranch(dir))
	}
}

func TestCreateIfModified_SkipPlusReal(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "go.mod", "module foo")
	testutil.WriteFile(t, dir, "main.go", "package main")
	testutil.GitMustRun(t, dir, "add", "go.mod", "main.go")

	branch.CreateIfModified(dir, "mixed-branch")
	if repo.CurrentBranch(dir) != "mixed-branch" {
		t.Errorf("expected mixed-branch, got %s", repo.CurrentBranch(dir))
	}
}

func TestCreateIfModified_BranchAlreadyExists(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "existing")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "app.go", "package main")
	testutil.GitMustRun(t, dir, "add", "app.go")

	branch.CreateIfModified(dir, "existing")
	if repo.CurrentBranch(dir) != "main" {
		t.Errorf("expected to stay on main, got %s", repo.CurrentBranch(dir))
	}
}

func TestCreateIfModified_CustomSkipFile(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, ".gitmulti-skip", "generated.pb.go\n")
	testutil.GitMustRun(t, dir, "add", ".gitmulti-skip")
	testutil.GitMustRun(t, dir, "commit", "-m", "add skip list")

	testutil.WriteFile(t, dir, "generated.pb.go", "// generated")
	testutil.GitMustRun(t, dir, "add", "generated.pb.go")

	branch.CreateIfModified(dir, "should-not-create")
	if repo.BranchExistsLocal(dir, "should-not-create") {
		t.Error("branch should not be created when only custom skip file changed")
	}
}

// ── skip helpers (tested via exported behaviour) ──────────────────────────────

func TestDefaultSkipFiles_NoBranchForLockFiles(t *testing.T) {
	lockFiles := []string{
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"Cargo.lock", "poetry.lock", "Pipfile.lock", "composer.lock",
	}
	for _, f := range lockFiles {
		t.Run(f, func(t *testing.T) {
			dir := testutil.InitRepo(t)
			testutil.WriteFile(t, dir, f, "lock content")
			testutil.GitMustRun(t, dir, "add", f)
			branch.CreateIfModified(dir, "should-skip")
			if repo.BranchExistsLocal(dir, "should-skip") {
				t.Errorf("branch created for lock file %s — should be skipped", f)
			}
		})
	}
}

func TestNonDefaultFilename_NotSkipped(t *testing.T) {
	// config.mod looks like it has a .mod suffix but basename is not "go.mod"
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "config.mod", "data")
	testutil.GitMustRun(t, dir, "add", "config.mod")

	branch.CreateIfModified(dir, "real-branch")
	if repo.CurrentBranch(dir) != "real-branch" {
		t.Errorf("config.mod should NOT be skipped, expected real-branch, got %s", repo.CurrentBranch(dir))
	}
}

// ── FindGitRepos (via repo package, tested here for integration) ──────────────

func TestListAll_NoPanic(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"repo-a", "repo-b"} {
		sub := filepath.Join(root, name)
		_ = os.MkdirAll(sub, 0o755)
		testutil.GitMustRun(t, sub, "init", "-b", "main")
		testutil.GitMustRun(t, sub, "config", "user.email", "t@t.com")
		testutil.GitMustRun(t, sub, "config", "user.name", "T")
		testutil.WriteFile(t, sub, "f.txt", "x")
		testutil.GitMustRun(t, sub, "add", ".")
		testutil.GitMustRun(t, sub, "commit", "-m", "init")
		testutil.GitMustRun(t, sub, "checkout", "-b", "shared-feature")
		testutil.GitMustRun(t, sub, "checkout", "main")
	}
	// Should not panic.
	branch.ListAll(root, "")
}

// ── stdin injection helper ────────────────────────────────────────────────────

func TestSwitch_DiscardPrompt_Yes(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "target")
	testutil.GitMustRun(t, dir, "checkout", "main")

	// Stage a change so the prompt is triggered.
	testutil.WriteFile(t, dir, "dirty.txt", "x")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("y\n")
	defer func() { ui.StdinReader = orig }()

	branch.Switch(dir, "target")
	if repo.CurrentBranch(dir) != "target" {
		t.Errorf("expected target after accepting discard prompt, got %s", repo.CurrentBranch(dir))
	}
}

func TestSwitch_DiscardPrompt_No(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "target")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "x")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("n\n")
	defer func() { ui.StdinReader = orig }()

	branch.Switch(dir, "target")
	if repo.CurrentBranch(dir) != "main" {
		t.Errorf("expected to stay on main after declining prompt, got %s", repo.CurrentBranch(dir))
	}
}
