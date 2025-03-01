#!/bin/bash

# Script to add the SSH public key from GitHub secret to the server
# This script extracts the public key and adds it to the server's authorized_keys file

set -e  # Exit on error

# Check if GitHub CLI is available
if ! command -v gh &> /dev/null; then
  echo "Error: GitHub CLI (gh) is not installed."
  echo "Please install it from https://cli.github.com/"
  exit 1
fi

# Check if server details are provided
if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Usage: $0 <server_username> <server_ip>"
  echo "Example: $0 gerald-bahati 209.97.128.72"
  exit 1
fi

SERVER_USER="$1"
SERVER_IP="$2"

# Read the SSH private key from GitHub secret
echo "Reading DO_SSH_PRIVATE_KEY from GitHub secrets..."
DO_SSH_PRIVATE_KEY=$(gh secret get DO_SSH_PRIVATE_KEY -o json | jq -r '.value')

if [ -z "$DO_SSH_PRIVATE_KEY" ]; then
  echo "Error: Could not read DO_SSH_PRIVATE_KEY from GitHub secrets."
  echo "Make sure you're authenticated to GitHub and have access to the repository secrets."
  exit 1
fi

# Extract public key from private key
TMPKEY=$(mktemp)
echo "$DO_SSH_PRIVATE_KEY" > $TMPKEY
chmod 600 $TMPKEY
SSH_PUBLIC_KEY=$(ssh-keygen -y -f $TMPKEY)
rm $TMPKEY

if [ -z "$SSH_PUBLIC_KEY" ]; then
  echo "Error: Could not extract public key from private key"
  exit 1
fi

echo "Extracted public key:"
echo "$SSH_PUBLIC_KEY"
echo ""

# Add the key to the server
echo "Adding public key to server..."
ssh -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP" "mkdir -p ~/.ssh && echo \"$SSH_PUBLIC_KEY\" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"

echo ""
echo "✅ SSH key successfully added to $SERVER_USER@$SERVER_IP!"
echo "You should now be able to use key-based authentication."
