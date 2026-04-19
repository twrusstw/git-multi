# Architecture

## Package structure

```
main.go               — registry map[string]*command.Command + thin dispatcher (~163 lines);
                        parses -C <dir>, discovers repos, dispatches Run()/Complete()
internal/
  command/            — Command struct with function fields (Run + Complete) + ArgOrEmpty helper
  branch/             — Switch, SwitchForce, CreateIfModified, Find, ListAll,
                        Delete, ForceDelete, Rename; skip-file rules (skip.go);
                        cmd.go: BranchCmd(), SwitchCmd()
  pull/               — PullAll (ff-only → stash+pull → group conflict menu),
                        PullRebase (buffered combined output); cmd.go: Cmd()
  push/               — Push (auto-sets upstream for new branches); cmd.go: Cmd()
  fetch/              — FetchAll (fetch + ahead/behind/dirty table); cmd.go: Cmd()
  stash/              — Stash, Pop, Apply, List; cmd.go: Cmd()
  status/             — ShowCurrentAll (table), ShowStatus, DiscardChangesMulti;
                        cmd.go: StatusCmd(), DiscardCmd()
  repo/               — FindGitRepos, IsGitRepo, BranchExists{Local,Remote}, CurrentBranch
  completion/         — RepoNames, BranchNames (cached across repos for tab-complete)
  gitutil/            — Git / GitRun / GitOK / GitCombined wrappers around os/exec
  ui/                 — PromptYN, PromptMenu, LockedPrint, color helpers,
                        PadRight (ANSI-aware), ShowHelp, Fatalf
  validate/           — BranchName / Keyword regex validation (blocks path traversal)
  util/               — ParallelMap, ParallelDo (bounded fan-out), NonEmpty, NormaliseBranchName
  testutil/           — InitRepo, CloneRepo, GitMustRun for integration tests
```

## CLI shape

`gitmulti <subcommand> [args] [-C <dir>]`. Subcommands: `pull`, `push`, `fetch`,
`switch`, `branch`, `status`, `stash`, `discard`. `-C` can appear anywhere and
narrows the working set to one repo; otherwise `repo.FindGitRepos(cwd)` is used.

`gitmulti __complete <tokens...>` powers `auto-completion.sh` — prints candidate
subcommands, flags, branch names, or repo dirs depending on position.

## Key behaviors

- **`pull`**: fetch → ff-only merge → stash+pull → group-prompt for hard reset. Never silently destructive.
- **`pull --rebase`**: per-repo `git pull --rebase` with buffered combined output (see tradeoff below).
- **`push`**: if the local branch has no upstream, push with `-u origin <branch>`; otherwise plain push.
- **`switch <branch>`**: only switches if branch exists locally or on remote; prompts before discarding changes.
- **`switch -f`**: `git checkout -f` (no prompt). **`switch -c`**: creates branch only in repos with non-trivial changes (same skip-file rules as below).
- **`branch`** (no args): table with `⇣pulls ⇡pushes ?untracked !unstaged +staged` indicators; parses remote URL for owner name.
- **`branch -a [keyword]`**: deduplicates branches across all repos.
- **`branch --find <keyword>`**: per-repo substring match.
- **`branch -d / -D [--remote]`**: confirms per repo; `-D` warns when branch is unmerged. `--remote` also deletes `origin/<branch>`.
- **`branch -m <old> <new>`**: rename; offers remote sync.
- **`switch -c`** (new-branch-if-modified) skip list — default: `go.mod`, `go.sum`, `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `bun.lockb`, `Cargo.lock`, `poetry.lock`, `Pipfile.lock`, `composer.lock`. Per-repo overrides via `.gitmulti-skip` (newline-separated, `#` comments).
- **`discard`**: full pre-flight summary, single confirmation, then `checkout . && clean -fd`.
- **`stash pop`**: on conflict, shows resolution hint; does not abort the batch.

## Input validation

`validate/input.go` enforces `^[a-zA-Z0-9._/\-]+$` and rejects `..` on all branch
name and keyword inputs to prevent command injection.

## Tab completion

`auto-completion.sh` registers a bash/zsh handler that calls `gitmulti __complete`
to suggest subcommands, flags, branch names, and repo dirs (for `-C`).

## Concurrency model & output tradeoffs

All multi-repo ops are bounded by `util.ParallelMap` / `util.ParallelDo`
(default `runtime.NumCPU()*2` workers) to avoid forking 50+ git processes at once.

Output handling falls into three patterns:

1. **Collect-then-print** — `ParallelMap` gathers results by index, caller prints serially afterwards. Used by `status.ShowCurrentAll`, `fetch.FetchAll`, `stash.List`, `branch.ListAllNames`, `status.DiscardChangesMulti`. Preserves input order.

2. **Concurrent LockedPrint** — parallel work uses `gitutil.Git` (stdout captured to buffer) and emits lines via `ui.LockedPrint`. Order follows completion time, not input order, but no mid-line interleaving. Used by `pull.PullAll`'s `cascade`, `stash.popOrApply`.

3. **Buffered combined output** — for ops where we need git's own stdout/stderr but can't let it stream to the terminal (parallel `git pull --rebase` / `git stash` would interleave). Uses `gitutil.GitCombined`, flushes under `printMu` when the op finishes. Used by `pull.PullRebase`, `stash.Stash`.

### Tradeoff of pattern 3

Pattern 3 **loses real-time progress**: a slow `pull --rebase` on a large repo
prints nothing until it finishes. This is intentional — the alternative
(`gitutil.GitRun` piping directly to `os.Stdout`) would let parallel rebases
shred each other's output line-by-line, which is worse.

If a user complains about lack of progress on `pull --rebase`:
- **Don't** revert to `GitRun` — that reintroduces interleaving.
- **Do** serialise the loop (drop concurrency for this op), or stream
  line-by-line through a mutex-guarded scanner.

`pull.PullRebase` doc-comment repeats this note so future editors see it inline.

## Serial vs. parallel ops

- **Serial** (for-loop in each subcommand's `cmd.go`): `push`, `switch` (all variants), `status`,
  `branch -d/-D/-m`. These either prompt the user or have per-repo interactive
  output that's easier to read one-at-a-time.
- **Parallel** (via `util.Parallel*`): `pull`, `pull --rebase`, `fetch`,
  `branch` (table), `branch -a`, `branch --find`, `stash`, `stash pop/apply/list`,
  `discard`.

Prompts (`ui.PromptYN`, `ui.PromptMenu`) MUST run on the main goroutine —
`ui.StdinReader` is a shared `bufio.Reader` and is not goroutine-safe. See
comment on `StdinReader`.
