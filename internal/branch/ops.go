package branch

import (
	"bytes"
	"fmt"
	"strings"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
	"gitmulti/internal/util"
)

type switchMode int

const (
	switchModePrompt switchMode = iota
	switchModeStash
	switchModeDiscard
	switchModeCancel
)

// isModifiedStatus reports whether a porcelain XY status code represents a
// relevant change (added, modified, deleted, renamed or copied in either the
// index or the worktree).
func isModifiedStatus(x, y byte) bool {
	return isChangeCode(x) || isChangeCode(y)
}

func isChangeCode(c byte) bool {
	switch c {
	case 'A', 'M', 'D', 'R', 'C':
		return true
	}
	return false
}

// parsePorcelainZ parses `git status --porcelain -z` output into paths.
// The -z format uses NUL separators and handles filenames with whitespace correctly.
// Rename/copy entries are followed by an extra NUL-delimited origin path which we skip.
func parsePorcelainZ(data []byte) []string {
	var paths []string
	i := 0
	for i < len(data) {
		end := bytes.IndexByte(data[i:], 0)
		if end < 0 {
			break
		}
		entry := data[i : i+end]
		i += end + 1
		if len(entry) < 3 {
			continue
		}
		x, y := entry[0], entry[1]
		path := string(entry[3:])
		if x == 'R' || x == 'C' {
			if end2 := bytes.IndexByte(data[i:], 0); end2 >= 0 {
				i += end2 + 1
			}
		}
		if isModifiedStatus(x, y) {
			paths = append(paths, path)
		}
	}
	return paths
}

func Switch(dir, branch string) {
	switchWithMode(dir, branch, switchModePrompt)
}

func SwitchStash(dir, branch string) {
	switchWithMode(dir, branch, switchModeStash)
}

func SwitchDiscard(dir, branch string) {
	switchWithMode(dir, branch, switchModeDiscard)
}

func switchWithMode(dir, branch string, mode switchMode) {
	label := repo.Label(dir)
	cur := repo.CurrentBranch(dir)
	if cur == branch {
		fmt.Printf("%s: Already on branch %s.\n", ui.Cyan(label), branch)
		return
	}

	// Only consult local refs here: git checkout can only DWIM-create a
	// tracking branch from a cached remote-tracking ref, so asking the
	// network via ls-remote adds no capability — just serial latency.
	// Run `gitmulti fetch` first if the remote-tracking ref is stale.
	if !repo.BranchExistsLocal(dir, branch) && !repo.BranchExistsRemoteLocal(dir, branch) {
		fmt.Printf("%s: Branch %s not found locally (fetch first?).\n", ui.Cyan(label), branch)
		return
	}

	changedFiles, _ := gitutil.Git(dir, "status", "--short")
	if changedFiles != "" {
		fmt.Printf("%s: Uncommitted changes detected:\n", ui.Cyan(label))
		fmt.Print(changedFiles)
		if !strings.HasSuffix(changedFiles, "\n") {
			fmt.Println()
		}
		fmt.Println()
		if mode == switchModePrompt {
			mode = promptSwitchMode(branch)
		}
		switch mode { //nolint:exhaustive // switchModePrompt is resolved above
		case switchModeStash:
			stashAndSwitch(dir, label, branch)
		case switchModeDiscard:
			discardAndSwitch(dir, label, branch)
		case switchModeCancel:
			fmt.Printf("%s: Cancelled.\n", ui.Cyan(label))
		default:
			ui.Errorf("%s: invalid switch mode\n", ui.Cyan(label))
		}
		return
	}

	fmt.Printf("%s: Switching to branch %s.\n", ui.Cyan(label), branch)
	if err := gitutil.GitRun(dir, "checkout", branch); err != nil {
		ui.Errorf("%s: checkout failed: %v\n", label, err)
	}
}

func promptSwitchMode(branch string) switchMode {
	choice := ui.PromptMenu([]string{
		fmt.Sprintf("stash and reapply before switching to %s", branch),
		fmt.Sprintf("discard all changes and switch to %s", branch),
		"cancel",
	})
	switch choice {
	case 1:
		return switchModeStash
	case 2:
		return switchModeDiscard
	default:
		return switchModeCancel
	}
}

