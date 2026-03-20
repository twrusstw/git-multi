package branch

import (
	"fmt"
	"strings"
	"sync"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
	"gitmulti/internal/util"
)

func Switch(dir, branch string) {
	label := repo.Label(dir)
	cur := repo.CurrentBranch(dir)
	if cur == branch {
		fmt.Printf("%s: Already on branch %s.\n", ui.Cyan(label), branch)
		return
	}

	changedFiles, _ := gitutil.Git(dir, "diff-index", "--name-only", "HEAD", "--")
	if changedFiles != "" {
		fmt.Printf("%s: Uncommitted changes would be overwritten:\n", ui.Cyan(label))
		fmt.Print(changedFiles)
		if !ui.PromptYN(fmt.Sprintf("Discard all changes and switch to branch %s?", branch)) {
			return
		}
		gitutil.GitRun(dir, "reset", "--hard", "HEAD")
	}

	if repo.BranchExistsLocal(dir, branch) || repo.BranchExistsRemote(dir, branch) {
		fmt.Printf("%s: Switching to branch %s.\n", ui.Cyan(label), branch)
		gitutil.GitRun(dir, "checkout", branch)
	} else {
		fmt.Printf("%s: Branch %s not found locally or on remote.\n", ui.Cyan(label), branch)
	}
}

func switchForce(dir, branch string, quiet, localRefs bool) {
	label := repo.Label(dir)
	cur := repo.CurrentBranch(dir)
	if cur == branch {
		if !quiet {
			fmt.Printf("%s: Already on branch %s.\n", ui.Cyan(label), branch)
		}
		return
	}
	var exists bool
	if localRefs {
		exists = repo.BranchExistsLocal(dir, branch) || repo.BranchExistsRemoteLocal(dir, branch)
	} else {
		exists = repo.BranchExistsLocal(dir, branch) || repo.BranchExistsRemote(dir, branch)
	}
	if exists {
		if !quiet {
			fmt.Printf("%s: Switching to branch %s.\n", ui.Cyan(label), branch)
		}
		gitutil.GitRun(dir, "checkout", "-f", branch)
	}
}

func SwitchForce(dir, branch string)      { switchForce(dir, branch, false, false) }
func SwitchForceQuiet(dir, branch string) { switchForce(dir, branch, true, true) }

func Find(dir, keyword string) {
	label := repo.Label(dir)
	if repo.BranchExistsLocal(dir, keyword) {
		fmt.Printf("%s\n", ui.Green("Branch found in "+label))
		return
	}
	out, _ := gitutil.Git(dir, "branch", "-a", "-r", "--list", "*"+keyword+"*")
	lines := util.NonEmpty(strings.Split(out, "\n"))
	if len(lines) > 0 {
		fmt.Printf("%s: Exact branch not found, similar branches:\n", ui.Red(label))
		for _, l := range lines {
			fmt.Println(strings.TrimSpace(strings.TrimPrefix(l, "*")))
		}
	} else {
		fmt.Printf("%s: Exact branch not found.\n", ui.Red(label))
	}
}

func ListAllNames(root, keyword string) []string {
	args := []string{"branch", "--all"}
	if keyword != "" {
		args = append(args, "--list", "*"+keyword+"*")
	}

	repos := repo.FindGitRepos(root)
	results := make([][]string, len(repos))
	var wg sync.WaitGroup
	for i, r := range repos {
		wg.Add(1)
		go func(i int, r string) {
			defer wg.Done()
			out, err := gitutil.Git(r, args...)
			if err != nil {
				return
			}
			var names []string
			for _, line := range strings.Split(out, "\n") {
				if name := util.NormaliseBranchName(line); name != "" {
					names = append(names, name)
				}
			}
			results[i] = names
		}(i, r)
	}
	wg.Wait()

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

func CreateIfModified(dir, branch string) {
	label := repo.Label(dir)

	out, _ := gitutil.Git(dir, "status", "--porcelain")
	var modified []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) >= 2 && (line[0] == 'A' || line[0] == 'M' || line[0] == 'D' ||
			line[1] == 'A' || line[1] == 'M' || line[1] == 'D') {
			modified = append(modified, strings.TrimSpace(line[3:]))
		}
	}

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
		fmt.Printf("%s: Error: branch %q already exists.\n", label, branch)
		return
	}

	fmt.Printf("%s: Uncommitted changes detected. Creating branch %s.\n", label, branch)
	if err := gitutil.GitRun(dir, "checkout", "-b", branch); err != nil {
		fmt.Printf("%s: Failed to create branch %s.\n", label, branch)
	} else {
		fmt.Printf("%s: Switched to new branch %s.\n", label, branch)
	}
}
