package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gitmulti/internal/branch"
	"gitmulti/internal/command"
	"gitmulti/internal/completion"
	"gitmulti/internal/fetch"
	"gitmulti/internal/pull"
	"gitmulti/internal/push"
	"gitmulti/internal/repo"
	"gitmulti/internal/stash"
	"gitmulti/internal/status"
	"gitmulti/internal/ui"
)

var registry = map[string]*command.Command{
	"pull":    pull.Cmd(),
	"push":    push.Cmd(),
	"fetch":   fetch.Cmd(),
	"switch":  branch.SwitchCmd(),
	"branch":  branch.Cmd(),
	"status":  status.Cmd(),
	"stash":   stash.Cmd(),
	"discard": status.DiscardCmd(),
}

func subcommandNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func runComplete(tokens []string) {
	root, _ := os.Getwd()

	if len(tokens) == 0 {
		for _, s := range subcommandNames() {
			fmt.Println(s)
		}
		return
	}

	cur := tokens[len(tokens)-1]

	if len(tokens) >= 2 && tokens[len(tokens)-2] == "-C" {
		for _, name := range completion.RepoNames(root) {
			if strings.HasPrefix(name, cur) {
				fmt.Println(name)
			}
		}
		return
	}

	clean := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "-C" {
			i++
		} else {
			clean = append(clean, tokens[i])
		}
	}
	if len(clean) == 0 {
		return
	}

	if len(clean) == 1 {
		for _, s := range subcommandNames() {
			if strings.HasPrefix(s, cur) {
				fmt.Println(s)
			}
		}
		return
	}

	sub := clean[0]
	rest := clean[1:]
	c, ok := registry[sub]
	if !ok {
		return
	}
	for _, candidate := range c.Complete(rest) {
		fmt.Println(candidate)
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

	root, err := os.Getwd()
	if err != nil {
		ui.Fatalf("cannot determine working directory: %v", err)
	}

	var repos []string
	if specifiedDir != "" {
		absDir, absErr := filepath.Abs(specifiedDir)
		if absErr != nil || !repo.IsGitRepo(absDir) {
			ui.Fatalf("invalid or non-git directory: %s", specifiedDir)
		}
		repos = []string{absDir}
		root = absDir
	} else {
		repos = repo.FindGitRepos(root)
		if len(repos) == 0 {
			fmt.Fprintln(os.Stderr, "No git repositories found in current directory.")
			os.Exit(1)
		}
	}

	c, ok := registry[sub]
	if !ok {
		ui.Fatalf("unknown subcommand: %s", sub)
	}
	if err := c.Run(root, repos, rest); err != nil {
		ui.Fatalf("%v", err)
	}
}
