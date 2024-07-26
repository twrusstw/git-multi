#!/usr/bin/env bash

function git_pull {
    branch_name="${1:-$(git branch --show-current)}"
    printf '\e[36m%s\e[0m: Pulling to branch %s.\n' "$cur_dir" "$branch_name"

    git fetch --all
    if ! git checkout "$branch_name"; then
        return
    fi

    if git pull; then
        printf '\e[36m%s\e[0m: Pulled changes from branch %s.\n' "$cur_dir"  "$branch_name"
    else
        printf '\e[36m%s\e[0m: Stashing local changes and pulling changes from branch %s.\n' "$cur_dir"  "$branch_name"
        git stash
        if git pull; then
            printf '\e[36m%s\e[0m: Pulled changes from branch %s.\n' "$cur_dir"  "$branch_name"
        else
            read -p "Do you want to force pull and discard all changes? This action cannot be undone. (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                git reset --hard origin/"$branch_name"
                printf '\e[36m%s\e[0m: Pulled changes from branch %s.\n' "$cur_dir" "$branch_name"
            else
                printf '\e[36m%s\e[0m: Operation cancelled.\n' "$cur_dir"
            fi
        fi
    fi
    printf '\n'
}

function git_pull_force {
    branch_name="$1"
    if [ -z "$branch_name" ]; then
        current_branch=$(git branch --show-current)
        branch_name="$current_branch"
    else
        branch_name="$1"
    fi
    printf '\e[36m%s\e[0m: Force Pulling to branch %s.\n' "$cur_dir" "$branch_name"

    git fetch --all
    switch_branch_force_no_print "$branch_name"
    git reset --hard origin/"$branch_name"

    printf '\e[36m%s\e[0m: Force Pulled changes from branch %s.\n' "$cur_dir"  "$branch_name"
    printf '\n'
}

function switch_branch {
    branch_name="$1"
    if [ -z "$branch_name" ]; then
        echo "Error: Missing arguments."
        show_help
        exit 1
    fi
    current_branch=$(git branch --show-current)
    if [ "$current_branch" = "$branch_name" ]; then
        printf '\e[36m%s\e[0m: Already on branch %s.\n' "$cur_dir" "$branch_name"
        return
    fi
    if ! git diff-index --quiet HEAD --; then
        printf '\e[36m%s\e[0m: Your local changes to the following files would be overwritten by checkout:\n' "$cur_dir"
        git diff-index --name-only HEAD --
        read -p "Do you want to discard all changes and switch to branch $branch_name? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            return
        fi
        git reset --hard HEAD
    fi
    if git show-ref --verify --quiet "refs/heads/$branch_name"; then
        printf '\e[36m%s\e[0m: Switching to branch %s.\n' "$cur_dir" "$branch_name"
        git checkout "$branch_name"
    fi
    if git show-ref --verify --quiet "refs/heads/$branch_name" || git ls-remote --exit-code --heads origin "$branch_name" >/dev/null; then
        printf '\e[36m%s\e[0m: Switching to branch %s.\n' "$cur_dir" "$branch_name"
        git checkout "$branch_name"
    fi
}


function switch_branch_force_no_print {
    branch_name="$1"
    if [ -z "$branch_name" ]; then
        return
    fi
    current_branch=$(git branch --show-current)
    if [ "$current_branch" = "$branch_name" ]; then
        return
    fi
    if git show-ref --verify --quiet "refs/heads/$branch_name" || git ls-remote --exit-code --heads origin "$branch_name" >/dev/null; then
        git checkout -f "$branch_name"
    fi
}

function switch_branch_force {
    branch_name="$1"
    if [ -z "$branch_name" ]; then
        echo "Error: Missing arguments."
        show_help
        exit 1
    fi
    current_branch=$(git branch --show-current)
    if [ "$current_branch" = "$branch_name" ]; then
        printf '\e[36m%s\e[0m: Already on branch %s.\n' "$cur_dir" "$branch_name"
    else
        if git show-ref --verify --quiet "refs/heads/$branch_name" || git ls-remote --exit-code --heads origin "$branch_name" >/dev/null; then
            printf '\e[36m%s\e[0m: Switching to branch %s.\n' "$cur_dir" "$branch_name"
            git checkout -f "$branch_name"
        fi
    fi
}

