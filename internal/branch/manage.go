package branch

import (
	"fmt"
	"strings"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
	"gitmulti/internal/util"
)

type repoWithBranch struct {
	dir   string
	label string
	has   bool
}

func scanForBranch(dirs []string, branchName string) []repoWithBranch {
	return util.ParallelMap(dirs, 0, func(dir string) repoWithBranch {
		return repoWithBranch{
			dir:   dir,
			label: repo.Label(dir),
			has:   repo.BranchExistsLocal(dir, branchName),
		}
	})
}

// Delete deletes branchName in repos that have it. deleteRemote also removes origin/<branchName>.
func Delete(dirs []string, branchName string, deleteRemote bool) {
	scanned := scanForBranch(dirs, branchName)

	fmt.Printf("Deleting branch: %s\n\n", ui.Bold(branchName))
	var targets []repoWithBranch
	for _, r := range scanned {
		if r.has {
			fmt.Printf("  %-20s has branch\n", r.label)
			targets = append(targets, r)
		} else {
			fmt.Printf("  %-20s no branch → skip\n", r.label)
		}
	}
	if len(targets) == 0 {
		fmt.Println("\nBranch not found in any repository.")
		return
	}

	fmt.Printf("\nDelete in %s?\n", joinLabels(targets))
	choice := ui.PromptMenu([]string{"yes to all", "confirm each", "cancel"})
	if choice == 3 {
		fmt.Println("Cancelled.")
		return
	}

	for _, r := range targets {
		if choice == 2 {
			if !ui.PromptYN(fmt.Sprintf("Delete %s in %s?", branchName, r.label)) {
				continue
			}
		}
		doDelete(r.dir, r.label, branchName, "-d", deleteRemote)
	}
}

// ForceDelete is like Delete but uses -D and warns on unmerged branches.
func ForceDelete(dirs []string, branchName string, deleteRemote bool) {
	scanned := scanForBranch(dirs, branchName)

	fmt.Printf("Force-deleting branch: %s\n\n", ui.Bold(branchName))
	var targets []repoWithBranch
	for _, r := range scanned {
		if r.has {
			fmt.Printf("  %-20s has branch\n", r.label)
			targets = append(targets, r)
		} else {
			fmt.Printf("  %-20s no branch → skip\n", r.label)
		}
	}
	if len(targets) == 0 {
		fmt.Println("\nBranch not found in any repository.")
		return
	}

	fmt.Printf("\nForce-delete in %s?\n", joinLabels(targets))
	choice := ui.PromptMenu([]string{"yes to all", "confirm each", "cancel"})
	if choice == 3 {
		fmt.Println("Cancelled.")
		return
	}

	for _, r := range targets {
		out, _ := gitutil.Git(r.dir, "branch", "--merged", "HEAD", "--list", branchName)
		merged := strings.TrimSpace(out) != ""
		if !merged {
			if !ui.PromptYN(fmt.Sprintf("! %s not merged in %s — force delete?", branchName, r.label)) {
				continue
			}
		} else if choice == 2 {
			if !ui.PromptYN(fmt.Sprintf("Force-delete %s in %s?", branchName, r.label)) {
				continue
			}
		}
		doDelete(r.dir, r.label, branchName, "-D", deleteRemote)
	}
}

func doDelete(dir, label, branchName, flag string, deleteRemote bool) {
	if err := gitutil.GitRun(dir, "branch", flag, branchName); err != nil {
		fmt.Printf("%s: delete failed\n", ui.Cyan(label))
		return
	}
	fmt.Printf("%s: deleted %s\n", ui.Cyan(label), branchName)
	if deleteRemote {
		if err := gitutil.GitRun(dir, "push", "origin", "--delete", branchName); err != nil {
			fmt.Printf("%s: remote delete failed\n", ui.Cyan(label))
		} else {
			fmt.Printf("%s: deleted origin/%s\n", ui.Cyan(label), branchName)
		}
	}
}

// Rename renames oldName to newName in repos that have oldName.
func Rename(dirs []string, oldName, newName string) {
	scanned := scanForBranch(dirs, oldName)

	fmt.Printf("Renaming: %s → %s\n\n", ui.Bold(oldName), ui.Bold(newName))
	var targets []repoWithBranch
	for _, r := range scanned {
		if r.has {
			fmt.Printf("  %-20s has branch → will rename\n", r.label)
			targets = append(targets, r)
		} else {
			fmt.Printf("  %-20s no branch  → skip\n", r.label)
		}
	}
	if len(targets) == 0 {
		fmt.Println("\nBranch not found in any repository.")
		return
	}

	if !ui.PromptYN(fmt.Sprintf("\nProceed? Rename %s → %s in %d repo(s)", oldName, newName, len(targets))) {
		fmt.Println("Cancelled.")
		return
	}

	for _, r := range targets {
		if err := gitutil.GitRun(r.dir, "branch", "-m", oldName, newName); err != nil {
			fmt.Printf("%s: rename failed\n", ui.Cyan(r.label))
		} else {
			fmt.Printf("%s: renamed %s → %s\n", ui.Cyan(r.label), oldName, newName)
		}
	}

	fmt.Printf("\nSync remote?\n")
	choice := ui.PromptMenu([]string{"yes to all", "confirm each", "skip"})
	if choice == 3 {
		return
	}
	for _, r := range targets {
		if choice == 2 {
			if !ui.PromptYN(fmt.Sprintf("Sync remote for %s?", r.label)) {
				continue
			}
		}
		if err := gitutil.GitRun(r.dir, "push", "origin", "--delete", oldName); err != nil {
			ui.Errorf("%s: failed to delete remote branch: %v\n", r.label, err)
			continue
		}
		if err := gitutil.GitRun(r.dir, "push", "-u", "origin", newName); err != nil {
			ui.Errorf("%s: failed to push renamed branch: %v\n", r.label, err)
			continue
		}
		fmt.Printf("%s: remote synced\n", ui.Cyan(r.label))
	}
}

func joinLabels(repos []repoWithBranch) string {
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.label
	}
	return strings.Join(names, ", ")
}
