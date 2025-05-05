package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/joho/godotenv"
	"github.com/ttli3/terminal-coding-agent/pkg/agent"
	"github.com/ttli3/terminal-coding-agent/pkg/tools"
)

func main() {
	// Load anthropic key from env
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %s\n", err.Error())
		fmt.Println("Looking for ANTHROPIC_API_KEY in environment variables...")
	}
	
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY not found in environment variables or .env file")
		fmt.Println("Please set your ANTHROPIC_API_KEY environment variable or create a .env file with ANTHROPIC_API_KEY=your_key")
		return
	}
	
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	// Get all tools
	toolDefinitions := tools.GetAllTools()
	
	// Create and run the agent
	codingAgent := agent.NewAgent(&client, getUserMessage, toolDefinitions)
	err := codingAgent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
