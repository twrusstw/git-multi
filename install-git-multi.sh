#!/bin/bash

INSTALL_DIR="/usr/local/bin/git-multi"

# Define the installation directory
mkdir -p $INSTALL_DIR || true
sudo cp ./core.sh $INSTALL_DIR
sudo cp ./auto-completion.sh $INSTALL_DIR

# Make the script executable
echo "Making gitmulti script executable..."
sudo chmod 777 $INSTALL_DIR/core.sh || exit 1
sudo chmod 777 $INSTALL_DIR/auto-completion.sh || exit 1

# Print installation complete message
echo "gitmulti installation complete!"


echo "alias gitmulti="$INSTALL_DIR/core.sh"" >> ~/.zshrc

# Prompt the user to install auto-completion
echo "Do you want to install auto-completion for gitmulti? [Y/n]"
read -r choice
if [[ $choice =~ ^[Yy]$ ]]; then
# Append the code to the end of the .bashrc or .zshrc file
# {
#   echo ""
#   "# Add gitmulti completion"
#   "source $INSTALL_DIR/auto-completion.sh"
#  } >> ~/.bashrc
{
  echo ""
  "# Add gitmulti completion"
  "source $INSTALL_DIR/auto-completion.sh"
 } >> ~/.zshrc
   echo "Auto-completion for gitmulti installed!"
else
  echo "Skip Auto-completion for gitmulti."
fi