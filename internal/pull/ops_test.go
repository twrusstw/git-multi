package pull_test

import (
	"strings"
	"testing"

	"gitmulti/internal/pull"
	"gitmulti/internal/repo"
	"gitmulti/internal/testutil"
	"gitmulti/internal/ui"
)

// ── Phase 1: ff-only ──────────────────────────────────────────────────────────

func TestAll_UpToDate(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)
	pull.All([]string{local}, "main")

	if repo.CurrentBranch(local) != "main" {
		t.Errorf("expected main, got %s", repo.CurrentBranch(local))
	}
}

func TestAll_WithRemoteCommit(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	testutil.WriteFile(t, src, "new.txt", "hello")
	testutil.GitMustRun(t, src, "add", "new.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "add new.txt")
	testutil.GitMustRun(t, src, "push")

	pull.All([]string{local}, "main")

	out, err := testutil.GitOutput(t, local, "log", "--oneline", "-2")
	if err != nil || len(out) == 0 {
		t.Error("expected log entries after pull")
	}
}

// ── Phase 2: stash → ff-only → pop ───────────────────────────────────────────

// TestAll_StashFallback_Success: local has unstaged changes to one region
// of file.txt, remote changed a different region. ff-only refuses due to file-level
// dirty check, but stash → ff-only → pop auto-merges successfully.
func TestAll_StashFallback_Success(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)

	// Multi-line file so changes to different regions don't conflict.
	testutil.WriteFile(t, src, "file.txt",
		"line1\nline2\nline3\nline4\nline5\n")
	testutil.GitMustRun(t, src, "add", "file.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "init file")
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Remote changes the last line.
	testutil.WriteFile(t, src, "file.txt",
		"line1\nline2\nline3\nline4\nline5-remote\n")
	testutil.GitMustRun(t, src, "add", "file.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "remote changes line5")
	testutil.GitMustRun(t, src, "push")

	// Local changes the first line (tracked, unstaged) → ff-only will refuse.
	testutil.WriteFile(t, local, "file.txt",
		"line1-local\nline2\nline3\nline4\nline5\n")

	// Provide "skip all" in case Phase 3 menu triggers unexpectedly.
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("4\n")
	defer func() { ui.StdinReader = orig }()

	pull.All([]string{local}, "main")

	// Remote's line5 change must be present after pull.
	content, err := testutil.GitOutput(t, local, "show", "HEAD:file.txt")
	if err != nil {
		t.Fatalf("git show failed: %v", err)
	}
	if !strings.Contains(content, "line5-remote") {
		t.Errorf("expected remote change (line5-remote) in HEAD, got:\n%s", content)
	}
}

// TestAll_StashFallback_PopConflict: local and remote both modify the same
// line. ff-only refuses, stash + ff-only succeeds, but stash pop conflicts →
// cascade enters Phase 3 and the user chooses "skip all".
func TestAll_StashFallback_PopConflict(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)

	testutil.WriteFile(t, src, "file.txt", "original content\n")
	testutil.GitMustRun(t, src, "add", "file.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "init")
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	// Remote changes the same line as local will.
	testutil.WriteFile(t, src, "file.txt", "remote content\n")
	testutil.GitMustRun(t, src, "add", "file.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "remote change")
	testutil.GitMustRun(t, src, "push")

	// Local changes the same line (unstaged).
	testutil.WriteFile(t, local, "file.txt", "local content\n")

	// Inject "skip all" for the Phase 3 conflict menu.
	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("4\n")
	defer func() { ui.StdinReader = orig }()

	// Should not panic; conflict is gracefully handled via the menu.
	pull.All([]string{local}, "main")
}

// ── Phase 3: conflict menu ────────────────────────────────────────────────────

// divergedLocalRemote creates a local repo and a bare remote where both have
// diverged (one commit on each side on main).
func divergedLocalRemote(t *testing.T) (local, bare string) {
	t.Helper()
	bare = testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local = testutil.CloneRepo(t, bare)

	// Remote advances.
	testutil.WriteFile(t, src, "remote.txt", "from remote")
	testutil.GitMustRun(t, src, "add", "remote.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "remote commit")
	testutil.GitMustRun(t, src, "push")

	// Local also advances (diverged).
	testutil.WriteFile(t, local, "local.txt", "from local")
	testutil.GitMustRun(t, local, "add", "local.txt")
	testutil.GitMustRun(t, local, "commit", "-m", "local commit")

	return local, bare
}

func TestAll_Phase3_ResetHard(t *testing.T) {
	local, _ := divergedLocalRemote(t)

	// Fetch first so origin/main is known.
	testutil.GitMustRun(t, local, "fetch", "origin")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("1\n") // reset --hard
	defer func() { ui.StdinReader = orig }()

	pull.All([]string{local}, "main")

	// After reset --hard, local should match origin/main (no local.txt).
	out, _ := testutil.GitOutput(t, local, "show", "HEAD:local.txt")
	if out != "" {
		t.Error("expected local.txt to be gone after reset --hard")
	}
	out, _ = testutil.GitOutput(t, local, "show", "HEAD:remote.txt")
	if out != "from remote" {
		t.Errorf("expected remote.txt to be present, got %q", out)
	}
}

func TestAll_Phase3_SkipAll(t *testing.T) {
	local, _ := divergedLocalRemote(t)
	testutil.GitMustRun(t, local, "fetch", "origin")

	localHeadBefore, _ := testutil.GitOutput(t, local, "rev-parse", "HEAD")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("4\n") // skip all
	defer func() { ui.StdinReader = orig }()

	pull.All([]string{local}, "main")

	// HEAD unchanged — local commit preserved.
	localHeadAfter, _ := testutil.GitOutput(t, local, "rev-parse", "HEAD")
	if localHeadBefore != localHeadAfter {
		t.Error("expected HEAD to be unchanged after skip all")
	}
}

func TestAll_Phase3_Merge(t *testing.T) {
	local, _ := divergedLocalRemote(t)
	testutil.GitMustRun(t, local, "fetch", "origin")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("2\n") // merge
	defer func() { ui.StdinReader = orig }()

	// Should not panic; merge may succeed (no content conflict).
	pull.All([]string{local}, "main")
}

func TestAll_Phase3_Rebase(t *testing.T) {
	local, _ := divergedLocalRemote(t)
	testutil.GitMustRun(t, local, "fetch", "origin")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("3\n") // rebase
	defer func() { ui.StdinReader = orig }()

	pull.All([]string{local}, "main")
}

// ── pull --rebase ─────────────────────────────────────────────────────────────

func TestRebase_WithRemoteCommit(t *testing.T) {
	bare := testutil.InitBareRepo(t)
	src := testutil.InitRepo(t)
	testutil.GitMustRun(t, src, "remote", "add", "origin", bare)
	testutil.GitMustRun(t, src, "push", "-u", "origin", "main")

	local := testutil.CloneRepo(t, bare)

	testutil.WriteFile(t, src, "remote.txt", "remote")
	testutil.GitMustRun(t, src, "add", "remote.txt")
	testutil.GitMustRun(t, src, "commit", "-m", "remote commit")
	testutil.GitMustRun(t, src, "push")

	pull.Rebase([]string{local}, "main")

	out, err := testutil.GitOutput(t, local, "log", "--oneline", "-2")
	if err != nil || len(out) == 0 {
		t.Error("expected log entries after pull --rebase")
	}
}
