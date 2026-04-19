# main.go Refactor: Command Interface + Registry

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce main.go from 345 lines to ~60 by introducing a `cmd.Command` interface and moving argument parsing + completion logic into each subcommand's own package.

**Architecture:** New `internal/cmd/` defines the `Command` interface (`Run` + `Complete`). Each `internal/<name>/` package gains a `cmd.go` file that wraps existing ops. `main.go` becomes a registry map + thin dispatcher. No existing ops files are modified.

**Tech Stack:** Go stdlib only. No new external dependencies.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/cmd/command.go` | Create | `Command` interface + `ArgOrEmpty` helper |
| `internal/pull/cmd.go` | Create | `pull.Cmd()` — arg parsing, validation, Complete |
| `internal/push/cmd.go` | Create | `push.Cmd()` — arg parsing, validation, Complete |
| `internal/fetch/cmd.go` | Create | `fetch.Cmd()` — delegates to FetchAll, no-op Complete |
| `internal/branch/cmd.go` | Create | `branch.SwitchCmd()` + `branch.BranchCmd()` |
| `internal/status/cmd.go` | Create | `status.StatusCmd()` + `status.DiscardCmd()` |
| `internal/stash/cmd.go` | Create | `stash.Cmd()` — subcommand dispatch + Complete |
| `main.go` | Rewrite | Registry map + -C parsing + repo discovery + dispatch |

---

## Task 1: internal/cmd — interface + ArgOrEmpty

**Files:**
- Create: `internal/cmd/command.go`
- Create: `internal/cmd/command_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/cmd/command_test.go
package cmd_test

import (
	"testing"

	"gitmulti/internal/cmd"
)

