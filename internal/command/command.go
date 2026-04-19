package command

import "strings"

type Command struct {
	Run      func(root string, repos []string, args []string) error
	Complete func(args []string) []string
}

func ArgOrEmpty(args []string) string {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}
