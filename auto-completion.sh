#!/bin/bash

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

function count_all_branches {
    keyword="$1"
    count=0
    for d in */ ; do
        if [ -d "$d/.git" ]; then
            cd "$d" || exit
            cmd=""
            if [ -z "$keyword" ]; then
                cmd="git branch -all"
            else
                cmd="git branch -a --list '*$keyword*'"
            fi
            while read -r branch; do
                count=$((count+1))
            done < <(eval "$cmd" | sed 's/\*//g' | tr -d ' ')
            cd ..
        fi
    done
    echo "$count"
}

# Define the _gitmulti_list_all_branches function
function _gitmulti_list_all_branches() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-p -fp -s"
    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
    if [ -z "$cur" ]; then
        return 0
    fi
    COMPREPLY=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")" | awk '{print $NF}') )
    return 0
}

# Register the _gitmulti_list_all_branches function as a completion handler for the gitmulti command
complete -F _gitmulti_list_all_branches gitmulti