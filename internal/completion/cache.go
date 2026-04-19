// Package completion provides cached lookups for shell completion hot paths.
//
// Each Tab press re-execs `gitmulti __complete`, so any work done here is paid
// per keystroke. Without a cache, completing branch names across 50 repos forks
// git 50+ times per Tab. A short file-backed TTL keeps interactive response
// snappy while still picking up newly-created branches within a few seconds.
package completion

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitmulti/internal/branch"
	"gitmulti/internal/repo"
)

const (
	branchTTL = 3 * time.Second
	repoTTL   = 10 * time.Second
)

// BranchNames returns all unique branch names across immediate sub-repos of root,
// served from a short-lived file cache.
func BranchNames(root string) []string {
	if names, ok := readCache(cachePath(root, "branches"), branchTTL); ok {
		return names
	}
	names := branch.ListAllNames(root, "")
	writeCache(cachePath(root, "branches"), names)
	return names
}

// RepoNames returns the immediate sub-repo directory names under root, cached.
func RepoNames(root string) []string {
	if names, ok := readCache(cachePath(root, "repos"), repoTTL); ok {
		return names
	}
	names := repo.FindGitRepoNames(root)
	writeCache(cachePath(root, "repos"), names)
	return names
}

func cachePath(root, kind string) string {
	sum := sha1.Sum([]byte(root))
	return filepath.Join(os.TempDir(), fmt.Sprintf("gitmulti-cache-%s-%x", kind, sum[:8]))
}

func readCache(path string, ttl time.Duration) ([]string, bool) {
	info, err := os.Stat(path)
	if err != nil || time.Since(info.ModTime()) > ttl {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return []string{}, true
	}
	return strings.Split(trimmed, "\n"), true
}

func writeCache(path string, names []string) {
	_ = os.WriteFile(path, []byte(strings.Join(names, "\n")+"\n"), 0o644)
}
