#!/usr/bin/env bash

_gitmulti_complete() {
  local IFS=$'\n'
  IFS=$'\n' read -r -d '' -a COMPREPLY < <(gitmulti __complete "${COMP_WORDS[@]:1}" 2>/dev/null; printf '\0')
}

if [[ -n "$ZSH_VERSION" ]]; then
  autoload -U bashcompinit && bashcompinit
fi

complete -F _gitmulti_complete gitmulti
