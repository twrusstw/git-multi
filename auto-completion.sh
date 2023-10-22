#!/usr/bin/env bash

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
    COMPREPLY=( $(cd "$(dirname "${cur}")" && gitmulti -al "$(basename "${cur}")" | awk '{print $NF}') )
    return 0
}

# Register the _gitmulti_list_all_branches function as a completion handler for the gitmulti command
complete -F _gitmulti_list_all_branches gitmulti