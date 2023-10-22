#!/bin/bash

# Define the installation directory
INSTALL_DIR="/usr/local/bin/git-multi"

# Define the download URL for the gitmulti script
GITMULTI_URL="https://gitlab.com/russliu/git-multi.git"

mdkir -p $INSTALL_DIR
echo "Downloading core.sh script..."
curl -sSL $GITMULTI_URL -o $INSTALL_DIR/core.sh
echo "Downloading auto-completion.sh script..."
curl -sSL $GITMULTI_URL -o $INSTALL_DIR/auto-completion.sh

# Make the script executable
echo "Making gitmulti script executable..."
chmod 777 $INSTALL_DIR/core.sh
chmod 777 $INSTALL_DIR/auto-completion.sh

# Print installation complete message
echo "gitmulti installation complete!"


# Append the code to the end of the .bashrc or .zshrc file
{
  echo ""
  "alias gitmulti="$INSTALL_DIR/core.sh""
  "# Add gitmulti completion"
  "source /path/to/auto-completion.sh"
 } >> ~/.bashrc
{
  echo ""
  "alias gitmulti="$INSTALL_DIR/core.sh""
  "# Add gitmulti completion"
  "source /path/to/auto-completion.sh"
 } >> ~/.zshrc