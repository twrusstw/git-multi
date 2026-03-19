package testutil

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// InitRepo creates a temp git repo with one initial commit and returns its path.
func InitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init", "-b", "main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %v\n%s", c, err, out)
		}
	}
	WriteFile(t, dir, "README.md", "init")
	GitMustRun(t, dir, "add", ".")
	GitMustRun(t, dir, "commit", "-m", "init")
	return dir
}

// InitBareRepo creates a bare git repo (acts as a remote) and returns its path.
func InitBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init bare: %v\n%s", err, out)
	}
	return dir
}

// CloneRepo clones src into a new temp dir and returns the clone path.
func CloneRepo(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "clone", src, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
	cmds := [][]string{
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("config %v: %v\n%s", c, err, out)
		}
	}
	return dir
}

// WriteFile writes content to filename inside dir.
func WriteFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", filename, err)
	}
}

// GitMustRun runs a git command in dir and fails the test on error.
func GitMustRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// NewStringReader returns a *bufio.Reader backed by s, for injecting stdin in tests.
func NewStringReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

// GitOutput runs a git command in dir and returns trimmed stdout.
func GitOutput(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