func stashAndSwitch(dir, label, branch string) {
	fmt.Printf("%s: Stashing changes before switching to branch %s.\n", ui.Cyan(label), branch)
	if _, err := gitutil.GitCombined(dir, "stash", "push", "-u"); err != nil {
		ui.Errorf("%s: stash failed\n", ui.Cyan(label))
		return
	}

	fmt.Printf("%s: Switching to branch %s.\n", ui.Cyan(label), branch)
	if err := gitutil.GitRun(dir, "checkout", branch); err != nil {
		if _, popErr := gitutil.GitCombined(dir, "stash", "pop"); popErr != nil {
			ui.Errorf("%s: checkout failed; stash restore also failed\n", ui.Cyan(label))
		} else {
			ui.Errorf("%s: checkout failed; restored stashed changes on original branch\n", ui.Cyan(label))
		}
		return
	}

	if _, err := gitutil.GitCombined(dir, "stash", "pop"); err != nil {
		printSwitchReapplyConflict(label, dir)
	}
}

func discardAndSwitch(dir, label, branch string) {
	fmt.Printf("%s: Discarding changes before switching to branch %s.\n", ui.Cyan(label), branch)
	if _, err := gitutil.GitCombined(dir, "reset", "--hard", "HEAD"); err != nil {
		ui.Errorf("%s: reset --hard failed\n", ui.Cyan(label))
		return
	}
	if _, err := gitutil.GitCombined(dir, "clean", "-fd"); err != nil {
		ui.Errorf("%s: clean -fd failed\n", ui.Cyan(label))
		return
	}

	fmt.Printf("%s: Switching to branch %s.\n", ui.Cyan(label), branch)
	if err := gitutil.GitRun(dir, "checkout", branch); err != nil {
		ui.Errorf("%s: checkout failed: %v\n", label, err)
	}
}

func printSwitchReapplyConflict(label, dir string) {
	statusOut, _ := gitutil.Git(dir, "status", "--porcelain")
	ui.Errorf("%s: stash reapply conflict\n", ui.Cyan(label))
	for _, line := range strings.Split(statusOut, "\n") {
		if len(line) >= 2 && (line[0] == 'U' || line[1] == 'U' || (line[0] == 'A' && line[1] == 'A')) {
			ui.Errorf("  %s\n", strings.TrimSpace(line[3:]))
		}
	}
	ui.Errorf("  → resolve manually, then: git add <file> && git stash drop\n")
}

func Find(dir, keyword string) {
	label := repo.Label(dir)
	if repo.BranchExistsLocal(dir, keyword) {
		ui.LockedPrint(func() {
			fmt.Printf("%s\n", ui.Green("Branch found in "+label))
		})
		return
	}
	out, _ := gitutil.Git(dir, "branch", "-a", "-r", "--list", "*"+keyword+"*")
	lines := util.NonEmpty(strings.Split(out, "\n"))
	ui.LockedPrint(func() {
		if len(lines) > 0 {
			fmt.Printf("%s: Exact branch not found, similar branches:\n", ui.Red(label))
			for _, l := range lines {
				fmt.Println(strings.TrimSpace(strings.TrimPrefix(l, "*")))
			}
		} else {
			fmt.Printf("%s: Exact branch not found.\n", ui.Red(label))
		}
	})
}

type findResult struct {
	label    string
	exact    bool
	upstream string
	similar  []string
}

func FindAll(dirs []string, keyword string) {
	results := util.ParallelMap(dirs, 0, func(dir string) findResult {
		return findInRepo(dir, keyword)
	})
	printFindResults(keyword, results)
}

func findInRepo(dir, keyword string) findResult {
	result := findResult{label: repo.Label(dir)}
	if repo.BranchExistsLocal(dir, keyword) {
		result.exact = true
		result.upstream = branchUpstream(dir, keyword)
		return result
	}
	result.similar = similarBranches(dir, keyword)
	return result
}

