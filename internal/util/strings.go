package util

import "strings"

// NonEmpty returns only the non-empty, trimmed strings from ss.
func NonEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// NormaliseBranchName strips the "remotes/<remote>/" prefix so that
// "remotes/origin/main" and "main" are treated as the same branch.
func NormaliseBranchName(raw string) string {
	raw = strings.TrimSpace(strings.TrimPrefix(raw, "*"))
	raw = strings.TrimSpace(raw)
	// remotes/origin/HEAD -> origin/main  (symbolic ref line, skip)
	if strings.Contains(raw, " -> ") || strings.HasSuffix(raw, "/HEAD") {
		return ""
	}
	if after, ok := strings.CutPrefix(raw, "remotes/"); ok {
		if idx := strings.Index(after, "/"); idx != -1 {
			return after[idx+1:]
		}
	}
	return raw
}
