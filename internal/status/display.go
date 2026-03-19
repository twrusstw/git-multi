package status

import (
	"fmt"
	"strings"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
)

func ShowCurrent(dir string, printHeader bool) {
	label := repo.Label(dir)

	branchName := repo.CurrentBranch(dir)
	remoteBranch, err := gitutil.Git(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	var pullCount, pushCount string
	if err == nil && remoteBranch != "" {
		pull, _ := gitutil.Git(dir, "rev-list", "--count", "HEAD.."+remoteBranch)
		push, _ := gitutil.Git(dir, "rev-list", "--count", remoteBranch+"..HEAD")
		pullCount = pull
		pushCount = push
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

	status := fmt.Sprintf("(⇣%s ⇡%s ?%d !%d +%d)", pullCount, pushCount, untracked, unstaged, staged)

	if printHeader {
		fmt.Printf("\033[1m%-20s %-30s %-20s %-5s\033[0m\n", "Group", "Repository", "Branch", "Status")
	}

	dirty := untracked > 0 || unstaged > 0 || staged > 0
	if dirty {
		fmt.Printf("%s %-30s %-20s \033[42;30m%s\033[0m\n",
			ui.PadRight(ui.Cyan(groupName), 20), label, branchName, status)
	} else {
		fmt.Printf("%s \033[33m%-30s\033[0m %-20s %s\n",
			ui.PadRight(ui.Cyan(groupName), 20), label, branchName, status)
	}
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
	out, _ := gitutil.Git(dir, "status", "--porcelain")
	if strings.TrimSpace(out) != "" {
		fmt.Printf("%s: Status of branch %s:\n", ui.Cyan(label), repo.CurrentBranch(dir))
		gitutil.GitRun(dir, "status")
	} else {
		fmt.Printf("%s: No changes to show.\n", ui.Cyan(label))
	}
}

// DiscardChangesMulti shows a pre-flight summary of all repos with changes,
// asks for a single confirmation, then discards only those repos.
func DiscardChangesMulti(repos []string) {
	type dirtyRepo struct {
		dir   string
		files string
	}
	var dirty []dirtyRepo
	for _, r := range repos {
		out, _ := gitutil.Git(r, "status", "--short")
		if strings.TrimSpace(out) != "" {
			dirty = append(dirty, dirtyRepo{dir: r, files: out})
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
