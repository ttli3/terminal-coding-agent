#!/bin/bash

# Terminal Coding Agent Installation Script

set -e

echo "Installing Terminal Coding Agent..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    echo "Visit https://golang.org/doc/install for installation instructions."
    exit 1
fi

# Build the binary
echo "Building the binary..."
go build -o coding-agent

# Create destination directory if it doesn't exist
INSTALL_DIR="/usr/local/bin"
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating directory $INSTALL_DIR..."
    sudo mkdir -p "$INSTALL_DIR"
fi

# Move binary to /usr/local/bin
echo "Installing binary to $INSTALL_DIR..."
sudo mv coding-agent "$INSTALL_DIR/"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "No .env file found."
    echo "Please enter your Anthropic API key:"
    read -r API_KEY
    echo "ANTHROPIC_API_KEY=$API_KEY" > ~/.coding-agent-env
    echo "Created ~/.coding-agent-env file with your API key."
    echo "You can also set the ANTHROPIC_API_KEY environment variable in your shell profile."
fi

echo "Installation complete!"
echo "You can now run the coding agent by typing 'coding-agent' in your terminal."
