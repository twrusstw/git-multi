package fetch

import "gitmulti/internal/cmd"

type fetchCmd struct{}

func Cmd() cmd.Command { return fetchCmd{} }

func (fetchCmd) Run(root string, repos []string, args []string) error {
	FetchAll(repos)
	return nil
}

func (fetchCmd) Complete(args []string) []string { return nil }
