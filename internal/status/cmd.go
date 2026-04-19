package status

import "gitmulti/internal/cmd"

type statusCmd struct{}

func StatusCmd() cmd.Command { return statusCmd{} }

func (statusCmd) Run(root string, repos []string, args []string) error {
	for _, r := range repos {
		ShowStatus(r)
	}
	return nil
}

func (statusCmd) Complete(args []string) []string { return nil }

type discardCmd struct{}

func DiscardCmd() cmd.Command { return discardCmd{} }

func (discardCmd) Run(root string, repos []string, args []string) error {
	DiscardChangesMulti(repos)
	return nil
}

func (discardCmd) Complete(args []string) []string { return nil }
