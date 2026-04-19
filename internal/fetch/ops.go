package fetch

import (
	"fmt"
	"strings"
	"sync"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
)

type result struct {
	label  string
	branch string
	ahead  string
	behind string
	dirty  bool
}

func FetchAll(dirs []string) {
	results := make([]result, len(dirs))
	var wg sync.WaitGroup
	for i, dir := range dirs {
		wg.Add(1)
		go func(i int, dir string) {
			defer wg.Done()
			results[i] = fetchOne(dir)
		}(i, dir)
	}
	wg.Wait()

	labelW, branchW := 10, 15
	for _, r := range results {
		if n := len(r.label); n > labelW {
			labelW = n
		}
		if n := len(r.branch); n > branchW {
			branchW = n
		}
	}

	for _, r := range results {
		dirtyStr := "clean"
		if r.dirty {
			dirtyStr = ui.Bold("dirty")
		}
		fmt.Printf("%-*s  %-*s  ↑%-3s ↓%-3s  %s\n",
			labelW, r.label,
			branchW, r.branch,
			r.ahead, r.behind,
			dirtyStr,
		)
	}
}

func fetchOne(dir string) result {
	label := repo.Label(dir)
	branch := repo.CurrentBranch(dir)

	gitutil.GitOK(dir, "fetch")

	ahead, behind := "N/A", "N/A"
	counts, err := gitutil.Git(dir, "rev-list", "--count", "--left-right", "HEAD...@{u}")
	if err == nil {
		parts := strings.SplitN(counts, "\t", 2)
		if len(parts) == 2 {
			ahead = parts[0]
			behind = parts[1]
		}
	}

	statusOut, _ := gitutil.Git(dir, "status", "--porcelain")
	dirty := strings.TrimSpace(statusOut) != ""

	return result{label: label, branch: branch, ahead: ahead, behind: behind, dirty: dirty}
}
