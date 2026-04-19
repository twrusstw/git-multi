package fetch

import (
	"fmt"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/status"
	"gitmulti/internal/ui"
	"gitmulti/internal/util"
)

func FetchAll(dirs []string) {
	results := util.ParallelMap(dirs, 0, fetchOne)

	labelW, branchW := 10, 15
	for _, r := range results {
		if n := len(r.Label); n > labelW {
			labelW = n
		}
		if n := len(r.Branch); n > branchW {
			branchW = n
		}
	}

	for _, r := range results {
		dirtyStr := "clean"
		if r.Dirty() {
			dirtyStr = ui.Bold("dirty")
		}
		fmt.Printf("%-*s  %-*s  ↑%-3s ↓%-3s  %s\n",
			labelW, r.Label,
			branchW, r.Branch,
			r.Ahead, r.Behind,
			dirtyStr,
		)
	}
}

// fetchOne runs `git fetch` then a single `status --branch --porcelain=v2`
// via status.Collect — replacing the previous three separate git calls
// (fetch + rev-list + status --porcelain).
func fetchOne(dir string) status.Info {
	gitutil.GitOK(dir, "fetch")
	return status.Collect(dir)
}
