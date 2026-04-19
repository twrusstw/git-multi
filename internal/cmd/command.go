package cmd

import "strings"

// Command is implemented by every gitmulti subcommand.
type Command interface {
	// Run executes the subcommand. root is the working directory or -C path.
	Run(root string, repos []string, args []string) error
	// Complete returns tab-completion candidates for the given args.
	// args includes the current partial token as the last element.
	Complete(args []string) []string
}

// ArgOrEmpty returns the first non-flag element of args, or "".
func ArgOrEmpty(args []string) string {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}
