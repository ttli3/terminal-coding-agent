package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/joho/godotenv"
	"github.com/invopop/jsonschema"
)

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []ToolDefinition
}

type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

func main() {
	//load anthropic key from env
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

	tools := []ToolDefinition{ReadFileDefinition, ListFilesDefinition, EditFileDefinition, RunCommandDefinition, GenerateDiffDefinition}
	agent := NewAgent(&client, getUserMessage, tools)
	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

var EditFileDefinition = ToolDefinition{
	Name: "edit_file",
	Description: `Make edits to a text file.

Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.

If the file specified with path doesn't exist, it will be created.
`,
	InputSchema: EditFileInputSchema,
	Function:    EditFile,
}

type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

var EditFileInputSchema = GenerateSchema[EditFileInput]()

func EditFile(input json.RawMessage) (string, error) {
	editFileInput := EditFileInput{}
	err := json.Unmarshal(input, &editFileInput)
	if err != nil {
		return "", err
	}

	if editFileInput.Path == "" || editFileInput.OldStr == editFileInput.NewStr {
		return "", fmt.Errorf("invalid input parameters")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) && editFileInput.OldStr == "" {
			// This is a new file creation case
			dir := filepath.Dir(editFileInput.Path)
			if dir != "." {
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					return "", fmt.Errorf("failed to create directory: %w", err)
				}
			}

			// Generate a diff for the new file (empty -> content)
			var diffResult strings.Builder
			diffResult.WriteString("Creating new file with content:\n")
			
			// Split the content into lines and format as additions
			lines := strings.Split(editFileInput.NewStr, "\n")
			for _, line := range lines {
				if line == "" {
					diffResult.WriteString("\u001b[32m+\u001b[0m\n")
					continue
				}
				diffResult.WriteString(fmt.Sprintf("\u001b[32m+ %s\u001b[0m\n", line))
			}

			err := os.WriteFile(editFileInput.Path, []byte(editFileInput.NewStr), 0644)
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}

			return fmt.Sprintf("Successfully created file %s\n\n%s", editFileInput.Path, diffResult.String()), nil
		}
		return "", err
	}

	oldContent := string(content)
	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, -1)

	if oldContent == newContent && editFileInput.OldStr != "" {
		return "", fmt.Errorf("old_str not found in file")
	}

	// Generate a diff to show the changes using line-by-line comparison
	// Split the code into lines for line-by-line comparison
	originalLines := strings.Split(oldContent, "\n")
	modifiedLines := strings.Split(newContent, "\n")
	
	// Format the diff for better readability
	var diffResult strings.Builder
	diffResult.WriteString("Changes to be applied:\n")
	
	// Use a simple line-by-line diff algorithm
	lcs := longestCommonSubsequence(originalLines, modifiedLines)
	
	i, j := 0, 0
	for k := 0; k < len(lcs); k++ {
		// Print deletions (lines in original but not in LCS)
		for i < lcs[k].originalIndex {
			diffResult.WriteString(fmt.Sprintf("\u001b[31m- %s\u001b[0m\n", originalLines[i]))
			i++
		}
		
		// Print additions (lines in modified but not in LCS)
		for j < lcs[k].modifiedIndex {
			diffResult.WriteString(fmt.Sprintf("\u001b[32m+ %s\u001b[0m\n", modifiedLines[j]))
			j++
		}
		
		// Print unchanged lines (lines in both)
		diffResult.WriteString(fmt.Sprintf("\u001b[90m  %s\u001b[0m\n", originalLines[i]))
		i++
		j++
	}
	
	// Print any remaining deletions
	for i < len(originalLines) {
		diffResult.WriteString(fmt.Sprintf("\u001b[31m- %s\u001b[0m\n", originalLines[i]))
		i++
	}
	
	// Print any remaining additions
	for j < len(modifiedLines) {
		diffResult.WriteString(fmt.Sprintf("\u001b[32m+ %s\u001b[0m\n", modifiedLines[j]))
		j++
	}

	// Write the changes to the file
	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("File updated successfully.\n\n%s", diffResult.String()), nil
}

var ListFilesDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	InputSchema: ListFilesInputSchema,
	Function:    ListFiles,
}

type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

func ListFiles(input json.RawMessage) (string, error) {
	listFilesInput := ListFilesInput{}
	err := json.Unmarshal(input, &listFilesInput)
	if err != nil {
		panic(err)
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		panic(err)
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

var RunCommandDefinition = ToolDefinition{
	Name: "run_command",
	Description: `Execute a terminal command.
	
This tool allows running shell commands like git commands, ls, etc. The command will be executed in the current working directory.
Be careful with commands that might modify the file system or have other side effects.`,
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

	// Split the command string into command and arguments
	parts := strings.Fields(runCommandInput.Command)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid command format")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command failed: %s\nOutput: %s", err.Error(), string(output)), nil
	}

	return string(output), nil
}

