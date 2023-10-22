# Git Multi

Git Multi is a Bash script for executing Git commands in multiple Git subdirectories.

## Installation
1. Download `install-git-multi.sh`
## Usage
1. Run `install-git-multi.sh` in the parent directory that contains multiple Git subdirectories.
2. Use the following commands:

```
Usage: gitmulti [OPTION] [BRANCH] [-d DIRECTORY]
Checkout, pull, switch, or discard changes in all Git repositories in the current directory and its subdirectories.

Options:
  -p    Pull the specified branch in each repository.
  -pf   Force pull the specified branch in each repository.
  -s    Switch to the specified branch in each repository.
  -f    Find the specified branch in each repository.
  -ls   Show the current branch in each repository.
  -al   List all branches in each repository.
  -d    Specify the directory to use. This option must be followed by the directory path.
  -dc   Discard changes in each repository.
  -st   Show the status of each repository.
  -h    Show this help message.

Examples:
  gitmulti -s feature-branch
  gitmulti -p master
  gitmulti -f hotfix-branch
```