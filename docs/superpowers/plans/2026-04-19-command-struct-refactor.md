# Command Struct Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the `cmd.Command` interface (method receivers on private structs) with a `command.Command` struct (function fields), and rename the package from `internal/cmd/` to `internal/command/`.

**Architecture:** Create `internal/command/command.go` with a struct type, then update each subcommand's `cmd.go` to drop the private struct and method receivers in favour of plain functions assigned to struct fields. Delete `internal/cmd/` after all callsites are migrated.

**Tech Stack:** Go 1.24, no external dependencies

---

### Task 1: Create `internal/command` package (TDD)

**Files:**
- Create: `internal/command/command.go`
- Create: `internal/command/command_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/command/command_test.go`:

```go
package command_test

import (
	"testing"

	"gitmulti/internal/command"
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
		if got := command.ArgOrEmpty(tc.args); got != tc.want {
			t.Errorf("ArgOrEmpty(%v) = %q, want %q", tc.args, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test gitmulti/internal/command`
Expected: FAIL — package does not exist yet

- [ ] **Step 3: Implement `internal/command/command.go`**

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test gitmulti/internal/command`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/command/command.go internal/command/command_test.go
git commit -m "feat: add internal/command package with Command struct"
```

---

### Task 2: Migrate `internal/pull/cmd.go`

**Files:**
- Modify: `internal/pull/cmd.go`
- Test: `internal/pull/cmd_test.go` (no changes needed — call syntax identical)

- [ ] **Step 1: Run existing tests to confirm baseline**

Run: `go test gitmulti/internal/pull`
Expected: PASS

- [ ] **Step 2: Replace `internal/pull/cmd.go`**

```go
package pull

import (
	"os"
	"strings"

	"gitmulti/internal/command"
	"gitmulti/internal/completion"
	"gitmulti/internal/validate"
)

func Cmd() *command.Command {
	return &command.Command{Run: run, Complete: complete}
}

func run(root string, repos []string, args []string) error {
	rebase := len(args) > 0 && args[0] == "--rebase"
	branchName := command.ArgOrEmpty(args)
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

func complete(args []string) []string {
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

- [ ] **Step 3: Run tests to verify they pass**

Run: `go test gitmulti/internal/pull`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/pull/cmd.go
git commit -m "refactor: pull.Cmd() uses command.Command struct"
```

---

### Task 3: Migrate `internal/push/cmd.go`

**Files:**
- Modify: `internal/push/cmd.go`

- [ ] **Step 1: Run existing tests to confirm baseline**

Run: `go test gitmulti/internal/push`
Expected: PASS

- [ ] **Step 2: Replace `internal/push/cmd.go`**

```go
package push

import (
	"os"
	"strings"

	"gitmulti/internal/command"
	"gitmulti/internal/completion"
	"gitmulti/internal/validate"
)

func Cmd() *command.Command {
	return &command.Command{Run: run, Complete: complete}
}

func run(root string, repos []string, args []string) error {
	branchName := command.ArgOrEmpty(args)
	if err := validate.BranchName(branchName); err != nil {
		return err
	}
	for _, r := range repos {
		Push(r, branchName)
	}
	return nil
}

func complete(args []string) []string {
	cur := ""
	if len(args) > 0 {
		cur = args[len(args)-1]
	}
	root, _ := os.Getwd()
	out := []string{}
	for _, name := range completion.BranchNames(root) {
		if strings.HasPrefix(name, cur) {
			out = append(out, name)
		}
	}
	return out
}
```

- [ ] **Step 3: Run tests**

Run: `go test gitmulti/internal/push`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/push/cmd.go
git commit -m "refactor: push.Cmd() uses command.Command struct"
```

---

### Task 4: Migrate `internal/fetch/cmd.go`

**Files:**
- Modify: `internal/fetch/cmd.go`

- [ ] **Step 1: Run existing tests to confirm baseline**

Run: `go test gitmulti/internal/fetch`
Expected: PASS

- [ ] **Step 2: Replace `internal/fetch/cmd.go`**

```go
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
```

- [ ] **Step 3: Run tests**

Run: `go test gitmulti/internal/fetch`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/fetch/cmd.go
git commit -m "refactor: fetch.Cmd() uses command.Command struct"
```

