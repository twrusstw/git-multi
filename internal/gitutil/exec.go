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
