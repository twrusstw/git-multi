#!/usr/bin/env bash

_gitmulti_complete() {
  local IFS=$'\n'
  COMPREPLY=( $(gitmulti __complete "${COMP_WORDS[@]:1}" 2>/dev/null) )
}

if [[ -n "$ZSH_VERSION" ]]; then
  autoload -U bashcompinit && bashcompinit
fi

complete -F _gitmulti_complete gitmulti
