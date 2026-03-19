#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="/usr/local/bin"
INSTALL_DIR="$HOME/.git-multi"

echo "Building gitmulti (Go)..."
cd "$SCRIPT_DIR"
go build -o gitmulti . || { echo "Build failed. Make sure Go is installed."; exit 1; }

echo "Installing gitmulti to $BIN_DIR..."
mkdir -p "$INSTALL_DIR"
sudo install -m 755 "$SCRIPT_DIR/gitmulti" "$BIN_DIR/gitmulti"

echo "Copying auto-completion script..."
cp "$SCRIPT_DIR/auto-completion.sh" "$INSTALL_DIR/auto-completion.sh"
chmod 644 "$INSTALL_DIR/auto-completion.sh"

echo "gitmulti (Go) installation complete!"

echo "Do you want to install auto-completion for gitmulti? [Y/n]"
read -r choice
if [[ "${choice:-Y}" =~ ^[Yy]$ ]]; then
  ZSHRC="$HOME/.zshrc"
  if ! grep -q "gitmulti completion" "$ZSHRC" 2>/dev/null; then
    {
      echo ""
      echo "# Add gitmulti completion"
      echo "source $INSTALL_DIR/auto-completion.sh"
    } >> "$ZSHRC"
    echo "Auto-completion installed."
  else
    echo "Auto-completion already present in $ZSHRC."
  fi
else
  echo "Skipping auto-completion."
fi
