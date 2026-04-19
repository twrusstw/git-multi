package pull

import (
	"fmt"
	"strings"
	"sync"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
)

type pullState struct {
	dir   string
	label string
	// "ok", "conflict"
	status        string
	conflictFiles []string
}

// PullAll runs the three-phase cascade across all dirs, then handles conflicts as a group.
func PullAll(dirs []string, branchName string) {
	states := make([]pullState, len(dirs))
	var wg sync.WaitGroup
	for i, dir := range dirs {
		wg.Add(1)
		go func(i int, dir string) {
			defer wg.Done()
			states[i] = cascade(dir, branchName)
		}(i, dir)
	}
	wg.Wait()

	var conflicts []pullState
	for _, s := range states {
		if s.status == "conflict" {
			conflicts = append(conflicts, s)
		}
	}
	if len(conflicts) > 0 {
		resolveConflicts(conflicts, branchName)
	}
}

// PullRebase runs git pull --rebase on each dir concurrently; no cascade.
func PullRebase(dirs []string, branchName string) {
	var wg sync.WaitGroup
	for _, dir := range dirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			label := repo.Label(dir)
			branch := effectiveBranch(dir, branchName)
			args := []string{"pull", "--rebase"}
			if branch != "" {
				args = append(args, "origin", branch)
			}
			if err := gitutil.GitRun(dir, args...); err != nil {
				ui.LockedPrint(func() {
					ui.Errorf("%s: pull --rebase failed\n", ui.Cyan(label))
				})
			}
		}(dir)
	}
	wg.Wait()
}

func cascade(dir, branchName string) pullState {
	label := repo.Label(dir)
	branch := effectiveBranch(dir, branchName)

	args := []string{"pull", "--ff-only"}
	if branch != "" {
		args = append(args, "origin", branch)
	}

	say := func(msg string) {
		ui.LockedPrint(func() { fmt.Printf("%s: %s\n", ui.Cyan(label), msg) })
	}

	// Phase 1: ff-only
	if gitutil.GitOK(dir, args...) {
		say("pulled (ff-only)")
		return pullState{dir: dir, label: label, status: "ok"}
	}

	// Phase 2: stash → pull --ff-only → stash pop.
	// `git stash push` returns exit 0 even on a clean tree — detect the no-op
	// via stdout so a later `stash pop` doesn't pop an unrelated previous stash.
	say("→ ff-only failed, stashing changes...")
	stashOut, stashErr := gitutil.Git(dir, "stash", "push")
	stashed := stashErr == nil && !strings.Contains(stashOut, "No local changes")
	if gitutil.GitOK(dir, args...) {
		if stashed {
			if gitutil.GitOK(dir, "stash", "pop") {
				say("pulled (stash+ff-only+pop)")
				return pullState{dir: dir, label: label, status: "ok"}
			}
			// stash pop conflict
			files := conflictFiles(dir)
			say("conflict after pull")
			return pullState{dir: dir, label: label, status: "conflict", conflictFiles: files}
		}
		say("pulled (ff-only after stash)")
		return pullState{dir: dir, label: label, status: "ok"}
	}
	// Still can't pull — unstash and mark conflict
	if stashed {
		gitutil.GitOK(dir, "stash", "pop")
	}
	say("conflict after pull")
	return pullState{dir: dir, label: label, status: "conflict", conflictFiles: conflictFiles(dir)}
}

func resolveConflicts(conflicts []pullState, branchName string) {
	for _, c := range conflicts {
		fmt.Printf("%s: conflict after pull\n", c.label)
	}
	fmt.Println()
	fmt.Println("All conflicted repos:")
	choice := ui.PromptMenu([]string{
		"reset --hard (discard local changes)",
		"merge",
		"rebase",
		"skip all",
	})

	switch choice {
	case 1: // reset --hard
		for _, c := range conflicts {
			branch := effectiveBranch(c.dir, branchName)
			target := "origin/" + branch
			if branch == "" {
				target = "HEAD"
			}
			if err := gitutil.GitRun(c.dir, "reset", "--hard", target); err != nil {
				ui.Errorf("%s: reset --hard failed\n", ui.Cyan(c.label))
				continue
			}
			fmt.Printf("%s: reset --hard to %s\n", ui.Cyan(c.label), target)
		}
	case 2, 3: // merge or rebase
		verb := "merge"
		continueCmd := "git merge --continue"
		if choice == 3 {
			verb = "rebase"
			continueCmd = "git rebase --continue"
		}
		for _, c := range conflicts {
			branch := effectiveBranch(c.dir, branchName)
			var err error
			if choice == 2 {
				err = gitutil.GitRun(c.dir, "merge", "origin/"+branch)
			} else {
				err = gitutil.GitRun(c.dir, "rebase", "origin/"+branch)
			}
			if err != nil {
				fmt.Printf("%s: %s conflict\n", ui.Cyan(c.label), verb)
				files := conflictFiles(c.dir)
				for _, f := range files {
					fmt.Printf("  M %s\n", f)
				}
				fmt.Printf("  → resolve manually, then run: %s\n", continueCmd)
			}
		}
	case 4: // skip all
		fmt.Println("Skipped all conflicted repos.")
	}
}

func effectiveBranch(dir, branchName string) string {
	if branchName != "" {
		return branchName
	}
	return repo.CurrentBranch(dir)
}

func conflictFiles(dir string) []string {
	out, _ := gitutil.Git(dir, "status", "--porcelain")
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 3 {
			continue
		}
		x, y := line[0], line[1]
		if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}
	return files
}
