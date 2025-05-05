#!/bin/bash

# Wrapper script for the coding-agent

# Check if ANTHROPIC_API_KEY is set
if [ -z "$ANTHROPIC_API_KEY" ]; then
    # Try to load from .env file in current directory
    if [ -f ".env" ]; then
        export $(grep -v '^#' .env | xargs)
    # Try to load from ~/.coding-agent-env
    elif [ -f "$HOME/.coding-agent-env" ]; then
        export $(grep -v '^#' "$HOME/.coding-agent-env" | xargs)
    fi
    
    # If still not set, prompt the user
    if [ -z "$ANTHROPIC_API_KEY" ]; then
        echo "ANTHROPIC_API_KEY not found. Please enter your Anthropic API key:"
        read -r API_KEY
        export ANTHROPIC_API_KEY="$API_KEY"
        
        # Ask if they want to save it
        echo "Would you like to save this API key for future use? (y/n)"
        read -r SAVE_KEY
        if [[ "$SAVE_KEY" == "y" || "$SAVE_KEY" == "Y" ]]; then
            echo "ANTHROPIC_API_KEY=$API_KEY" > "$HOME/.coding-agent-env"
            echo "API key saved to $HOME/.coding-agent-env"
        fi
    fi
fi

# Find and execute the coding-agent binary
if command -v coding-agent &> /dev/null; then
    # If the binary is in PATH
    coding-agent
elif [ -f "./coding-agent" ]; then
    # If the binary is in the current directory
    ./coding-agent
else
    # Try to run with go run
    if [ -f "./main.go" ]; then
        go run main.go
    else
        echo "Error: Could not find the coding-agent binary or main.go file."
        echo "Please make sure you're in the correct directory or that the agent is installed."
        exit 1
    fi
fi
