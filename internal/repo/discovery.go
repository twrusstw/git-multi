package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gitmulti/internal/gitutil"
)

// CurrentBranch returns the name of the current branch in dir.
func CurrentBranch(dir string) string {
	b, _ := gitutil.Git(dir, "branch", "--show-current")
	return b
}

// HasUncommittedChanges reports whether the working tree has any uncommitted changes.
func HasUncommittedChanges(dir string) bool {
	cmd := exec.Command("git", "diff-index", "--quiet", "HEAD", "--")
	cmd.Dir = dir
	return cmd.Run() != nil // non-zero exit = changes exist
}

// BranchExistsLocal reports whether branch exists in the local repo.
func BranchExistsLocal(dir, branch string) bool {
	return gitutil.GitOK(dir, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
}

// BranchExistsRemote reports whether branch exists on origin via live network query.
func BranchExistsRemote(dir, branch string) bool {
	return gitutil.GitOK(dir, "ls-remote", "--exit-code", "--heads", "origin", branch)
}

// BranchExistsRemoteLocal reports whether a remote branch exists using locally cached
// remote-tracking refs. Use only after git fetch has already been performed.
func BranchExistsRemoteLocal(dir, branch string) bool {
	return gitutil.GitOK(dir, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
}

// FindGitRepos returns absolute paths to immediate subdirectories that are git repos.
func FindGitRepos(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading directory %s: %v\n", root, err)
		return nil
	}
	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		abs := filepath.Join(root, e.Name())
		if info, err := os.Stat(filepath.Join(abs, ".git")); err == nil && info.IsDir() {
			repos = append(repos, abs)
		}
	}
	return repos
}

// FindGitRepoNames returns the directory names (not full paths) of immediate subdirectories that are git repos.
func FindGitRepoNames(root string) []string {
	repos := FindGitRepos(root)
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = filepath.Base(r)
	}
	return names
}

// Label returns the last path segment (directory name) for display.
func Label(dir string) string {
	return filepath.Base(dir)
}

// IsGitRepo reports whether dir contains a .git directory.
func IsGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}
