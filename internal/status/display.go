package status

import (
	"fmt"
	"strings"
	"sync"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
)

type repoStatus struct {
	label      string
	groupName  string
	branchName string
	pullCount  string
	pushCount  string
	untracked  int
	unstaged   int
	staged     int
}

func collectStatus(dir string) repoStatus {
	label := repo.Label(dir)
	branchName := repo.CurrentBranch(dir)

	var pullCount, pushCount string
	remoteBranch, err := gitutil.Git(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err == nil && remoteBranch != "" {
		counts, _ := gitutil.Git(dir, "rev-list", "--count", "--left-right", "HEAD..."+remoteBranch)
		parts := strings.SplitN(counts, "\t", 2)
		if len(parts) == 2 {
			pushCount = parts[0]
			pullCount = parts[1]
		}
	} else {
		pullCount = "N/A"
		pushCount = "N/A"
	}

	statusOut, _ := gitutil.Git(dir, "status", "--porcelain")
	var untracked, unstaged, staged int
	for _, line := range strings.Split(statusOut, "\n") {
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]
		if x == '?' && y == '?' {
			untracked++
		}
		if y == 'M' || y == 'D' {
			unstaged++
		}
		if x == 'M' || x == 'A' || x == 'D' || x == 'R' || x == 'C' {
			staged++
		}
	}

	remoteURL, _ := gitutil.Git(dir, "remote", "get-url", "origin")
	groupName := extractOwner(remoteURL)

	return repoStatus{
		label:      label,
		groupName:  groupName,
		branchName: branchName,
		pullCount:  pullCount,
		pushCount:  pushCount,
		untracked:  untracked,
		unstaged:   unstaged,
		staged:     staged,
	}
}

func printStatus(s repoStatus, printHeader bool) {
	status := fmt.Sprintf("(⇣%s ⇡%s ?%d !%d +%d)", s.pullCount, s.pushCount, s.untracked, s.unstaged, s.staged)
	if printHeader {
		fmt.Printf("\033[1m%-20s %-30s %-20s %-5s\033[0m\n", "Group", "Repository", "Branch", "Status")
	}
	dirty := s.untracked > 0 || s.unstaged > 0 || s.staged > 0
	if dirty {
		fmt.Printf("%s %-30s %-20s \033[42;30m%s\033[0m\n",
			ui.PadRight(ui.Cyan(s.groupName), 20), s.label, s.branchName, status)
	} else {
		fmt.Printf("%s \033[33m%-30s\033[0m %-20s %s\n",
			ui.PadRight(ui.Cyan(s.groupName), 20), s.label, s.branchName, status)
	}
}

func ShowCurrentAll(dirs []string) {
	results := make([]repoStatus, len(dirs))
	var wg sync.WaitGroup
	for i, dir := range dirs {
		wg.Add(1)
		go func(i int, dir string) {
			defer wg.Done()
			results[i] = collectStatus(dir)
		}(i, dir)
	}
	wg.Wait()

	for i, s := range results {
		printStatus(s, i == 0)
	}
}

func ShowCurrent(dir string, printHeader bool) {
	printStatus(collectStatus(dir), printHeader)
}

// extractOwner pulls the owner/org name from a git remote URL.
// Handles both HTTPS (https://host/owner/repo) and SSH (git@host:owner/repo) forms.
func extractOwner(remoteURL string) string {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	// SSH: git@github.com:owner/repo — distinguished from HTTPS by absence of "://"
	if !strings.Contains(remoteURL, "://") {
		if idx := strings.Index(remoteURL, ":"); idx != -1 {
			parts := strings.Split(remoteURL[idx+1:], "/")
			if len(parts) >= 1 {
				return parts[0]
			}
		}
	}
	// HTTPS: https://github.com/owner/repo
	parts := strings.Split(remoteURL, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return ""
}

func ShowStatus(dir string) {
	label := repo.Label(dir)
	out, _ := gitutil.Git(dir, "status")
	if strings.Contains(out, "nothing to commit") {
		fmt.Printf("%s: No changes to show.\n", ui.Cyan(label))
	} else {
		fmt.Printf("%s:\n%s\n", ui.Cyan(label), out)
	}
}

// DiscardChangesMulti shows a pre-flight summary of all repos with changes,
// asks for a single confirmation, then discards only those repos.
func DiscardChangesMulti(repos []string) {
	type dirtyRepo struct {
		dir   string
		files string
	}

	collected := make([]dirtyRepo, len(repos))
	var wg sync.WaitGroup
	for i, r := range repos {
		wg.Add(1)
		go func(i int, r string) {
			defer wg.Done()
			out, _ := gitutil.Git(r, "status", "--short")
			if strings.TrimSpace(out) != "" {
				collected[i] = dirtyRepo{dir: r, files: out}
			}
		}(i, r)
	}
	wg.Wait()

	var dirty []dirtyRepo
	for _, d := range collected {
		if d.dir != "" {
			dirty = append(dirty, d)
		}
	}
	if len(dirty) == 0 {
		fmt.Println("No changes to discard in any repository.")
		return
	}
	for _, r := range dirty {
		fmt.Printf("%s:\n", ui.Cyan(repo.Label(r.dir)))
		for _, line := range strings.Split(strings.TrimSpace(r.files), "\n") {
			fmt.Printf("  %s\n", line)
		}
	}
	fmt.Println()
	if !ui.PromptYN(fmt.Sprintf("Discard all changes in the above %d repo(s)? This cannot be undone.", len(dirty))) {
		fmt.Println("Cancelled.")
		return
	}
	for _, r := range dirty {
		label := repo.Label(r.dir)
		gitutil.GitRun(r.dir, "reset", "--hard", "HEAD")
		gitutil.GitRun(r.dir, "clean", "-fd")
		fmt.Printf("%s: Discarded.\n", ui.Cyan(label))
	}
}

// DiscardChanges is kept for the runOp dispatch table.
func DiscardChanges(dir string) {
	DiscardChangesMulti([]string{dir})
}
