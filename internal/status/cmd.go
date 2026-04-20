package status

import "gitmulti/internal/command"

func Cmd() *command.Command {
	return &command.Command{Run: statusRun, Complete: statusComplete}
}

func statusRun(_ string, repos []string, _ []string) error {
	for _, r := range repos {
		ShowStatus(r)
	}
	return nil
}

func statusComplete(_ []string) []string { return nil }

func DiscardCmd() *command.Command {
	return &command.Command{Run: discardRun, Complete: discardComplete}
}

func discardRun(_ string, repos []string, _ []string) error {
	DiscardChangesMulti(repos)
	return nil
}

func discardComplete(_ []string) []string { return nil }
