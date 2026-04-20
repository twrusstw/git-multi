package stash

import (
	"fmt"
	"strings"

	"gitmulti/internal/command"
)

func Cmd() *command.Command {
	return &command.Command{Run: run, Complete: complete}
}

func run(_ string, repos []string, args []string) error {
	if len(args) == 0 {
		Stash(repos)
		return nil
	}
	switch args[0] {
	case "pop":
		Pop(repos)
	case "apply":
		Apply(repos)
	case "list":
		List(repos)
	default:
		return fmt.Errorf("unknown stash subcommand: %s", args[0])
	}
	return nil
}

func complete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	if len(args) == 1 {
		var out []string
		for _, s := range []string{"pop", "apply", "list"} {
			if strings.HasPrefix(s, cur) {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
