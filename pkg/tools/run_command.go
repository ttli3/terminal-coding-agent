package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

var RunCommandDefinition = ToolDefinition{
	Name: "run_command",
	Description: `Execute a terminal command.
	
The command will be executed in the current working directory. The output of the command will be returned.
Be careful with commands that may modify the file system or have other side effects.
`,
	InputSchema: RunCommandInputSchema,
	Function:    RunCommand,
}

type RunCommandInput struct {
	Command string `json:"command" jsonschema_description:"The terminal command to execute"`
}

var RunCommandInputSchema = GenerateSchema[RunCommandInput]()

func RunCommand(input json.RawMessage) (string, error) {
	runCommandInput := RunCommandInput{}
	err := json.Unmarshal(input, &runCommandInput)
	if err != nil {
		return "", err
	}

	if runCommandInput.Command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	// Execute the command
	cmd := exec.Command("sh", "-c", runCommandInput.Command)
	output, err := cmd.CombinedOutput()
	
	// Format the output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Command: %s\n\n", runCommandInput.Command))
	result.WriteString("Output:\n")
	result.WriteString(string(output))
	
	if err != nil {
		result.WriteString(fmt.Sprintf("\nError: %s\n", err.Error()))
		return result.String(), nil // Return the error in the output, not as an error
	}
	
	return result.String(), nil
}