function find_branch {
    keyword="$1"
    if git show-ref --verify --quiet "refs/heads/$keyword"; then
        printf '\e[32mBranch found in %s\e[0m\n' "$cur_dir"
    else
        printf '\e[31m(%s)Same branch not found' "$cur_dir"
        found_cnt=$(git branch -a -r --list "*$keyword*" | wc -l | tr -d ' ')
        if [ "$found_cnt" != 0 ]; then
            printf ', similar branch\e[0m:\n'
            git branch  -a -r --list "*$keyword*" | sed 's/\*//g' | tr -d ' '
        else
            printf '\n'
        fi
    fi
}

function list_all_branches {
    branches=()
    keyword="$1"
    for d in */ ; do
        cur_dir="${d%/}"
        if [ -d "$d/.git" ]; then
            cd "$cur_dir" || exit
            cmd=""
            if [ -z "$keyword" ]; then
                cmd="git branch --all"
            else
                cmd="git branch -a --list '*$keyword*'"
            fi
            while read -r branch; do
                if [[ ! "${branches[*]}" =~ $branch ]]; then
                    branches+=("$branch")
                fi
            done < <(eval "$cmd" | sed 's/\*//g' | tr -d ' ')
            cd ..
        fi
    done
    printf "%s\n" "${branches[@]}"
}

function discard_changes {
    printf '\e[36m%s\e[0m: Discarding changes.\n' "$cur_dir"
    git checkout .
    git clean -fd
    printf '\e[36m%s\e[0m: Discarded changes.\n' "$cur_dir"
    printf '\n'
}

function show_current_branch {
    current_branch=$(git branch --show-current)
    remote_branch=$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null)
    if [ -n "$remote_branch" ]; then
        behind=$(git rev-list --count "$remote_branch"..HEAD)
        ahead=$(git rev-list --count HEAD.."$remote_branch")
    else
        behind="N/A"
        ahead="N/A"
    fi
    if [ "$1" = true ]; then
        printf "\e[1m%-20s %-30s %-20s %-5s\e[0m\\n"  "Group" "Repository" "Branch" "Status"
    fi
    untracked=$(git status --porcelain | grep -Ec '^ ??')
    unstaged=$(git status --porcelain | grep -c '^ [MD]')
    staged=$(git diff --cached --name-only | wc -l|tr -d ' ')
    group_name=$(git remote get-url origin | sed 's/.*:\/\/[^/]*\/\([^/]*\)\/.*/\1/')
    branch_status="(⇣$ahead ⇡$behind ?$untracked !$unstaged +$staged)"
    if [ "$unstaged" -gt 0 ] || [ "$untracked" -gt 0 ] || [ "$staged" -gt 0 ]; then
        printf "\e[36m%-20s\e[0m %-30s %-20s \e[42;30m%-5s\e[0m\n"  "$group_name" "$cur_dir" "$current_branch" "$branch_status"
    else
        printf "\e[36m%-20s\e[0m \e[33m%-30s\e[0m %-20s %-5s\n" "$group_name" "$cur_dir" "$current_branch" "$branch_status"
    fi
}

function create_branch_if_modified {
    local branch_name="$1"
    local skip_extensions="(.mod|.sum)$" # Define extensions to skip branch creation

    # Check if branch name is provided
    if [ -z "$branch_name" ]; then
        echo "Error: Branch name is required."
        return 1
    fi

    # Check for uncommitted changes using git status
    if git status --porcelain | grep -q '^ [AMD]'; then
        # Check if all changes are only for specified extensions
        if ! git status --porcelain | grep '^ [AMD]' | grep -vqE "$skip_extensions"; then
            echo "All changes are in files with extensions to skip (.mod, .sum). No new branch created."
            return 0
        fi

        echo "Uncommitted changes detected. Creating a new branch: $branch_name"

        # Check if the branch already exists
        if git show-ref --verify --quiet "refs/heads/$branch_name"; then
            echo "Error: Branch '$branch_name' already exists."
            return 1
        fi

        # Create and switch to the new branch

        if git checkout -b "$branch_name"
        then
            echo "Successfully created and switched to branch '$branch_name'."
        else
            echo "Failed to create branch '$branch_name'."
            return 1
        fi
    else
        echo "No uncommitted changes. No need to create a new branch."
    fi
}

