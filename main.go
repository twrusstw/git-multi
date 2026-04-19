package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitmulti/internal/branch"
	"gitmulti/internal/completion"
	"gitmulti/internal/fetch"
	"gitmulti/internal/pull"
	"gitmulti/internal/push"
	"gitmulti/internal/repo"
	"gitmulti/internal/stash"
	"gitmulti/internal/status"
	"gitmulti/internal/ui"
	"gitmulti/internal/validate"
)

var subcommands = []string{
	"pull", "push", "fetch",
	"switch", "branch", "status",
	"stash", "discard",
}

func runComplete(tokens []string) {
	root, _ := os.Getwd()

	if len(tokens) == 0 {
		for _, s := range subcommands {
			fmt.Println(s)
		}
		return
	}

	cur := tokens[len(tokens)-1]

	// If completing -C argument, list repo directory names.
	if len(tokens) >= 2 && tokens[len(tokens)-2] == "-C" {
		for _, name := range completion.RepoNames(root) {
			if strings.HasPrefix(name, cur) {
				fmt.Println(name)
			}
		}
		return
	}

	// Strip -C <path> pairs to get clean subcommand tokens.
	clean := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "-C" {
			i++ // skip path value
		} else {
			clean = append(clean, tokens[i])
		}
	}
	if len(clean) == 0 {
		return
	}

	// One token means user is still typing the subcommand word.
	if len(clean) == 1 {
		for _, s := range subcommands {
			if strings.HasPrefix(s, cur) {
				fmt.Println(s)
			}
		}
		return
	}

	sub := clean[0]
	rest := clean[1:] // rest[len-1] == cur

	printBranches := func() {
		for _, name := range completion.BranchNames(root) {
			if strings.HasPrefix(name, cur) {
				fmt.Println(name)
			}
		}
	}

	switch sub {
	case "pull":
		if len(rest) == 1 {
			printBranches()
			if strings.HasPrefix("--rebase", cur) {
				fmt.Println("--rebase")
			}
		} else if rest[0] == "--rebase" {
			printBranches()
		}

	case "push":
		printBranches()

	case "switch":
		if len(rest) == 1 {
			printBranches()
			for _, f := range []string{"-f", "-c"} {
				if strings.HasPrefix(f, cur) {
					fmt.Println(f)
				}
			}
		} else if rest[0] == "-f" && len(rest) == 2 {
			printBranches()
		}

	case "branch":
		if len(rest) == 1 {
			for _, flag := range []string{"-a", "--find", "-d", "-D", "-m"} {
				if strings.HasPrefix(flag, cur) {
					fmt.Println(flag)
				}
			}
		} else if len(rest) == 2 {
			switch rest[0] {
			case "--find", "-d", "-D", "-m":
				printBranches()
			}
		}
		// branch -m <old> <TAB>: new name, no completion

	case "stash":
		if len(rest) == 1 {
			for _, s := range []string{"pop", "apply", "list"} {
				if strings.HasPrefix(s, cur) {
					fmt.Println(s)
				}
			}
		}
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		ui.ShowHelp()
		return
	}

	if args[0] == "__complete" {
		runComplete(args[1:])
		return
	}

	// Extract -C <path> flag (can appear anywhere).
	var specifiedDir string
	var filtered []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-C" {
			if i+1 >= len(args) {
				ui.Fatalf("-C requires a directory path.")
			}
			specifiedDir = filepath.Clean(args[i+1])
			i++
		} else {
			filtered = append(filtered, args[i])
		}
	}
	args = filtered

	if len(args) == 0 {
		ui.Fatalf("no subcommand provided.")
	}

	sub := args[0]
	rest := args[1:]

	if sub == "help" || sub == "-h" || sub == "--help" {
		ui.ShowHelp()
		return
	}

	// Resolve working set of repos.
	var repos []string
	if specifiedDir != "" {
		absDir, err := filepath.Abs(specifiedDir)
		if err != nil || !repo.IsGitRepo(absDir) {
			ui.Fatalf("invalid or non-git directory: %s", specifiedDir)
		}
		repos = []string{absDir}
	} else {
		root, err := os.Getwd()
		if err != nil {
			ui.Fatalf("cannot determine working directory: %v", err)
		}
		repos = repo.FindGitRepos(root)
		if len(repos) == 0 {
			fmt.Fprintln(os.Stderr, "No git repositories found in current directory.")
			os.Exit(1)
		}
	}

	switch sub {
	case "pull":
		rebase := len(rest) > 0 && rest[0] == "--rebase"
		branchName := ""
		for _, a := range rest {
			if !strings.HasPrefix(a, "-") {
				branchName = a
				break
			}
		}
		if err := validate.BranchName(branchName); err != nil {
			ui.Fatalf("%v", err)
		}
		if rebase {
			pull.PullRebase(repos, branchName)
		} else {
			pull.PullAll(repos, branchName)
		}

	case "push":
		branchName := argOrEmpty(rest)
		if err := validate.BranchName(branchName); err != nil {
			ui.Fatalf("%v", err)
		}
		for _, r := range repos {
			push.Push(r, branchName)
		}

	case "fetch":
		fetch.FetchAll(repos)

	case "switch":
		if len(rest) == 0 {
			ui.Fatalf("switch requires a branch name.")
		}
		switch rest[0] {
		case "-f":
			if len(rest) < 2 {
				ui.Fatalf("switch -f requires a branch name.")
			}
			branchName := rest[1]
			if err := validate.BranchName(branchName); err != nil {
				ui.Fatalf("%v", err)
			}
			for _, r := range repos {
				branch.SwitchForce(r, branchName)
			}
		case "-c":
			if len(rest) < 2 {
				ui.Fatalf("switch -c requires a branch name.")
			}
			branchName := rest[1]
			if err := validate.BranchName(branchName); err != nil {
				ui.Fatalf("%v", err)
			}
			for _, r := range repos {
				branch.CreateIfModified(r, branchName)
			}
		default:
			branchName := rest[0]
			if err := validate.BranchName(branchName); err != nil {
				ui.Fatalf("%v", err)
			}
			for _, r := range repos {
				branch.Switch(r, branchName)
			}
		}

	case "branch":
		root, _ := os.Getwd()
		if specifiedDir != "" {
			root = specifiedDir
		}
		if len(rest) == 0 {
			status.ShowCurrentAll(repos)
			return
		}
		switch rest[0] {
		case "-a":
			keyword := argOrEmpty(rest[1:])
			if err := validate.Keyword(keyword); err != nil {
				ui.Fatalf("%v", err)
			}
			branch.ListAll(root, keyword)
		case "--find":
			if len(rest) < 2 {
				ui.Fatalf("branch --find requires a keyword.")
			}
			keyword := rest[1]
			for _, r := range repos {
				branch.Find(r, keyword)
			}
		case "-d":
			if len(rest) < 2 {
				ui.Fatalf("branch -d requires a branch name.")
			}
			deleteRemote := containsFlag(rest, "--remote")
			branch.Delete(repos, rest[1], deleteRemote)
		case "-D":
			if len(rest) < 2 {
				ui.Fatalf("branch -D requires a branch name.")
			}
			deleteRemote := containsFlag(rest, "--remote")
			branch.ForceDelete(repos, rest[1], deleteRemote)
		case "-m":
			if len(rest) < 3 {
				ui.Fatalf("branch -m requires <old> <new>.")
			}
			branch.Rename(repos, rest[1], rest[2])
		default:
			ui.Fatalf("unknown branch flag: %s", rest[0])
		}

	case "status":
		for _, r := range repos {
			status.ShowStatus(r)
		}

	case "stash":
		if len(rest) == 0 {
			stash.Stash(repos)
			return
		}
		switch rest[0] {
		case "pop":
			stash.Pop(repos)
		case "apply":
			stash.Apply(repos)
		case "list":
			stash.List(repos)
		default:
			ui.Fatalf("unknown stash subcommand: %s", rest[0])
		}

	case "discard":
		status.DiscardChangesMulti(repos)

	default:
		ui.Fatalf("unknown subcommand: %s", sub)
	}
}

func argOrEmpty(args []string) string {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}

func containsFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}
