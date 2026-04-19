# Design: Refactor main.go via Command Interface + Registry

## Context

`main.go` has grown to 345 lines. It does too many things: argument parsing for every subcommand, completion logic that duplicates the dispatch switch, and no abstraction boundary between "routing" and "execution". Adding a new subcommand requires touching four places in main.go. This refactor introduces a `Command` interface so main.go becomes a thin dispatcher (~60 lines) and each subcommand owns its own argument parsing and completion logic.

## Goal

- `main.go` reduced to ~60 lines (registry + -C parsing + repo discovery + dispatch)
- Each subcommand package encapsulates its own arg parsing and tab-completion
- Adding a new subcommand requires only: new package + one line in the registry
- No new external dependencies

---

## Architecture

### 1. New package: `internal/cmd/`

Defines the interface and shared arg utility:

```go
// internal/cmd/command.go
package cmd

type Command interface {
    Run(root string, repos []string, args []string) error
    Complete(args []string) []string
}

// ArgOrEmpty returns the first non-flag argument, or "".
// Moved from main.go's argOrEmpty().
func ArgOrEmpty(args []string) string
```

`root` is the working directory (or `-C` path). Most commands ignore it; `branch -a` needs it for `branch.ListAll(root, keyword)`.

### 2. Per-package `cmd.go` files

Each `internal/<name>/` package gains a `cmd.go` exporting a constructor:

| File | Constructor | Handles |
|------|-------------|---------|
| `internal/pull/cmd.go` | `pull.Cmd()` | `"pull"` |
| `internal/push/cmd.go` | `push.Cmd()` | `"push"` |
| `internal/fetch/cmd.go` | `fetch.Cmd()` | `"fetch"` |
| `internal/branch/cmd.go` | `branch.BranchCmd()`, `branch.SwitchCmd()` | `"branch"`, `"switch"` |
| `internal/status/cmd.go` | `status.StatusCmd()`, `status.DiscardCmd()` | `"status"`, `"discard"` |
| `internal/stash/cmd.go` | `stash.Cmd()` | `"stash"` |

Each `Run()` is responsible for its own `validate` calls (currently done in main.go before dispatch).

### 3. Revised `main.go`

```go
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
```

Dispatch:
```go
c, ok := registry[sub]
if !ok {
    ui.Fatalf("unknown subcommand: %s", sub)
}
if err := c.Run(root, repos, rest); err != nil {
    ui.Fatalf("%v", err)
}
```

`runComplete()` replaces its per-subcommand switch with:
```go
if c, ok := registry[sub]; ok {
    c.Complete(tokens)
}
```

The `subcommands` slice is removed; completion lists subcommand names from `registry` keys (sorted).

---

## Files Changed

| File | Action |
|------|--------|
| `internal/cmd/command.go` | **New** — Command interface + ArgOrEmpty |
| `internal/pull/cmd.go` | **New** — pull.Cmd() |
| `internal/push/cmd.go` | **New** — push.Cmd() |
| `internal/fetch/cmd.go` | **New** — fetch.Cmd() |
| `internal/branch/cmd.go` | **New** — BranchCmd() + SwitchCmd() |
| `internal/status/cmd.go` | **New** — StatusCmd() + DiscardCmd() |
| `internal/stash/cmd.go` | **New** — stash.Cmd() |
| `main.go` | **Modified** — registry + thin dispatch only |

Existing logic files (`pull/ops.go`, `branch/ops.go`, etc.) are **not modified**. The `cmd.go` files are thin wrappers that call existing exported functions.

---

## Constraints

- `ui.PromptYN` / `ui.PromptMenu` must stay on the main goroutine — this is already enforced by the existing ops; `cmd.go` wrappers don't introduce goroutines.
- `validate` calls move from `main.go` into each `Run()`.
- No mocks in tests — `testutil.InitRepo` / `testutil.CloneRepo` pattern continues.

---

## Verification

```bash
go build ./...          # must compile cleanly
go vet ./...            # no issues
go test ./...           # all existing tests pass
gitmulti pull --help    # smoke test each subcommand
gitmulti branch -a
gitmulti switch -c feat/test
gitmulti __complete switch   # tab completion still works
```
