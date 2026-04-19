# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Does

**git-multi** is a zero-dependency Go CLI (`gitmulti`) that runs Git operations across all repositories in a directory. It scans for immediate subdirectories containing `.git` and dispatches commands to each — sequentially for interactive ops, concurrently for safe ones.

## Build & Install

```bash
make build       # build locally
make install     # build + install to /usr/local/bin/gitmulti
make completion  # append auto-completion to ~/.zshrc
make uninstall   # remove binary, ~/.git-multi, and zshrc entry
```

## Development

```bash
go test ./...                   # run all tests
go test ./internal/branch/...   # run tests for a specific package
go vet ./...                    # static analysis
```

No external dependencies — `go.sum` exists but is empty.

## Usage

Full subcommand reference: see `.claude_doc/usage.md`

For architecture details: see `.claude_doc/architecture.md`

## Code conventions

- All git calls go through `gitutil` — never call `exec.Command("git", ...)` directly.
- Interactive prompts use `ui.PromptYN` / `ui.PromptMenu` — they share `ui.StdinReader`, which is not goroutine-safe; prompts MUST run on the main goroutine.
- Multi-repo fan-out uses `util.ParallelMap` / `util.ParallelDo` — don't spawn raw goroutines.
- New subcommands belong in their own `internal/<name>/` package; wire them into `main.go`'s `switch sub` and into `subcommands` + `runComplete` for tab-completion.
- Tests use `testutil.InitRepo` / `testutil.CloneRepo` for real git repos — no mocks.

## After changing code

After any non-trivial change (new subcommand, package added/removed/renamed,
concurrency or output-handling change, serial↔parallel shift), re-check
`.claude_doc/architecture.md` and update it if the package list, key behaviors,
or serial/parallel classification no longer matches reality. Keep `.claude_doc/usage.md`
in sync with `ui.ShowHelp` in `internal/ui/terminal.go` and with `README.md`.
