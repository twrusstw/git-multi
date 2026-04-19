package fetch

import "gitmulti/internal/command"

func Cmd() *command.Command {
	return &command.Command{Run: run, Complete: complete}
}

func run(root string, repos []string, args []string) error {
	FetchAll(repos)
	return nil
}

func complete(args []string) []string { return nil }
