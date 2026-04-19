# Design: Command struct with function fields (Go idiomatic style)

## Context

The current `internal/cmd/` package defines a `Command` interface with method receivers on private structs. This is non-idiomatic Go — the stdlib (`cmd/go/internal/base`) uses a `Command` struct with function fields instead. Additionally, the package name `cmd` conflicts with Go's convention of using `cmd/` for binary entry points. This refactor aligns with Go stdlib style while keeping error propagation (for testability).

## Goal

- Replace `type Command interface { Run(...) error; Complete(...) []string }` with `type Command struct { Run func(...) error; Complete func(...) []string }`
- Rename package `internal/cmd/` → `internal/command/`
- Each subcommand `cmd.go` drops struct + method receivers in favour of plain functions assigned to struct fields
- All existing tests remain valid (Run still returns error)
- No behaviour change at runtime

---

## Architecture

### internal/command/command.go (renamed from internal/cmd/)

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

### Per-package cmd.go pattern

Before:
```go
type pullCmd struct{}
func Cmd() cmd.Command         { return pullCmd{} }
func (pullCmd) Run(...) error  { ... }
func (pullCmd) Complete(...) []string { ... }
```

After:
```go
func Cmd() *command.Command {
    return &command.Command{Run: run, Complete: complete}
}
func run(root string, repos []string, args []string) error { ... }
func complete(args []string) []string { ... }
```

### main.go

Registry type changes from `map[string]cmd.Command` to `map[string]*command.Command`. Dispatch and error handling unchanged:

```go
if err := registry[sub].Run(root, repos, rest); err != nil {
    ui.Fatalf("%v", err)
}
```

---

## Files Changed

| File | Action |
|------|--------|
| `internal/cmd/command.go` | **Delete** |
| `internal/cmd/command_test.go` | **Move → internal/command/command_test.go** |
| `internal/command/command.go` | **New** — struct definition + ArgOrEmpty |
| `internal/pull/cmd.go` | **Modify** — drop struct/receiver, use function fields |
| `internal/push/cmd.go` | **Modify** |
| `internal/fetch/cmd.go` | **Modify** |
| `internal/branch/cmd.go` | **Modify** |
| `internal/status/cmd.go` | **Modify** |
| `internal/stash/cmd.go` | **Modify** |
| `main.go` | **Modify** — import path + registry type |

---

## Constraints

- All existing tests pass unchanged (Run still returns error)
- No new external dependencies
- `internal/command/` remains a leaf package (imports only `strings`)

---

## Verification

```bash
go build ./...
go vet ./...
go test ./...
gitmulti __complete switch   # tab completion works
gitmulti status              # smoke test
```
