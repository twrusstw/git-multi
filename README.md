# git-multi

A Go CLI utility that executes Git commands across multiple repositories at once. Run it from a parent directory and it operates on every immediate subdirectory that contains a `.git` folder.

## Installation

```bash
git clone https://github.com/twrusstw/git-multi.git
cd git-multi
chmod +x install.sh
./install.sh
exec zsh  # or exec bash
```

The install script builds the binary, copies it to `/usr/local/bin/gitmulti`, and optionally sets up tab completion in your shell.

## Usage

```
gitmulti <subcommand> [args] [-C directory]
```

| Subcommand | Description |
|------------|-------------|
| `pull [--rebase] [branch]` | Pull (ff-only → stash+pull → group conflict menu) |
| `push [branch]` | Push; auto-sets upstream for new branches |
| `fetch` | Fetch then show ahead/behind/dirty table |
| `switch <branch>` | Switch branch (prompts on uncommitted changes) |
| `switch -f <branch>` | Force switch (no prompt) |
| `switch -c <branch>` | Create new branch in repos with changes |
| `branch` | Show ahead/behind/dirty status table |
| `branch -a [keyword]` | List all unique branches (optional filter) |
| `branch -ag [keyword]` | List branches grouped by repository |
| `branch --find <keyword>` | Find branches matching keyword across repos |
| `branch -d <name> [--remote]` | Delete branch (confirm per repo) |
| `branch -D <name> [--remote]` | Force-delete branch (warns if unmerged) |
| `branch -m <old> <new>` | Rename branch (offers remote sync) |
| `status` | Show file-level changes for repos with changes |
| `stash` | Stash changes in all dirty repos |
| `stash pop` | Pop stash (shows conflicts with resolution hint) |
| `stash apply` | Apply stash without dropping it |
| `stash list` | List stash entries for all repos |
| `discard` | Discard all changes (prompts first) |

**Flags:**

| Flag | Description |
|------|-------------|
| `-C <path>` | Target a single repository directory |

## Examples

```bash
# Switch all repos to develop
gitmulti switch develop

# Pull main in all repos
gitmulti pull main

# Show branch status across all repos
gitmulti branch

# List all branches matching a keyword
gitmulti branch -a hotfix

# List branches grouped by repo
gitmulti branch -ag

# Operate on a single repo
gitmulti pull main -C my-service

# Create a new branch in repos with changes
gitmulti switch -c feat/my-feature
```

## Tab Completion

Tab completion is set up automatically during installation for both bash and zsh.

```bash
gitmulti <TAB>             # list subcommands
gitmulti switch <TAB>      # list available branches
gitmulti pull -C <TAB>     # list git repo directories
```

## Uninstall

```bash
chmod +x uninstall.sh
./uninstall.sh
```

Removes the binary, the install directory, and the completion entry from `~/.zshrc` and `~/.bashrc`.

## License

MIT
