package pull

import (
	"fmt"

	"gitmulti/internal/branch"
	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
)

func Pull(dir, branchName string) {
	if branchName == "" {
		branchName = repo.CurrentBranch(dir)
	}
	label := repo.Label(dir)
	fmt.Printf("%s: Pulling branch %s.\n", ui.Cyan(label), branchName)

	gitutil.GitRun(dir, "fetch", "--all") //nolint — errors printed by git

	if err := gitutil.GitRun(dir, "checkout", branchName); err != nil {
		fmt.Printf("%s: Could not checkout branch %s.\n", ui.Cyan(label), branchName)
		return
	}

	if err := gitutil.GitRun(dir, "pull"); err == nil {
		fmt.Printf("%s: Pulled changes from branch %s.\n", ui.Cyan(label), branchName)
	} else {
		fmt.Printf("%s: Stashing local changes and pulling branch %s.\n", ui.Cyan(label), branchName)
		stashed := gitutil.GitOK(dir, "stash")
		if err2 := gitutil.GitRun(dir, "pull"); err2 == nil {
			fmt.Printf("%s: Pulled changes from branch %s.\n", ui.Cyan(label), branchName)
			if stashed {
				if popErr := gitutil.GitRun(dir, "stash", "pop"); popErr != nil {
					fmt.Printf("%s: WARNING: stash pop has conflicts — repo left in conflict state.\n", ui.Cyan(label))
					fmt.Printf("  Conflicting files:\n")
					gitutil.GitRun(dir, "diff", "--name-only", "--diff-filter=U")
					fmt.Printf("  To resolve : cd %s && fix conflicts, then: git add . && git stash drop\n", dir)
					fmt.Printf("  To abandon : cd %s && git checkout . && git stash drop\n", dir)
				}
			}
		} else {
			if ui.PromptYN("Do you want to force pull and discard all changes? This cannot be undone.") {
				gitutil.GitRun(dir, "reset", "--hard", "origin/"+branchName)
				fmt.Printf("%s: Force pulled branch %s.\n", ui.Cyan(label), branchName)
			} else {
				fmt.Printf("%s: Operation cancelled.\n", ui.Cyan(label))
			}
		}
	}
	fmt.Println()
}

func PullForce(dir, branchName string) {
	if branchName == "" {
		branchName = repo.CurrentBranch(dir)
	}
	label := repo.Label(dir)
	fmt.Printf("%s: Force pulling branch %s.\n", ui.Cyan(label), branchName)

	gitutil.GitRun(dir, "fetch", "--all")
	branch.SwitchForceQuiet(dir, branchName)
	gitutil.GitRun(dir, "reset", "--hard", "origin/"+branchName)

	fmt.Printf("%s: Force pulled branch %s.\n", ui.Cyan(label), branchName)
	fmt.Println()
}
