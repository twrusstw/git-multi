package push

import (
	"os"
	"strings"

	"gitmulti/internal/cmd"
	"gitmulti/internal/completion"
	"gitmulti/internal/validate"
)

type pushCmd struct{}

func Cmd() cmd.Command { return pushCmd{} }

func (pushCmd) Run(root string, repos []string, args []string) error {
	branchName := cmd.ArgOrEmpty(args)
	if err := validate.BranchName(branchName); err != nil {
		return err
	}
	for _, r := range repos {
		Push(r, branchName)
	}
	return nil
}

func (pushCmd) Complete(args []string) []string {
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
