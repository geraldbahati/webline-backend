#!/bin/bash

# Script to add the SSH public key to the server
# This script extracts the public key from your private key and adds it to the server's authorized_keys file

set -e  # Exit on error

# Check if SSH private key and server details are provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <private_key_file> <server_username> <server_ip>"
  echo "Example: $0 ~/.ssh/id_rsa gerald-bahati 209.97.128.72"
  exit 1
fi

PRIVATE_KEY_FILE="$1"
SERVER_USER="$2"
SERVER_IP="$3"

# Check if private key file exists
if [ ! -f "$PRIVATE_KEY_FILE" ]; then
  echo "Error: Private key file '$PRIVATE_KEY_FILE' not found"
  exit 1
fi

# Extract public key from private key
SSH_PUBLIC_KEY=$(ssh-keygen -y -f "$PRIVATE_KEY_FILE")
if [ -z "$SSH_PUBLIC_KEY" ]; then
  echo "Error: Could not extract public key from '$PRIVATE_KEY_FILE'"
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
