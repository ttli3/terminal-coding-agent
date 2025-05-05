package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/ttli3/terminal-coding-agent/pkg/tools"
)

// Agent represents the coding agent
type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []tools.ToolDefinition
}

// NewAgent creates a new agent
func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), tools []tools.ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

// Run starts the agent
func (a *Agent) Run(ctx context.Context) error {
	fmt.Println("doChat with Claude (use 'ctrl-c' to quit)")

	// Initialize the conversation
	conversation := []anthropic.MessageParam{
		{
			Role: "user",
			Content: []anthropic.ContentBlockParam{
				{
					Type: "text",
					Text: "You are a coding assistant. You can help me with programming tasks. I'll give you tasks, and you can use tools to help me complete them.",
				},
			},
		},
	}

	// Main conversation loop
	for {
		// Get user message
		fmt.Print("You: ")
		userMsg, ok := a.getUserMessage()
		if !ok {
			break
		}

		// Add user message to conversation
		conversation = append(conversation, anthropic.MessageParam{
			Role: "user",
			Content: []anthropic.ContentBlockParam{
				{
					Type: "text",
					Text: userMsg,
				},
			},
		})

		// Get response from Claude
		msg, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}

		// Add Claude's response to conversation
		conversation = append(conversation, anthropic.MessageParam{
			Role: "assistant",
			Content: msg.Content,
		})

		// Print Claude's response
		fmt.Println("Claude:", a.formatResponse(msg))
	}

	return nil
}

// executeTool executes a tool and returns the result
func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParam {
	// Find the tool
	var tool *tools.ToolDefinition
	for _, t := range a.tools {
		if t.Name == name {
			tool = &t
			break
		}
	}

	if tool == nil {
		return anthropic.ContentBlockParam{
			Type: "tool_result",
			ToolResult: &anthropic.ToolResultBlockParam{
				ToolUseID: id,
				Content:   fmt.Sprintf("Error: Tool %s not found", name),
			},
		}
	}

	// Execute the tool
	result, err := tool.Function(input)
	if err != nil {
		return anthropic.ContentBlockParam{
			Type: "tool_result",
			ToolResult: &anthropic.ToolResultBlockParam{
				ToolUseID: id,
				Content:   fmt.Sprintf("Error: %s", err.Error()),
			},
		}
	}

	return anthropic.ContentBlockParam{
		Type: "tool_result",
		ToolResult: &anthropic.ToolResultBlockParam{
			ToolUseID: id,
			Content:   result,
		},
	}
}

// runInference runs the inference with Claude
func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	// Convert tools to the format expected by Claude
	var anthropicTools []anthropic.ToolUnionParam
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	// Create a channel to receive the API response
	resultCh := make(chan struct {
		message *anthropic.Message
		err     error
	})

	// Start the API call in a goroutine
	go func() {
		message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3Opus20240229,
			MaxTokens: int64(4096),
			Messages:  conversation,
			Tools:     anthropicTools,
		})
		resultCh <- struct {
			message *anthropic.Message
			err     error
		}{message, err}
	}()

	// Display loading message with elapsed time
	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Clear the loading message when we're done
	defer func() {
		fmt.Print("\r\033[K") // Clear the current line
	}()

	// Wait for either the API response or a tick to update the loading message
	for {
		select {
		case result := <-resultCh:
			// Process tool calls
			for i, block := range result.message.Content {
				if block.Type == "tool_use" && block.ToolUse != nil {
					// Execute the tool
					toolResult := a.executeTool(block.ToolUse.ID, block.ToolUse.Name, block.ToolUse.Input)
					
					// Print tool execution
					fmt.Printf("tool: %s(%s)\n", block.ToolUse.Name, string(block.ToolUse.Input))
					
					// Replace the tool call with the result
					result.message.Content[i] = toolResult
				}
			}
			return result.message, result.err
		case <-ticker.C:
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\rThinking... %.1fs elapsed", elapsed)
		}
	}
}

// formatResponse formats Claude's response for display
func (a *Agent) formatResponse(msg *anthropic.Message) string {
	var result string
	for _, block := range msg.Content {
		if block.Type == "text" {
			result += block.Text
		} else if block.Type == "tool_result" && block.ToolResult != nil {
			result += fmt.Sprintf("result: %s\n", block.ToolResult.Content)
		}
	}
	return result
}
