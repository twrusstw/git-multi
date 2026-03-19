package branch

import (
	"os"
	"path/filepath"
	"strings"
)

// defaultSkipFiles lists filenames (by basename) that are ignored when deciding whether to create a branch.
var defaultSkipFiles = map[string]bool{
	// Go
	"go.mod": true, "go.sum": true,
	// Node.js
	"package-lock.json": true, "yarn.lock": true,
	"pnpm-lock.yaml": true, "bun.lockb": true,
	// Rust
	"Cargo.lock": true,
	// Python
	"poetry.lock": true, "Pipfile.lock": true,
	// PHP
	"composer.lock": true,
}

// loadSkipSet merges defaultSkipFiles with any entries in <dir>/.gitmulti-skip.
func loadSkipSet(dir string) map[string]bool {
	set := make(map[string]bool, len(defaultSkipFiles))
	for k := range defaultSkipFiles {
		set[k] = true
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitmulti-skip"))
	if err != nil {
		return set
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		set[line] = true
	}
	return set
}

// isSkipFile reports whether path's basename is in the skip set.
func isSkipFile(path string, skipSet map[string]bool) bool {
	return skipSet[filepath.Base(path)]
}
