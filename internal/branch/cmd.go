package branch

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"gitmulti/internal/command"
	"gitmulti/internal/status"
	"gitmulti/internal/validate"
)

// ---- switch subcommand ----

func SwitchCmd() *command.Command {
	return &command.Command{Run: switchRun, Complete: switchComplete}
}

func switchRun(root string, repos []string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("switch requires a branch name")
	}
	switch args[0] {
	case "-f":
		if len(args) < 2 {
			return fmt.Errorf("switch -f requires a branch name")
		}
		if err := validate.BranchName(args[1]); err != nil {
			return err
		}
		for _, r := range repos {
			SwitchForce(r, args[1])
		}
	case "-c":
		if len(args) < 2 {
			return fmt.Errorf("switch -c requires a branch name")
		}
		if err := validate.BranchName(args[1]); err != nil {
			return err
		}
		for _, r := range repos {
			CreateIfModified(r, args[1])
		}
	default:
		if err := validate.BranchName(args[0]); err != nil {
			return err
		}
		for _, r := range repos {
			Switch(r, args[0])
		}
	}
	return nil
}

func switchComplete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()

	branches := func() []string {
		var out []string
		for _, name := range ListAllNames(root, "") {
			if strings.HasPrefix(name, cur) {
				out = append(out, name)
			}
		}
		return out
	}

	if len(args) == 1 {
		out := branches()
		for _, f := range []string{"-f", "-c"} {
			if strings.HasPrefix(f, cur) {
				out = append(out, f)
			}
		}
		return out
	}
	if len(args) == 2 && args[0] == "-f" {
		return branches()
	}
	return nil
}

// ---- branch subcommand ----

func BranchCmd() *command.Command {
	return &command.Command{Run: branchRun, Complete: branchComplete}
}

func branchRun(root string, repos []string, args []string) error {
	if len(args) == 0 {
		status.ShowCurrentAll(repos)
		return nil
	}
	switch args[0] {
	case "-a":
		keyword := command.ArgOrEmpty(args[1:])
		if err := validate.Keyword(keyword); err != nil {
			return err
		}
		ListAll(root, keyword)
	case "--find":
		if len(args) < 2 {
			return fmt.Errorf("branch --find requires a keyword")
		}
		for _, r := range repos {
			Find(r, args[1])
		}
	case "-d":
		if len(args) < 2 {
			return fmt.Errorf("branch -d requires a branch name")
		}
		deleteRemote := slices.Contains(args, "--remote")
		Delete(repos, args[1], deleteRemote)
	case "-D":
		if len(args) < 2 {
			return fmt.Errorf("branch -D requires a branch name")
		}
		deleteRemote := slices.Contains(args, "--remote")
		ForceDelete(repos, args[1], deleteRemote)
	case "-m":
		if len(args) < 3 {
			return fmt.Errorf("branch -m requires <old> <new>")
		}
		Rename(repos, args[1], args[2])
	default:
		return fmt.Errorf("unknown branch flag: %s", args[0])
	}
	return nil
}

func branchComplete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()

	if len(args) == 1 {
		var out []string
		for _, flag := range []string{"-a", "--find", "-d", "-D", "-m"} {
			if strings.HasPrefix(flag, cur) {
				out = append(out, flag)
			}
		}
		return out
	}
	if len(args) == 2 {
		switch args[0] {
		case "--find", "-d", "-D", "-m":
			var out []string
			for _, name := range ListAllNames(root, "") {
				if strings.HasPrefix(name, cur) {
					out = append(out, name)
				}
			}
			return out
		}
	}
	return nil
}
