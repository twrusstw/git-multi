#!/usr/bin/env bash

echo "Installing git-multi..."
INSTALL_DIR="$HOME/.git-multi"
BIN_DIR="/usr/local/bin"

# Define the installation directory
sudo mkdir -p "$INSTALL_DIR"/ || exit 1
sudo cp ./core.sh "$BIN_DIR"/gitmulti || exit 1
sudo cp ./auto-completion.sh "$INSTALL_DIR"/auto-completion.sh

# Make the script executable
echo "Making gitmulti script executable..."
sudo chmod 777 "$BIN_DIR"/gitmulti || exit 1
sudo chmod 777 "$INSTALL_DIR"/auto-completion.sh || exit 1

# Print installation complete message
echo "gitmulti installation complete!"


# Prompt the user to install auto-completion
echo "Do you want to install auto-completion for gitmulti? [Y/n]"
read -r choice
if [[ $choice =~ ^[Yy]$ ]]; then
# Append the code to the end of the .bashrc or .zshrc file
  # {
  #   echo ""
  #   echo "# Add gitmulti completion"
  #   echo "source $INSTALL_DIR/auto-completion.sh"
  # } >> ~/.bashrc
  {
    echo ""
    echo "# Add gitmulti completion"
    echo "source $INSTALL_DIR/auto-completion.sh"
  } >> ~/.zshrc
   echo "Auto-completion for gitmulti installed!"
else
  echo "Skip Auto-completion for gitmulti."
fi