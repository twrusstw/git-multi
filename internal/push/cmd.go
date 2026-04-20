package push

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

func run(_ string, repos []string, args []string) error {
	branchName := command.ArgOrEmpty(args)
	if err := validate.BranchName(branchName); err != nil {
		return err
	}
	for _, r := range repos {
		Push(r, branchName)
	}
	return nil
}

func complete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()
	out := []string{}
	for _, name := range completion.BranchNames(root) {
		if strings.HasPrefix(name, cur) {
			out = append(out, name)
		}
	}
	return out
}
