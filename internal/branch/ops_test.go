package branch_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestSwitchStash_WithUntrackedChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "feature")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "change\n")

	branch.SwitchStash(dir, "feature")
	if repo.CurrentBranch(dir) != "feature" {
		t.Errorf("expected feature, got %s", repo.CurrentBranch(dir))
	}
	out, _ := testutil.GitOutput(t, dir, "status", "--short")
	if out != "?? dirty.txt" {
		t.Errorf("expected untracked file to be restored, got %q", out)
	}
}

func TestSwitchDiscard_RemovesUntrackedChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "feature")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "change\n")

	branch.SwitchDiscard(dir, "feature")
	if repo.CurrentBranch(dir) != "feature" {
		t.Errorf("expected feature, got %s", repo.CurrentBranch(dir))
	}
	if _, err := os.Stat(filepath.Join(dir, "dirty.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected dirty.txt to be removed, err=%v", err)
	}
	out, _ := testutil.GitOutput(t, dir, "status", "--short")
	if out != "" {
		t.Errorf("expected clean repo after discard, got %q", out)
	}
}

func TestSwitchStash_PopConflict_LeavesTargetAndStashEntry(t *testing.T) {
	dir := testutil.InitRepo(t)

	testutil.WriteFile(t, dir, "file.txt", "original\n")
	testutil.GitMustRun(t, dir, "add", "file.txt")
	testutil.GitMustRun(t, dir, "commit", "-m", "add file")

	testutil.GitMustRun(t, dir, "checkout", "-b", "target")
	testutil.WriteFile(t, dir, "file.txt", "target change\n")
	testutil.GitMustRun(t, dir, "add", "file.txt")
	testutil.GitMustRun(t, dir, "commit", "-m", "target change")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "file.txt", "local change\n")

	branch.SwitchStash(dir, "target")

	if repo.CurrentBranch(dir) != "target" {
		t.Fatalf("expected to remain on target after stash conflict, got %s", repo.CurrentBranch(dir))
	}
	statusOut, _ := testutil.GitOutput(t, dir, "status", "--short")
	if strings.TrimSpace(statusOut) == "" {
		t.Fatal("expected conflict state after stash reapply conflict")
	}
	list, _ := testutil.GitOutput(t, dir, "stash", "list")
	if strings.TrimSpace(list) == "" {
		t.Fatal("expected stash entry to remain after failed stash pop")
	}
}

func TestSwitch_NonInteractiveNonExistentBranch_NoSideEffects(t *testing.T) {
	for _, tc := range []struct {
		name              string
		run               func(string, string)
		assertNoStashSide bool
	}{
		{name: "stash", run: branch.SwitchStash, assertNoStashSide: true},
		{name: "discard", run: branch.SwitchDiscard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := testutil.InitRepo(t)
			testutil.WriteFile(t, dir, "dirty.txt", "change\n")

			tc.run(dir, "ghost")

			if repo.CurrentBranch(dir) != "main" {
				t.Fatalf("expected to stay on main, got %s", repo.CurrentBranch(dir))
			}
			statusOut, _ := testutil.GitOutput(t, dir, "status", "--short")
			if statusOut != "?? dirty.txt" {
				t.Fatalf("expected dirty state to be preserved, got %q", statusOut)
			}
			if tc.assertNoStashSide {
				list, _ := testutil.GitOutput(t, dir, "stash", "list")
				if strings.TrimSpace(list) != "" {
					t.Fatalf("expected no stash entry when branch is missing, got %q", list)
				}
			}
			if _, err := os.Stat(filepath.Join(dir, "dirty.txt")); err != nil {
				t.Fatalf("expected dirty.txt to still exist, err=%v", err)
			}
		})
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

func initNamedRepo(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir repo %s: %v", name, err)
	}
	testutil.GitMustRun(t, dir, "init", "-b", "main")
	testutil.GitMustRun(t, dir, "config", "user.email", "test@test.com")
	testutil.GitMustRun(t, dir, "config", "user.name", "Test")
	testutil.WriteFile(t, dir, "README.md", "init")
	testutil.GitMustRun(t, dir, "add", ".")
	testutil.GitMustRun(t, dir, "commit", "-m", "init")
	return dir
}

