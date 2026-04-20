package fetch

import "gitmulti/internal/command"

func Cmd() *command.Command {
	return &command.Command{Run: run, Complete: complete}
}

func run(_ string, repos []string, _ []string) error {
	All(repos)
	return nil
}

func complete(_ []string) []string { return nil }
