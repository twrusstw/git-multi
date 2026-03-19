package validate

import (
	"fmt"
	"regexp"
	"strings"
)

// validBranchName enforces safe branch name characters, preventing command injection.
// Git branch names may contain alphanumeric, '.', '-', '_', and '/' (for namespacing).
var validBranchName = regexp.MustCompile(`^[a-zA-Z0-9._/\-]+$`)

// BranchName returns an error if the name is unsafe.
func BranchName(name string) error {
	if name == "" {
		return nil
	}
	if !validBranchName.MatchString(name) {
		return fmt.Errorf("invalid branch name %q: only alphanumeric, '.', '-', '_', '/' allowed", name)
	}
	// Reject path traversal attempts.
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid branch name %q: must not contain '..'", name)
	}
	return nil
}

// Keyword enforces safe keyword characters for branch search.
func Keyword(kw string) error {
	if kw == "" {
		return nil
	}
	if !validBranchName.MatchString(kw) {
		return fmt.Errorf("invalid keyword %q: only alphanumeric, '.', '-', '_', '/' allowed", kw)
	}
	return nil
}
