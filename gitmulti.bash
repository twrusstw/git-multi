_gitmulti_complete() {
  local line
  COMPREPLY=()
  while IFS= read -r line; do
    [[ -n "$line" ]] && COMPREPLY+=("$line")
  done < <(gitmulti __complete "${COMP_WORDS[@]:1}" 2>/dev/null)
}

complete -F _gitmulti_complete gitmulti
