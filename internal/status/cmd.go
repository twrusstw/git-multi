package status

import "gitmulti/internal/command"

func StatusCmd() *command.Command {
	return &command.Command{Run: statusRun, Complete: statusComplete}
}

func statusRun(root string, repos []string, args []string) error {
	for _, r := range repos {
		ShowStatus(r)
	}
	return nil
}

func statusComplete(args []string) []string { return nil }

func DiscardCmd() *command.Command {
	return &command.Command{Run: discardRun, Complete: discardComplete}
}

func discardRun(root string, repos []string, args []string) error {
	DiscardChangesMulti(repos)
	return nil
}

func discardComplete(args []string) []string { return nil }
