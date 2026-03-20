package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gitmulti/internal/branch"
	"gitmulti/internal/pull"
	"gitmulti/internal/repo"
	"gitmulti/internal/status"
	"gitmulti/internal/ui"
	"gitmulti/internal/validate"
)

var allFlags = []string{"-p", "-pf", "-s", "-sf", "-F", "-b", "-al", "-dc", "-st", "-nb", "-d", "-h"}

// noArgFlags are flags that take no branch/value argument.
var noArgFlags = map[string]bool{"-b": true, "-st": true, "-dc": true, "-h": true, "-al": true}

// parallelOps lists ops that are safe to run concurrently (no per-repo interactive prompts).
var parallelOps = map[string]bool{"-pf": true, "-sf": true, "-F": true, "-st": true, "-nb": true}

func runComplete(prev, cur string) {
	root, _ := os.Getwd()
	switch {
	case prev == "-d":
		for _, name := range repo.FindGitRepoNames(root) {
			if strings.HasPrefix(name, cur) {
				fmt.Println(name)
			}
		}
	case strings.HasPrefix(cur, "-"):
		for _, f := range allFlags {
			if strings.HasPrefix(f, cur) {
				fmt.Println(f)
			}
		}
	case noArgFlags[prev]:
		// no completions
	default:
		for _, name := range branch.ListAllNames(root, cur) {
			fmt.Println(name)
		}
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		ui.Fatalf("no arguments provided.")
	}

	op := args[0]
	if op == "-h" {
		ui.ShowHelp()
		return
	}

	if op == "__complete" {
		rest := args[1:]
		var prev, cur string
		if len(rest) >= 1 {
			prev = rest[0]
		}
		if len(rest) >= 2 {
			cur = rest[1]
		}
		runComplete(prev, cur)
		return
	}

	var branchName, specifiedDir string
	rest := args[1:]
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "-d":
			if i+1 >= len(rest) {
				ui.Fatalf("-d requires a directory path.")
			}
			specifiedDir = filepath.Clean(rest[i+1])
			i++
		default:
			if branchName == "" && !strings.HasPrefix(rest[i], "-") {
				branchName = rest[i]
			}
		}
	}

	// -al is special: iterates internally and exits.
	if op == "-al" {
		if err := validate.Keyword(branchName); err != nil {
			ui.Fatalf("%v", err)
		}
		root, _ := os.Getwd()
		branch.ListAll(root, branchName)
		return
	}

	if err := validate.BranchName(branchName); err != nil {
		ui.Fatalf("%v", err)
	}

	switch op {
	case "-s", "-sf", "-F", "-nb":
		if branchName == "" {
			ui.Fatalf("option %s requires a branch name.", op)
		}
	}

	// Single-repo mode.
	if specifiedDir != "" {
		absDir, err := filepath.Abs(specifiedDir)
		if err != nil || !repo.IsGitRepo(absDir) {
			ui.Fatalf("invalid or non-git directory: %s", specifiedDir)
		}
		if op == "-dc" {
			status.DiscardChangesMulti([]string{absDir})
			return
		}
		runOp(op, absDir, branchName, true)
		return
	}

	// Multi-repo mode.
	root, err := os.Getwd()
	if err != nil {
		ui.Fatalf("cannot determine working directory: %v", err)
	}

	repos := repo.FindGitRepos(root)
	if len(repos) == 0 {
		fmt.Fprintln(os.Stderr, "No git repositories found in current directory.")
		os.Exit(1)
	}

	if op == "-dc" {
		status.DiscardChangesMulti(repos)
		return
	}

	if op == "-b" {
		status.ShowCurrentAll(repos)
		return
	}

	if parallelOps[op] {
		var wg sync.WaitGroup
		for _, r := range repos {
			wg.Add(1)
			go func(r string) {
				defer wg.Done()
				runOp(op, r, branchName, false)
			}(r)
		}
		wg.Wait()
		return
	}

	isFirst := true
	for _, r := range repos {
		runOp(op, r, branchName, isFirst)
		isFirst = false
	}
}

// runOp dispatches a single operation for one repository.
func runOp(op, dir, branchName string, isFirst bool) {
	switch op {
	case "-p":
		pull.Pull(dir, branchName)
	case "-pf":
		pull.PullForce(dir, branchName)
	case "-s":
		branch.Switch(dir, branchName)
	case "-sf":
		branch.SwitchForce(dir, branchName)
	case "-F":
		branch.Find(dir, branchName)
	case "-b":
		status.ShowCurrent(dir, isFirst)
	case "-dc":
		status.DiscardChanges(dir)
	case "-st":
		status.ShowStatus(dir)
	case "-nb":
		branch.CreateIfModified(dir, branchName)
	default:
		ui.Fatalf("unknown option: %s", op)
	}
}
