package pull

import (
	"os"
	"strings"

	"gitmulti/internal/command"
	"gitmulti/internal/completion"
	"gitmulti/internal/validate"
)

func Cmd() *command.Command {
	return &command.Command{Run: run, Complete: complete}
}

func run(root string, repos []string, args []string) error {
	rebase := len(args) > 0 && args[0] == "--rebase"
	branchName := command.ArgOrEmpty(args)
	if err := validate.BranchName(branchName); err != nil {
		return err
	}
	if rebase {
		PullRebase(repos, branchName)
	} else {
		PullAll(repos, branchName)
	}
	return nil
}

func complete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()

	branches := func() []string {
		var out []string
		for _, name := range completion.BranchNames(root) {
			if strings.HasPrefix(name, cur) {
				out = append(out, name)
			}
		}
		return out
	}

	if len(args) == 1 {
		out := branches()
		if strings.HasPrefix("--rebase", cur) {
			out = append(out, "--rebase")
		}
		return out
	}
	if len(args) == 2 && args[0] == "--rebase" {
		return branches()
	}
	return nil
}
