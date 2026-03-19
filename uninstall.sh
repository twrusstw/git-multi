#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="/usr/local/bin"
INSTALL_DIR="$HOME/.git-multi"

echo "Removing gitmulti binary..."
sudo rm -f "$BIN_DIR/gitmulti"

echo "Removing installation directory..."
rm -rf "$INSTALL_DIR"

remove_completion_lines() {
  local rc="$1"
  if [[ ! -f "$rc" ]]; then
    return
  fi
  if grep -q "gitmulti completion" "$rc" 2>/dev/null; then
    sed -i '' '/# Add gitmulti completion/d' "$rc"
    sed -i '' '/source.*auto-completion\.sh/d' "$rc"
    echo "Removed auto-completion entry from $rc."
  fi
}

remove_completion_lines "$HOME/.zshrc"
remove_completion_lines "$HOME/.bashrc"

echo "Uninstall complete. Please restart your shell or run: exec \$SHELL"