func TestArgOrEmpty(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"main"}, "main"},
		{[]string{"--rebase", "main"}, "main"},
		{[]string{"--rebase"}, ""},
		{[]string{"-f", "feat/x"}, "feat/x"},
	}
	for _, tc := range cases {
		if got := cmd.ArgOrEmpty(tc.args); got != tc.want {
			t.Errorf("ArgOrEmpty(%v) = %q, want %q", tc.args, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/cmd/...
```

Expected: `no Go files in .../internal/cmd` or compile error.

- [ ] **Step 3: Write the implementation**

```go
// internal/cmd/command.go
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/cmd/...
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/cmd/command.go internal/cmd/command_test.go
git commit -m "feat: add cmd.Command interface and ArgOrEmpty helper"
```

---

## Task 2: pull command

**Files:**
- Create: `internal/pull/cmd.go`
- Create: `internal/pull/cmd_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/pull/cmd_test.go
package pull_test

import (
	"testing"

	"gitmulti/internal/pull"
)

func TestPullComplete(t *testing.T) {
	c := pull.Cmd()

	// first arg: should include --rebase
	got := c.Complete([]string{"--reb"})
	found := false
	for _, s := range got {
		if s == "--rebase" {
			found = true
		}
	}
	if !found {
		t.Errorf("Complete([\"--reb\"]) missing \"--rebase\", got %v", got)
	}

	// after --rebase: no extra flags
	got2 := c.Complete([]string{"--rebase", "--"})
	for _, s := range got2 {
		if s == "--rebase" {
			t.Errorf("Complete after --rebase should not suggest --rebase again, got %v", got2)
		}
	}
}

func TestPullRunInvalidBranch(t *testing.T) {
	c := pull.Cmd()
	err := c.Run("", []string{}, []string{"bad branch!!"})
	if err == nil {
		t.Error("expected error for invalid branch name, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/pull/... -run TestPullComplete -run TestPullRunInvalidBranch
```

Expected: compile error (`pull.Cmd undefined`).

- [ ] **Step 3: Write the implementation**

```go
// internal/pull/cmd.go
package pull

import (
	"os"
	"strings"

	"gitmulti/internal/cmd"
	"gitmulti/internal/completion"
	"gitmulti/internal/validate"
)

type pullCmd struct{}

func Cmd() cmd.Command { return pullCmd{} }

func (pullCmd) Run(root string, repos []string, args []string) error {
	rebase := len(args) > 0 && args[0] == "--rebase"
	branchName := cmd.ArgOrEmpty(args)
	if err := validate.BranchName(branchName); err != nil {
		return err
	}
	if rebase {
		PullRebase(repos, branchName)
	} else {
		PullAll(repos, branchName)
	}
	return nil
}

func (pullCmd) Complete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()

	branches := func() []string {
		var out []string
		for _, name := range completion.BranchNames(root) {
			if strings.HasPrefix(name, cur) {
				out = append(out, name)
			}
		}
		return out
	}

	if len(args) == 1 {
		out := branches()
		if strings.HasPrefix("--rebase", cur) {
			out = append(out, "--rebase")
		}
		return out
	}
	if len(args) == 2 && args[0] == "--rebase" {
		return branches()
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/pull/... -run TestPullComplete -run TestPullRunInvalidBranch
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/pull/cmd.go internal/pull/cmd_test.go
git commit -m "feat: pull.Cmd() implements cmd.Command"
```

---

## Task 3: push command

**Files:**
- Create: `internal/push/cmd.go`
- Create: `internal/push/cmd_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/push/cmd_test.go
package push_test

import (
	"testing"

	"gitmulti/internal/push"
)

func TestPushRunInvalidBranch(t *testing.T) {
	c := push.Cmd()
	err := c.Run("", []string{}, []string{"bad!!"})
	if err == nil {
		t.Error("expected error for invalid branch name, got nil")
	}
}

func TestPushCompleteReturnsSlice(t *testing.T) {
	c := push.Cmd()
	// Complete must return a non-nil slice (even if empty) when called
	got := c.Complete([]string{""})
	if got == nil {
		t.Error("Complete should return non-nil slice")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/push/... -run TestPushRunInvalidBranch -run TestPushCompleteReturnsSlice
```

Expected: compile error (`push.Cmd undefined`).

- [ ] **Step 3: Write the implementation**

```go
// internal/push/cmd.go
package push

import (
	"os"
	"strings"

	"gitmulti/internal/cmd"
	"gitmulti/internal/completion"
	"gitmulti/internal/validate"
)

type pushCmd struct{}

func Cmd() cmd.Command { return pushCmd{} }

func (pushCmd) Run(root string, repos []string, args []string) error {
	branchName := cmd.ArgOrEmpty(args)
	if err := validate.BranchName(branchName); err != nil {
		return err
	}
	for _, r := range repos {
		Push(r, branchName)
	}
	return nil
}

func (pushCmd) Complete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()
	var out []string
	for _, name := range completion.BranchNames(root) {
		if strings.HasPrefix(name, cur) {
			out = append(out, name)
		}
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/push/... -run TestPushRunInvalidBranch -run TestPushCompleteReturnsSlice
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/push/cmd.go internal/push/cmd_test.go
git commit -m "feat: push.Cmd() implements cmd.Command"
```

---

## Task 4: fetch command

**Files:**
- Create: `internal/fetch/cmd.go`

No test file: `fetch.Cmd()` is a zero-arg delegator. FetchAll is already tested by existing tests.

- [ ] **Step 1: Write the implementation**

```go
// internal/fetch/cmd.go
package fetch

import "gitmulti/internal/cmd"

type fetchCmd struct{}

func Cmd() cmd.Command { return fetchCmd{} }

func (fetchCmd) Run(root string, repos []string, args []string) error {
	FetchAll(repos)
	return nil
}

func (fetchCmd) Complete(args []string) []string { return nil }
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/fetch/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/fetch/cmd.go
git commit -m "feat: fetch.Cmd() implements cmd.Command"
```

---

## Task 5: branch + switch commands

**Files:**
- Create: `internal/branch/cmd.go`
- Create: `internal/branch/cmd_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/branch/cmd_test.go
package branch_test

import (
	"testing"

	"gitmulti/internal/branch"
)

func TestSwitchRunValidation(t *testing.T) {
	c := branch.SwitchCmd()

	// no args
	if err := c.Run("", []string{}, []string{}); err == nil {
		t.Error("expected error for missing branch name")
	}
	// invalid branch
	if err := c.Run("", []string{}, []string{"bad!branch"}); err == nil {
		t.Error("expected error for invalid branch name")
	}
	// -f with no branch
	if err := c.Run("", []string{}, []string{"-f"}); err == nil {
		t.Error("expected error: -f requires branch name")
	}
	// -c with no branch
	if err := c.Run("", []string{}, []string{"-c"}); err == nil {
		t.Error("expected error: -c requires branch name")
	}
}

func TestSwitchComplete(t *testing.T) {
	c := branch.SwitchCmd()
	got := c.Complete([]string{"-"})
	flags := map[string]bool{}
	for _, s := range got {
		flags[s] = true
	}
	if !flags["-f"] || !flags["-c"] {
		t.Errorf("SwitchCmd.Complete([\"-\"]) should include -f and -c, got %v", got)
	}
}

func TestBranchRunValidation(t *testing.T) {
	c := branch.BranchCmd()
	// unknown flag
	if err := c.Run("", []string{}, []string{"--unknown"}); err == nil {
		t.Error("expected error for unknown branch flag")
	}
	// -d with no name
	if err := c.Run("", []string{}, []string{"-d"}); err == nil {
		t.Error("expected error: -d requires branch name")
	}
}

func TestBranchComplete(t *testing.T) {
	c := branch.BranchCmd()
	got := c.Complete([]string{"-"})
	want := map[string]bool{"-a": true, "--find": true, "-d": true, "-D": true, "-m": true}
	for _, s := range got {
		delete(want, s)
	}
	if len(want) > 0 {
		t.Errorf("BranchCmd.Complete([\"-\"]) missing flags: %v", want)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/branch/... -run TestSwitchRunValidation -run TestSwitchComplete -run TestBranchRunValidation -run TestBranchComplete
```

Expected: compile error (`branch.SwitchCmd undefined`).

- [ ] **Step 3: Write the implementation**

```go
// internal/branch/cmd.go
package branch

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"gitmulti/internal/cmd"
	"gitmulti/internal/completion"
	"gitmulti/internal/status"
	"gitmulti/internal/validate"
)

// ---- switch subcommand ----

type switchCmd struct{}

func SwitchCmd() cmd.Command { return switchCmd{} }

func (switchCmd) Run(root string, repos []string, args []string) error {
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

func (switchCmd) Complete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()

	branches := func() []string {
		var out []string
		for _, name := range completion.BranchNames(root) {
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

type branchCmd struct{}

func BranchCmd() cmd.Command { return branchCmd{} }

func (branchCmd) Run(root string, repos []string, args []string) error {
	if len(args) == 0 {
		status.ShowCurrentAll(repos)
		return nil
	}
	switch args[0] {
	case "-a":
		keyword := cmd.ArgOrEmpty(args[1:])
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

func (branchCmd) Complete(args []string) []string {
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
			for _, name := range completion.BranchNames(root) {
				if strings.HasPrefix(name, cur) {
					out = append(out, name)
				}
			}
			return out
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/branch/... -run TestSwitchRunValidation -run TestSwitchComplete -run TestBranchRunValidation -run TestBranchComplete
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/branch/cmd.go internal/branch/cmd_test.go
git commit -m "feat: branch.SwitchCmd() and branch.BranchCmd() implement cmd.Command"
```

---

## Task 6: status + discard commands

**Files:**
- Create: `internal/status/cmd.go`

No test file: both are single-line delegators with no argument parsing.

- [ ] **Step 1: Write the implementation**

```go
// internal/status/cmd.go
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/status/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/status/cmd.go
git commit -m "feat: status.StatusCmd() and status.DiscardCmd() implement cmd.Command"
```

---

## Task 7: stash command

**Files:**
- Create: `internal/stash/cmd.go`
- Create: `internal/stash/cmd_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/stash/cmd_test.go
package stash_test

import (
	"testing"

	"gitmulti/internal/stash"
)

func TestStashRunUnknownSubcommand(t *testing.T) {
	c := stash.Cmd()
	err := c.Run("", []string{}, []string{"badsubcmd"})
	if err == nil {
		t.Error("expected error for unknown stash subcommand")
	}
}

func TestStashComplete(t *testing.T) {
	c := stash.Cmd()
	got := c.Complete([]string{"p"})
	found := false
	for _, s := range got {
		if s == "pop" {
			found = true
		}
	}
	if !found {
		t.Errorf("Complete([\"p\"]) should include \"pop\", got %v", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/stash/... -run TestStashRunUnknownSubcommand -run TestStashComplete
```

Expected: compile error (`stash.Cmd undefined`).

- [ ] **Step 3: Write the implementation**

```go
// internal/stash/cmd.go
package stash

import (
	"fmt"
	"strings"

	"gitmulti/internal/cmd"
)

type stashCmd struct{}

func Cmd() cmd.Command { return stashCmd{} }

func (stashCmd) Run(root string, repos []string, args []string) error {
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

func (stashCmd) Complete(args []string) []string {
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/stash/... -run TestStashRunUnknownSubcommand -run TestStashComplete
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/stash/cmd.go internal/stash/cmd_test.go
git commit -m "feat: stash.Cmd() implements cmd.Command"
```

---

## Task 8: Rewrite main.go

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Replace main.go with the new registry-based dispatcher**

Replace the entire contents of `main.go` with:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gitmulti/internal/branch"
	"gitmulti/internal/cmd"
	"gitmulti/internal/completion"
	"gitmulti/internal/fetch"
	"gitmulti/internal/pull"
	"gitmulti/internal/push"
	"gitmulti/internal/repo"
	"gitmulti/internal/stash"
	"gitmulti/internal/status"
	"gitmulti/internal/ui"
)

var registry = map[string]cmd.Command{
	"pull":    pull.Cmd(),
	"push":    push.Cmd(),
	"fetch":   fetch.Cmd(),
	"switch":  branch.SwitchCmd(),
	"branch":  branch.BranchCmd(),
	"status":  status.StatusCmd(),
	"stash":   stash.Cmd(),
	"discard": status.DiscardCmd(),
}

func subcommandNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func runComplete(tokens []string) {
	root, _ := os.Getwd()

	if len(tokens) == 0 {
		for _, s := range subcommandNames() {
			fmt.Println(s)
		}
		return
	}

	cur := tokens[len(tokens)-1]

	if len(tokens) >= 2 && tokens[len(tokens)-2] == "-C" {
		for _, name := range completion.RepoNames(root) {
			if strings.HasPrefix(name, cur) {
				fmt.Println(name)
			}
		}
		return
	}

	clean := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "-C" {
			i++
		} else {
			clean = append(clean, tokens[i])
		}
	}
	if len(clean) == 0 {
		return
	}

	if len(clean) == 1 {
		for _, s := range subcommandNames() {
			if strings.HasPrefix(s, cur) {
				fmt.Println(s)
			}
		}
		return
	}

	sub := clean[0]
	rest := clean[1:]
	c, ok := registry[sub]
	if !ok {
		return
	}
	for _, candidate := range c.Complete(rest) {
		fmt.Println(candidate)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		ui.ShowHelp()
		return
	}

	if args[0] == "__complete" {
		runComplete(args[1:])
		return
	}

	var specifiedDir string
	var filtered []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-C" {
			if i+1 >= len(args) {
				ui.Fatalf("-C requires a directory path.")
			}
			specifiedDir = filepath.Clean(args[i+1])
			i++
		} else {
			filtered = append(filtered, args[i])
		}
	}
	args = filtered

	if len(args) == 0 {
		ui.Fatalf("no subcommand provided.")
	}

	sub := args[0]
	rest := args[1:]

	if sub == "help" || sub == "-h" || sub == "--help" {
		ui.ShowHelp()
		return
	}

	root, err := os.Getwd()
	if err != nil {
		ui.Fatalf("cannot determine working directory: %v", err)
	}

	var repos []string
	if specifiedDir != "" {
		absDir, absErr := filepath.Abs(specifiedDir)
		if absErr != nil || !repo.IsGitRepo(absDir) {
			ui.Fatalf("invalid or non-git directory: %s", specifiedDir)
		}
		repos = []string{absDir}
		root = absDir
	} else {
		repos = repo.FindGitRepos(root)
		if len(repos) == 0 {
			fmt.Fprintln(os.Stderr, "No git repositories found in current directory.")
			os.Exit(1)
		}
	}

	c, ok := registry[sub]
	if !ok {
		ui.Fatalf("unknown subcommand: %s", sub)
	}
	if err := c.Run(root, repos, rest); err != nil {
		ui.Fatalf("%v", err)
	}
}
```

- [ ] **Step 2: Build to verify it compiles**

```bash
go build ./...
```

Expected: no output, binary `gitmulti` created.

- [ ] **Step 3: Run all tests**

```bash
go test ./...
```

Expected: all `PASS`, no failures.

- [ ] **Step 4: Run vet**

```bash
go vet ./...
```

Expected: no output.

- [ ] **Step 5: Smoke test**

```bash
cd /tmp && mkdir smoke-test && cd smoke-test
git init repoA && git init repoB
cd ..
# From parent:
./gitmulti status
./gitmulti __complete switch
./gitmulti __complete branch
./gitmulti __complete stash p
```

Expected: `status` runs without panic; `__complete` outputs subcommand candidates.

- [ ] **Step 6: Commit**

```bash
git add main.go
git commit -m "refactor: replace main.go switch/case with cmd.Command registry"
```

---

## Task 9: Final verification + architecture doc update

- [ ] **Step 1: Run full test suite**

```bash
go test ./... -v 2>&1 | tail -20
```

Expected: all `PASS`.

- [ ] **Step 2: Check main.go line count**

```bash
wc -l main.go
```

Expected: ≤ 80 lines.

- [ ] **Step 3: Update .claude_doc/architecture.md**

Per the project's `CLAUDE.md` rule — after a structural change, update `.claude_doc/architecture.md` to reflect:
- New `internal/cmd/` package (Command interface + ArgOrEmpty)
- Each `internal/<name>/cmd.go` file added
- main.go now a thin registry dispatcher

- [ ] **Step 4: Commit**

```bash
git add .claude_doc/architecture.md
git commit -m "docs: update architecture.md for cmd.Command refactor"
```
