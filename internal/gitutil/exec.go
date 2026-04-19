package gitutil

import (
	"os"
	"os/exec"
	"strings"
)

// Git runs a git command in dir and returns trimmed stdout.
// All arguments are passed as explicit strings — no shell interpolation.
func Git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// GitBytes runs a git command in dir and returns raw stdout bytes (no trimming).
// Use this when the output contains NUL separators (e.g. `--porcelain -z`).
func GitBytes(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Output()
}

// GitRun runs a git command attached to the terminal (for interactive output).
func GitRun(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GitOK runs git and returns only the exit success/failure.
func GitOK(dir string, args ...string) bool {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run() == nil
}
