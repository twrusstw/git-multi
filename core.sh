#!/bin/bash

function git_pull {
    branch_name="$1"
    if [ -z "$branch_name" ]; then
        current_branch=$(git branch --show-current)
        branch_name="$current_branch"
    else
        branch_name="$1"
    fi
    printf '\e[36m%s\e[0m: Pulling to branch %s.\n' "$cur_dir" "$branch_name"

    git checkout "$branch_name"
    if git pull; then
        printf '\e[36m%s\e[0m: Pulled changes from branch %s.\n' "$cur_dir"  "$branch_name"
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
    git checkout "$branch_name"
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
    if [ "$current_branch" = "$1" ]; then
        printf '\e[36m%s\e[0m: Already on branch %s.\n' "$cur_dir" "$1"
    elif git show-ref --verify --quiet "refs/heads/$1"; then
        printf '\e[36m%s\e[0m: Switching to branch %s.\n' "$cur_dir" "$1"
        git checkout "$1"
    # else
    #     printf '%s: Branch %s does not exist.\n' "$cur_dir" "$1"
    fi
}

function find_branch {
    keyword="$1"
    if git show-ref --verify --quiet "refs/heads/$keyword"; then
        printf '\e[32mBranch found in %s\e[0m\n' "$cur_dir"
    else
        printf '\e[31mSame name branch not found in %s, ' "$cur_dir"
        found_cnt=$(git branch --list "*$keyword*" | wc -l | tr -d ' ')
        if [ "$found_cnt" != 0 ]; then
            printf 'similar branch\e[0m:\n'
            git branch --list "*$keyword*" | sed 's/\*//g' | tr -d ' '
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
                cmd="git branch -all"
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
    untracked=$(git status --porcelain | grep -c '^??' | tr -d ' ')
    unstaged=$(git status --porcelain | grep -c '^M' | tr -d ' ')
    branch_status="(⇣$ahead ⇡$behind !$unstaged ?$untracked)"
    group_name=$(git remote get-url origin | sed 's/.*:\/\/[^/]*\/\([^/]*\)\/.*/\1/')
    printf "\e[36m%-20s\e[0m %-30s %-20s %-5s\n"  "$group_name" "$cur_dir" "$current_branch" "$branch_status"
}

function show_help {
    echo "Usage: gitmulti [OPTION] [BRANCH] [-d DIRECTORY]"
    echo "Checkout, pull, switch, or discard changes in all Git repositories in the current directory and its subdirectories."
    echo ""
    echo "Options:"
    echo "  -p    Pull the specified branch in each repository."
    echo "  -pf   Force pull the specified branch in each repository."
    echo "  -s    Switch to the specified branch in each repository."
    echo "  -f    Find the specified branch in each repository."
    echo "  -ls   Show the current branch in each repository."
    echo "  -al   List all branches in each repository."
    echo "  -d    Specify the directory to use. This option must be followed by the directory path."
    echo "  -dc   Discard changes in each repository."
    echo "  -st   Show the status of each repository."
    echo "  -h    Show this help message."
    echo ""
    echo "Examples:"
    echo "  gitmulti -s feature-branch"
    echo "  gitmulti -p master"
    echo "  gitmulti -f hotfix-branch"
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
    -p|-pf|-s|-ls|-dc|-st)
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
                    -ls) show_current_branch true ;;
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
    -f)
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
            -s) switch_branch "$2" ;;
            -ls)
                show_current_branch "$first_loop"
                first_loop=false
                ;;
            -f) find_branch "$2" ;;
            -dc) discard_changes ;;
            -st)
                if git status --porcelain | grep -q .; then
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
    else
        echo "$cur_dir is not a Git repository, skipping..."
    fi
done