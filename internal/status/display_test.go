package status_test

import (
	"testing"

	"gitmulti/internal/repo"
	"gitmulti/internal/status"
	"gitmulti/internal/testutil"
	"gitmulti/internal/ui"
)

// ── extractOwner (tested via ShowCurrent integration) ─────────────────────────
// Unit tests for extractOwner live here via an exported wrapper only for tests.
// Since extractOwner is unexported, we test it indirectly through ShowCurrent,
// and add direct cases via a white-box test in the same package.

// ── ShowStatus ────────────────────────────────────────────────────────────────

func TestShowStatus_Clean(t *testing.T) {
	dir := testutil.InitRepo(t)
	status.ShowStatus(dir) // must not panic
}

func TestShowStatus_WithChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "dirty.txt", "change")
	testutil.GitMustRun(t, dir, "add", "dirty.txt")
	status.ShowStatus(dir) // must not panic
}

// ── ShowCurrent ───────────────────────────────────────────────────────────────

func TestShowCurrent_NoPanic(t *testing.T) {
	dir := testutil.InitRepo(t)
	status.ShowCurrent(dir, true)
	status.ShowCurrent(dir, false)
}

// ── DiscardChangesMulti ───────────────────────────────────────────────────────

func TestDiscardChangesMulti_NoChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	status.DiscardChangesMulti([]string{dir}) // must not panic
}

func TestDiscardChangesMulti_WithChanges_Accept(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "will-discard.txt", "dirty")
	testutil.GitMustRun(t, dir, "add", "will-discard.txt")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("y\n")
	defer func() { ui.StdinReader = orig }()

	status.DiscardChangesMulti([]string{dir})

	if repo.HasUncommittedChanges(dir) {
		t.Error("expected changes to be discarded")
	}
}

func TestDiscardChangesMulti_WithChanges_Cancel(t *testing.T) {
	dir := testutil.InitRepo(t)
	testutil.WriteFile(t, dir, "keep.txt", "keep")
	testutil.GitMustRun(t, dir, "add", "keep.txt")

	orig := ui.StdinReader
	ui.StdinReader = testutil.NewStringReader("n\n")
	defer func() { ui.StdinReader = orig }()

	status.DiscardChangesMulti([]string{dir})

	if !repo.HasUncommittedChanges(dir) {
		t.Error("expected changes to remain after cancellation")
	}
}