---

### Task 5: Migrate `internal/branch/cmd.go`

**Files:**
- Modify: `internal/branch/cmd.go`

- [ ] **Step 1: Run existing tests to confirm baseline**

Run: `go test gitmulti/internal/branch`
Expected: PASS

- [ ] **Step 2: Replace `internal/branch/cmd.go`**

```go
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
```

- [ ] **Step 3: Run tests**

Run: `go test gitmulti/internal/branch`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/branch/cmd.go
git commit -m "refactor: branch.SwitchCmd() and BranchCmd() use command.Command struct"
```

---

### Task 6: Migrate `internal/status/cmd.go`

**Files:**
- Modify: `internal/status/cmd.go`

- [ ] **Step 1: Run existing tests to confirm baseline**

Run: `go test gitmulti/internal/status`
Expected: PASS

- [ ] **Step 2: Replace `internal/status/cmd.go`**

```go
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
```

- [ ] **Step 3: Run tests**

Run: `go test gitmulti/internal/status`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/status/cmd.go
git commit -m "refactor: status.StatusCmd() and DiscardCmd() use command.Command struct"
```

---

### Task 7: Migrate `internal/stash/cmd.go`

**Files:**
- Modify: `internal/stash/cmd.go`

- [ ] **Step 1: Run existing tests to confirm baseline**

Run: `go test gitmulti/internal/stash`
Expected: PASS

- [ ] **Step 2: Replace `internal/stash/cmd.go`**

```go
package stash

import (
	"fmt"
	"strings"

	"gitmulti/internal/command"
)

func Cmd() *command.Command {
	return &command.Command{Run: run, Complete: complete}
}

func run(root string, repos []string, args []string) error {
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

func complete(args []string) []string {
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

- [ ] **Step 3: Run tests**

Run: `go test gitmulti/internal/stash`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/stash/cmd.go
git commit -m "refactor: stash.Cmd() uses command.Command struct"
```

---

### Task 8: Update `main.go` — switch import to `internal/command`

**Files:**
- Modify: `main.go`

The only change is replacing the import path `gitmulti/internal/cmd` → `gitmulti/internal/command` and updating the registry type from `map[string]cmd.Command` to `map[string]*command.Command`.

- [ ] **Step 1: Update the import in `main.go`**

Replace:
```go
"gitmulti/internal/cmd"
```
with:
```go
"gitmulti/internal/command"
```

- [ ] **Step 2: Update the registry type declaration**

Replace:
```go
var registry = map[string]cmd.Command{
```
with:
```go
var registry = map[string]*command.Command{
```

- [ ] **Step 3: Build to verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Run all tests**

Run: `go test ./...`
Expected: all packages PASS

- [ ] **Step 5: Commit**

```bash
git add main.go
git commit -m "refactor: main.go uses internal/command registry"
```

---

### Task 9: Delete `internal/cmd/` package and verify

**Files:**
- Delete: `internal/cmd/command.go`
- Delete: `internal/cmd/command_test.go`

- [ ] **Step 1: Delete the old package**

```bash
rm internal/cmd/command.go internal/cmd/command_test.go
rmdir internal/cmd
```

- [ ] **Step 2: Build to confirm no dangling imports**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Run full test suite**

Run: `go test ./...`
Expected: all packages PASS, no `[no test files]` regression in previously-tested packages

- [ ] **Step 4: Run vet**

Run: `go vet ./...`
Expected: no issues

- [ ] **Step 5: Smoke test tab completion**

```bash
gitmulti __complete switch
```
Expected: prints `-c`, `-f`, and any local branch names (no panic, no "unknown subcommand")

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor: delete internal/cmd, migration to internal/command complete"
```
