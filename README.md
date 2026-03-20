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
gitmulti [OPTION] [BRANCH] [-d DIRECTORY]
```

| Option | Argument | Description |
|--------|----------|-------------|
| `-p` | `[branch]` | Pull branch (stash fallback, then prompts for hard reset) |
| `-pf` | `[branch]` | Force pull — hard reset to origin |
| `-s` | `<branch>` | Switch branch (prompts if uncommitted changes exist) |
| `-sf` | `<branch>` | Force switch branch |
| `-F` | `<keyword>` | Find branch by keyword across all repos |
| `-b` | | Show current branch, ahead/behind counts, and dirty state |
| `-al` | `[keyword]` | List all unique branches (optionally filtered by keyword) |
| `-st` | | Show git status for repos with uncommitted changes |
| `-dc` | | Discard all changes (`checkout . && clean -fd`) |
| `-nb` | `<branch>` | Create new branch for repos with uncommitted changes |
| `-d` | `<path>` | Target a single specific repository |
| `-h` | | Show help |

## Examples

```bash
# Switch all repos to develop
gitmulti -s develop

# Pull main in all repos
gitmulti -p main

# Show branch status across all repos
gitmulti -b

# List all branches matching a keyword
gitmulti -al hotfix

# Operate on a single repo
gitmulti -p main -d my-service

# Create a new branch in repos with changes (skips .mod/.sum only changes)
gitmulti -nb feat/my-feature
```

## Tab Completion

Tab completion is set up automatically during installation for both bash and zsh.

```bash
gitmulti -<TAB>        # list all flags
gitmulti -s <TAB>      # list available branches
gitmulti -d <TAB>      # list git repo directories
```

## Uninstall

```bash
chmod +x uninstall.sh
./uninstall.sh
```

Removes the binary, the install directory, and the completion entry from `~/.zshrc` and `~/.bashrc`.

## License

MIT