func branchUpstream(dir, branchName string) string {
	out, err := gitutil.Git(dir, "rev-parse", "--abbrev-ref", branchName+"@{upstream}")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func similarBranches(dir, keyword string) []string {
	out, err := gitutil.Git(dir, "branch", "--all", "--list", "*"+keyword+"*")
	if err != nil {
		return nil
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if name := util.NormaliseBranchName(line); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func printFindResults(keyword string, results []findResult) {
	var ready, localOnly, similar []findResult
	noMatch := 0
	for _, r := range results {
		switch {
		case r.exact && r.upstream != "":
			ready = append(ready, r)
		case r.exact:
			localOnly = append(localOnly, r)
		case len(r.similar) > 0:
			similar = append(similar, r)
		default:
			noMatch++
		}
	}

	fmt.Printf("Find branch: %s\n\n", keyword)
	labelW := findLabelWidth(results)
	printFindExactGroup("Ready to use", ready, labelW, func(r findResult) string {
		return r.upstream
	})
	printFindExactGroup("Local only", localOnly, labelW, func(findResult) string {
		return "no upstream"
	})
	printFindSimilarGroup(similar)

	exact := len(ready) + len(localOnly)
	if len(ready) > 0 || len(localOnly) > 0 || len(similar) > 0 {
		fmt.Println()
	}
	fmt.Printf(
		"Summary: exact %d, with upstream %d, no upstream %d, similar %d, no match %d\n",
		exact,
		len(ready),
		len(localOnly),
		len(similar),
		noMatch,
	)
}

func findLabelWidth(results []findResult) int {
	width := 0
	for _, r := range results {
		if len(r.label) > width {
			width = len(r.label)
		}
	}
	return width
}

func printFindExactGroup(title string, results []findResult, labelW int, detail func(findResult) string) {
	if len(results) == 0 {
		return
	}
	fmt.Println(title)
	for _, r := range results {
		fmt.Printf("  %-*s  %s\n", labelW, r.label, detail(r))
	}
	fmt.Println()
}

func printFindSimilarGroup(results []findResult) {
	if len(results) == 0 {
		return
	}
	fmt.Println("Similar matches")
	for _, r := range results {
		fmt.Printf("  %s\n", r.label)
		names, more := truncateSimilarBranches(r.similar, 3)
		for _, name := range names {
			fmt.Printf("    %s\n", name)
		}
		if more > 0 {
			fmt.Printf("    ... +%d more\n", more)
		}
	}
}

func truncateSimilarBranches(names []string, limit int) ([]string, int) {
	if len(names) <= limit {
		return names, 0
	}
	return names[:limit], len(names) - limit
}

func ListAllNames(root, keyword string) []string {
	args := []string{"branch", "--all"}
	if keyword != "" {
		args = append(args, "--list", "*"+keyword+"*")
	}

	repos := repo.FindGitRepos(root)
	results := util.ParallelMap(repos, 0, func(r string) []string {
		out, err := gitutil.Git(r, args...)
		if err != nil {
			return nil
		}
		var names []string
		for _, line := range strings.Split(out, "\n") {
			if name := util.NormaliseBranchName(line); name != "" {
				names = append(names, name)
			}
		}
		return names
	})

	seen := map[string]bool{}
	var names []string
	for _, batch := range results {
		for _, name := range batch {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}

func ListAll(root, keyword string) {
	for _, name := range ListAllNames(root, keyword) {
		fmt.Println(name)
	}
}

func ListAllGrouped(root, keyword string) {
	args := []string{"branch", "--all"}
	if keyword != "" {
		args = append(args, "--list", "*"+keyword+"*")
	}

	repos := repo.FindGitRepos(root)
	type entry struct {
		label    string
		branches []string
	}
	results := util.ParallelMap(repos, 0, func(r string) entry {
		out, err := gitutil.Git(r, args...)
		if err != nil {
			return entry{}
		}
		var names []string
		for _, line := range strings.Split(out, "\n") {
			if name := util.NormaliseBranchName(line); name != "" {
				names = append(names, name)
			}
		}
		return entry{label: repo.Label(r), branches: names}
	})

	for _, e := range results {
		if len(e.branches) == 0 {
			continue
		}
		fmt.Println(ui.Cyan(e.label))
		for _, b := range e.branches {
			fmt.Println("  " + b)
		}
	}
}

func CreateIfModified(dir, branch string) {
	label := repo.Label(dir)

	out, _ := gitutil.GitBytes(dir, "status", "--porcelain=v1", "-z")
	modified := parsePorcelainZ(out)

	if len(modified) == 0 {
		fmt.Printf("%s: No uncommitted changes. No need to create a new branch.\n", label)
		return
	}

	skipSet := loadSkipSet(dir)

	allSkipped := true
	for _, f := range modified {
		if !isSkipFile(f, skipSet) {
			allSkipped = false
			break
		}
	}
	if allSkipped {
		fmt.Printf("%s: All changes are in skipped files. No new branch created.\n", label)
		return
	}

	if repo.BranchExistsLocal(dir, branch) {
		ui.Errorf("%s: Error: branch %q already exists.\n", label, branch)
		return
	}

	fmt.Printf("%s: Uncommitted changes detected. Creating branch %s.\n", label, branch)
	if err := gitutil.GitRun(dir, "checkout", "-b", branch); err != nil {
		ui.Errorf("%s: Failed to create branch %s.\n", label, branch)
	} else {
		fmt.Printf("%s: Switched to new branch %s.\n", label, branch)
	}
}
