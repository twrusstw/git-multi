#!/usr/bin/env bash

_gitmulti_complete() {
  local cur prev
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  COMPREPLY=( $(gitmulti __complete "$prev" "$cur" 2>/dev/null) )
}

if [[ -n "$ZSH_VERSION" ]]; then
  autoload -U bashcompinit && bashcompinit
fi

complete -F _gitmulti_complete gitmulti
