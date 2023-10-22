#!/bin/bash

# Define the installation directory
INSTALL_DIR="/usr/local/bin/gitmulti"

# Define the download URL for the gitmulti script
GITMULTI_URL="https://gitlab.com/russliu/git-multi.git"

mdkir -p $INSTALL_DIR
echo "Downloading gitmulti.sh script..."
curl -sSL $GITMULTI_URL -o $INSTALL_DIR/gitmulti.sh
echo "Downloading gitmulti-completion.sh script..."
curl -sSL $GITMULTI_URL -o $INSTALL_DIR/gitmulti-completion.sh

# Make the script executable
echo "Making gitmulti script executable..."
chmod 777 $INSTALL_DIR/gitmulti.sh
chmod 777 $INSTALL_DIR/gitmulti-completion.sh

# Print installation complete message
echo "gitmulti installation complete!"


# Append the code to the end of the .bashrc or .zshrc file
{
  echo ""
  "alias gitmulti="$INSTALL_DIR/gitmulti.sh""
  "# Add gitmulti completion"
  "source /path/to/gitmulti-completion.sh"
 } >> ~/.bashrc
{
  echo ""
  "alias gitmulti="$INSTALL_DIR/gitmulti.sh""
  "# Add gitmulti completion"
  "source /path/to/gitmulti-completion.sh"
 } >> ~/.zshrc