var GenerateDiffDefinition = ToolDefinition{
	Name: "generate_diff",
	Description: `Generate a diff between two versions of code.
	
This tool shows the differences between original code and modified code, highlighting additions and removals.
It's useful for visualizing changes before applying them to a file.`,
	InputSchema: GenerateDiffInputSchema,
	Function:    GenerateDiff,
}

type GenerateDiffInput struct {
	OriginalCode string `json:"original_code" jsonschema_description:"The original version of the code"`
	ModifiedCode string `json:"modified_code" jsonschema_description:"The modified version of the code"`
}

var GenerateDiffInputSchema = GenerateSchema[GenerateDiffInput]()

func GenerateDiff(input json.RawMessage) (string, error) {
	diffInput := GenerateDiffInput{}
	err := json.Unmarshal(input, &diffInput)
	if err != nil {
		return "", err
	}

	if diffInput.OriginalCode == diffInput.ModifiedCode {
		return "No changes detected. The original and modified code are identical.", nil
	}

	// Split the code into lines for line-by-line comparison
	originalLines := strings.Split(diffInput.OriginalCode, "\n")
	modifiedLines := strings.Split(diffInput.ModifiedCode, "\n")
	
	// Format the diff for better readability
	var result strings.Builder
	
	// Use a simple line-by-line diff algorithm
	// This is a simplified implementation of the Myers diff algorithm
	lcs := longestCommonSubsequence(originalLines, modifiedLines)
	
	i, j := 0, 0
	for k := 0; k < len(lcs); k++ {
		// Print deletions (lines in original but not in LCS)
		for i < lcs[k].originalIndex {
			result.WriteString(fmt.Sprintf("\u001b[31m- %s\u001b[0m\n", originalLines[i]))
			i++
		}
		
		// Print additions (lines in modified but not in LCS)
		for j < lcs[k].modifiedIndex {
			result.WriteString(fmt.Sprintf("\u001b[32m+ %s\u001b[0m\n", modifiedLines[j]))
			j++
		}
		
		// Print unchanged lines (lines in both)
		result.WriteString(fmt.Sprintf("\u001b[90m  %s\u001b[0m\n", originalLines[i]))
		i++
		j++
	}
	
	// Print any remaining deletions
	for i < len(originalLines) {
		result.WriteString(fmt.Sprintf("\u001b[31m- %s\u001b[0m\n", originalLines[i]))
		i++
	}
	
	// Print any remaining additions
	for j < len(modifiedLines) {
		result.WriteString(fmt.Sprintf("\u001b[32m+ %s\u001b[0m\n", modifiedLines[j]))
		j++
	}
	
	return result.String(), nil
}

type commonLine struct {
	originalIndex int
	modifiedIndex int
}

func longestCommonSubsequence(originalLines, modifiedLines []string) []commonLine {
	modifiedMap := make(map[string][]int)
	for i, line := range modifiedLines {
		modifiedMap[line] = append(modifiedMap[line], i)
	}
	
	var common []commonLine
	for i, line := range originalLines {
		if indices, ok := modifiedMap[line]; ok {
			bestIndex := -1
			for _, j := range indices {
				valid := true
				for _, c := range common {
					if c.originalIndex > i && c.modifiedIndex < j {
						valid = false
						break
					}
				}
				
				if valid && (bestIndex == -1 || bestIndex > j) {
					bestIndex = j
				}
			}
			
			if bestIndex != -1 {
				common = append(common, commonLine{i, bestIndex})
			}
		}
	}
	
	sort.Slice(common, func(i, j int) bool {
		return common[i].originalIndex < common[j].originalIndex
	})
	
	return common
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), tools []ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	
	// For edit_file operations, print the response (which includes the diff) to the console
	if name == "edit_file" {
		fmt.Printf("\u001b[92mresult\u001b[0m: %s\n", response)
	}
	
	return anthropic.NewToolResultBlock(id, response, false)
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
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
			Model:     anthropic.ModelClaude3_7SonnetLatest,
			MaxTokens: int64(1024),
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
			return result.message, result.err
		case <-ticker.C:
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\r\033[K\u001b[93mCooking...\u001b[0m (%.0fs)", elapsed)
		}
	}
}