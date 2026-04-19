package push

import (
	"fmt"
	"os/exec"
	"strings"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
)

func Push(dir, branchName string) {
	if branchName == "" {
		branchName = repo.CurrentBranch(dir)
	}
	label := repo.Label(dir)

	// Check if upstream exists.
	_, err := gitutil.Git(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		// No upstream — push and set tracking.
		fmt.Printf("%s: No upstream for %s, pushing with -u origin.\n", ui.Cyan(label), branchName)
		cmd := exec.Command("git", "push", "-u", "origin", branchName)
		cmd.Dir = dir
		out, pushErr := cmd.CombinedOutput()
		if pushErr != nil {
			fmt.Printf("%s: push failed: %s\n", ui.Cyan(label), strings.TrimSpace(string(out)))
		} else {
			fmt.Printf("%s: pushed %s → origin/%s\n", ui.Cyan(label), branchName, branchName)
		}
		return
	}

	cmd := exec.Command("git", "push")
	cmd.Dir = dir
	out, pushErr := cmd.CombinedOutput()
	if pushErr != nil {
		fmt.Printf("%s: push rejected: %s\n", ui.Cyan(label), strings.TrimSpace(string(out)))
		return
	}
	fmt.Printf("%s: pushed %s\n", ui.Cyan(label), branchName)
}