func TestFindAll_GroupsExactBranchesByUpstream(t *testing.T) {
	root := t.TempDir()
	source := initNamedRepo(t, root, "source")
	ready := filepath.Join(root, "ready-repo")
	testutil.GitMustRun(t, root, "clone", source, ready)
	testutil.GitMustRun(t, ready, "config", "user.email", "test@test.com")
	testutil.GitMustRun(t, ready, "config", "user.name", "Test")
	testutil.GitMustRun(t, ready, "checkout", "-b", "feature/ready")
	testutil.GitMustRun(t, ready, "push", "-u", "origin", "feature/ready")
	testutil.GitMustRun(t, ready, "checkout", "main")

	localOnly := initNamedRepo(t, root, "local-only")
	testutil.GitMustRun(t, localOnly, "checkout", "-b", "feature/ready")
	testutil.GitMustRun(t, localOnly, "checkout", "main")

	out := testutil.CaptureStdout(t, func() {
		branch.FindAll([]string{ready, localOnly}, "feature/ready")
	})

	for _, want := range []string{
		"Find branch: feature/ready",
		"Ready to use",
		"ready-repo",
		"origin/feature/ready",
		"Local only",
		"local-only",
		"no upstream",
		"Summary: exact 2, with upstream 1, no upstream 1, similar 0, no match 0",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestFindAll_ShowsSimilarAndHidesNoMatchRepos(t *testing.T) {
	root := t.TempDir()
	similar := initNamedRepo(t, root, "similar-repo")
	testutil.GitMustRun(t, similar, "checkout", "-b", "feature/ready-fix")
	testutil.GitMustRun(t, similar, "checkout", "main")
	noMatch := initNamedRepo(t, root, "no-match")

	out := testutil.CaptureStdout(t, func() {
		branch.FindAll([]string{similar, noMatch}, "feature/ready")
	})

	for _, want := range []string{
		"Similar matches",
		"similar-repo",
		"feature/ready-fix",
		"Summary: exact 0, with upstream 0, no upstream 0, similar 1, no match 1",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "no-match") {
		t.Fatalf("expected no-match repo to be hidden from main output, got:\n%s", out)
	}
}

func TestFindAll_TruncatesSimilarMatches(t *testing.T) {
	root := t.TempDir()
	dir := initNamedRepo(t, root, "many-similar")
	for _, name := range []string{
		"feature/many-a",
		"feature/many-b",
		"feature/many-c",
		"feature/many-d",
		"feature/many-e",
	} {
		testutil.GitMustRun(t, dir, "checkout", "-b", name)
		testutil.GitMustRun(t, dir, "checkout", "main")
	}

	out := testutil.CaptureStdout(t, func() {
		branch.FindAll([]string{dir}, "feature/many")
	})

	if !strings.Contains(out, "... +2 more") {
		t.Fatalf("expected truncated similar matches, got:\n%s", out)
	}
	if strings.Contains(out, "feature/many-d") || strings.Contains(out, "feature/many-e") {
		t.Fatalf("expected later similar matches to be hidden, got:\n%s", out)
	}
}

// ── stdin injection helper ────────────────────────────────────────────────────

func TestSwitch_StashPrompt_ChoiceOne(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "target")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "x\n")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\n")
	defer func() { ui.StdinReader = orig }()

	branch.Switch(dir, "target")
	if repo.CurrentBranch(dir) != "target" {
		t.Errorf("expected target after stash prompt, got %s", repo.CurrentBranch(dir))
	}
	out, _ := testutil.GitOutput(t, dir, "status", "--short")
	if out != "?? dirty.txt" {
		t.Errorf("expected dirty.txt to be restored after stash prompt, got %q", out)
	}
}

func TestSwitch_DiscardPrompt_ChoiceTwo(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "target")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "x")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("2\n")
	defer func() { ui.StdinReader = orig }()

	branch.Switch(dir, "target")
	if repo.CurrentBranch(dir) != "target" {
		t.Errorf("expected to switch after discard prompt, got %s", repo.CurrentBranch(dir))
	}
	out, _ := testutil.GitOutput(t, dir, "status", "--short")
	if out != "" {
		t.Errorf("expected clean repo after discard prompt, got %q", out)
	}
}

func TestSwitch_CancelPrompt_ChoiceThree(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "target")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "x")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("3\n")
	defer func() { ui.StdinReader = orig }()

	branch.Switch(dir, "target")
	if repo.CurrentBranch(dir) != "main" {
		t.Errorf("expected to stay on main after cancel prompt, got %s", repo.CurrentBranch(dir))
	}
	out, _ := testutil.GitOutput(t, dir, "status", "--short")
	if out != "A  dirty.txt" {
		t.Errorf("expected dirty state to be preserved after cancel, got %q", out)
	}
}

func TestSwitch_InvalidPromptInput_Cancels(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.GitMustRun(t, dir, "checkout", "-b", "target")
	testutil.GitMustRun(t, dir, "checkout", "main")

	testutil.WriteFile(t, dir, "dirty.txt", "x")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("wat\n")
	defer func() { ui.StdinReader = orig }()

	branch.Switch(dir, "target")
	if repo.CurrentBranch(dir) != "main" {
		t.Fatalf("expected to stay on main after invalid prompt input, got %s", repo.CurrentBranch(dir))
	}
	out, _ := testutil.GitOutput(t, dir, "status", "--short")
	if out != "A  dirty.txt" {
		t.Fatalf("expected dirty state to be preserved after invalid input, got %q", out)
	}
}

func TestSwitchCmd_MultiRepoCancelOnlyAffectsCurrentRepo(t *testing.T) {
	dir1 := testutil.InitRepo(t)
	dir2 := testutil.InitRepo(t)

	for _, dir := range []string{dir1, dir2} {
		testutil.GitMustRun(t, dir, "checkout", "-b", "target")
		testutil.GitMustRun(t, dir, "checkout", "main")
		testutil.WriteFile(t, dir, "dirty.txt", "x\n")
	}

	c := branch.SwitchCmd()
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("3\n1\n")
	defer func() { ui.StdinReader = orig }()

	if err := c.Run("", []string{dir1, dir2}, []string{"target"}); err != nil {
		t.Fatalf("switch command failed: %v", err)
	}

	if repo.CurrentBranch(dir1) != "main" {
		t.Fatalf("expected first repo to stay on main after cancel, got %s", repo.CurrentBranch(dir1))
	}
	out1, _ := testutil.GitOutput(t, dir1, "status", "--short")
	if out1 != "?? dirty.txt" {
		t.Fatalf("expected first repo dirty state to remain, got %q", out1)
	}

	if repo.CurrentBranch(dir2) != "target" {
		t.Fatalf("expected second repo to switch after second prompt, got %s", repo.CurrentBranch(dir2))
	}
	out2, _ := testutil.GitOutput(t, dir2, "status", "--short")
	if out2 != "?? dirty.txt" {
		t.Fatalf("expected second repo to restore untracked file after stash, got %q", out2)
	}
}
