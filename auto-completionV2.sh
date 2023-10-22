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

# Define the _gitmulti_pull function
function _gitmulti_pull() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local branches=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")") )
    COMPREPLY=( "${branches[@]}" )
    return 0
}

# Define the _gitmulti_force_pull function
function _gitmulti_force_pull() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local branches=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")") )
    COMPREPLY=( "${branches[@]}" )
    return 0
}

# Define the _gitmulti_switch function
function _gitmulti_switch() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local branches=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")") )
    COMPREPLY=( "${branches[@]}" )
    return 0
}

# Define the _gitmulti_find function
function _gitmulti_find() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local branches=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")") )
    COMPREPLY=( "${branches[@]}" )
    return 0
}

# Define the _gitmulti_branch function
function _gitmulti_branch() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local branches=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")") )
    COMPREPLY=( "${branches[@]}" )
    return 0
}

# Define the _gitmulti_list_all_branches function
function _gitmulti_list_all_branches() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local branches=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")") )
    COMPREPLY=( "${branches[@]}" )
    return 0
}

# Define the _gitmulti_directory function
function _gitmulti_directory() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local directories=( */ )
    COMPREPLY=( "${directories[@]}" )
    return 0
}

# Define the _gitmulti_discard_changes function
function _gitmulti_discard_changes() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local branches=( $(cd "$(dirname "${cur}")" && list_all_branches "$(basename "${cur}")") )
    COMPREPLY=( "${branches[@]}" )
    return 0
}

# Define the _gitmulti_status function
function _gitmulti_status() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ -z "$cur" ]; then
        return 0
    fi
    local directories=( */ )
    COMPREPLY=( "${directories[@]}" )
    return 0
}

# Define the _gitmulti_help function
function _gitmulti_help() {
    COMPREPLY=()
    return 0
}

# Register the completion functions as completion handlers for the gitmulti command
complete -F _gitmulti_pull gitmulti -p
complete -F _gitmulti_force_pull gitmulti -pf
complete -F _gitmulti_switch gitmulti -s
complete -F _gitmulti_find gitmulti -f
complete -F _gitmulti_branch gitmulti -b
complete -F _gitmulti_list_all_branches gitmulti -al
complete -F _gitmulti_directory gitmulti -d
complete -F _gitmulti_discard_changes gitmulti -dc
complete -F _gitmulti_status gitmulti -st
complete -F _gitmulti_help gitmulti -h