function show_help {
    echo "Usage: gitmulti [OPTION] [BRANCH] [-d DIRECTORY]"
    echo "Checkout, pull, switch, or discard changes in all Git repositories in the current directory and its subdirectories."
    echo ""
    echo "Options:"
    echo "  -p    Pull the specified branch in each repository."
    echo "  -pf   Force pull the specified branch in each repository."
    echo "  -s    Switch to the specified branch in each repository."
    echo "  -sf   Force switch to the specified branch in each repository."
    echo "  -F    Find the specified branch in each repository."
    echo "  -b    Show the current branch in each repository."
    echo "  -al   List all branches in each repository."
    echo "  -d    Specify the directory to use. This option must be followed by the directory path."
    echo "  -dc   Discard changes in each repository."
    echo "  -st   Show the status of each repository."
    echo "  -nb   Create a new branch if there are uncommitted changes."
    echo "  -h    Show this help message."
    echo ""
    echo "Examples:"
    echo "  gitmulti -s feature-branch"
    echo "  gitmulti -p master"
    echo "  gitmulti -F hotfix-branch"
}

if [ $# -eq 0 ]; then
    echo "Error: No arguments provided."
    show_help
    exit 1
fi

if [ "$1" = "-h" ]; then
    show_help
    exit 0
fi

case "$1" in
    -p|-pf|-s|-sf|-nb|-b|-dc|-st)
        is_specified_dir=false
        if [ "$2" = "-d" ]; then
            cur_dir="${3%/}"
            is_specified_dir=true
        elif [ "$3" = "-d" ]; then
            cur_dir="${4%/}"
            branch_name="$2"
            is_specified_dir=true
        fi

        if [ "$is_specified_dir" = true ]; then
            if [ -n "$cur_dir" ] && [ -d "$cur_dir" ] && [ -d "$cur_dir/.git" ]; then
                cd "$cur_dir" || exit
                case "$1" in
                    -p) git_pull "${branch_name:-}" ;;
                    -pf) git_pull_force "${branch_name:-}" ;;
                    -s) switch_branch "${branch_name}" ;;
                    -sf) switch_branch_force "${branch_name}" ;;
                    -b) show_current_branch true ;;
                    -dc) discard_changes ;;
                    -st) git status ;;
                esac
                cd ..
                exit 0
            else
                echo "Error: Invalid directory."
                show_help
                exit 1
            fi
        fi
        ;;
    -F)
        if [ $# -lt 2 ]; then
            echo "Error: Missing arguments."
            show_help
            exit 1
        fi
        ;;
    -al) list_all_branches "$2" && exit 0;;
    *)
        echo "Error: Invalid option." && show_help && exit 1
        ;;
esac

if [ "$1" = "-al" ]; then
    list_all_branches
    exit 0
fi

first_loop=true
for d in */ ; do
    cur_dir="${d%/}"
    if [ -d "$d/.git" ]; then
        cd "$cur_dir" || exit
        case "$1" in
            -p) git_pull "$2" ;;
            -pf) git_pull_force "$2" ;;
            -sf) switch_branch_force "$2" ;;
            -s) switch_branch "$2" ;;
            -nb) create_branch_if_modified "$2" ;;
            -b)
                show_current_branch "$first_loop"
                first_loop=false
                ;;
            -F) find_branch "$2" ;;
            -dc) discard_changes ;;
            -st)
                if git status --porcelain | grep -q .; then
                    printf '\e[36m%s\e[0m: Status of branch %s:\n' "$cur_dir" "$(git branch --show-current)"
                    git status
                else
                    printf '\e[36m%s\e[0m: No changes to show.\n' "$cur_dir"
                fi
                ;;
            *)
                echo "Error: Invalid option." && show_help && exit 1
                ;;
        esac
        cd ..
    fi
done