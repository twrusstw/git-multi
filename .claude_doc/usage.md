# Usage Reference

```
gitmulti <subcommand> [args] [-C directory]

Subcommands:
  pull [--rebase] [branch]     Pull (ff-only → stash+pull → group conflict menu)
  push [branch]                Push; auto-sets upstream for new branches
  fetch                        Fetch then show ahead/behind/dirty table
  switch <branch>              Switch branch (prompts on uncommitted changes)
  switch -f <branch>           Force switch (no prompt)
  switch -c <branch>           Create new branch only in repos with non-trivial changes
  branch                       Show ahead/behind/dirty status table
  branch -a [keyword]          List all unique branches (optional filter)
  branch --find <keyword>      Find branches matching keyword across repos
  branch -d <name> [--remote]  Delete branch (confirm per repo)
  branch -D <name> [--remote]  Force-delete branch (warns if unmerged)
  branch -m <old> <new>        Rename branch (offers remote sync)
  status                       Show file-level changes for repos with changes
  stash                        Stash changes in all dirty repos
  stash pop | apply | list     Pop / apply / list stash entries
  discard                      Discard all changes (prompts first)

Flags:
  -C <path>   Target a single repository directory (can appear anywhere)
```

`push`, `switch`, `status`, and `branch -d/-D/-m` run sequentially
(interactive / per-repo output). All other multi-repo ops run concurrently.
