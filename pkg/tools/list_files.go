package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
		return "", err
	}

	path := "."
	if listFilesInput.Path != "" {
		path = listFilesInput.Path
	}

	// Check if the path exists
	_, err = os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	// List files and directories
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	// Format the output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of %s:\n\n", path))

	// Separate directories and files
	var dirs []string
	var files []string

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			dirs = append(dirs, name+"/")
		} else {
			files = append(files, name)
		}
	}

	// Print directories first
	if len(dirs) > 0 {
		result.WriteString("Directories:\n")
		for _, dir := range dirs {
			result.WriteString(fmt.Sprintf("  %s\n", dir))
		}
		result.WriteString("\n")
	}

	// Then print files
	if len(files) > 0 {
		result.WriteString("Files:\n")
		for _, file := range files {
			// Get file info for size
			info, err := os.Stat(filepath.Join(path, file))
			if err == nil {
				result.WriteString(fmt.Sprintf("  %s (%d bytes)\n", file, info.Size()))
			} else {
				result.WriteString(fmt.Sprintf("  %s\n", file))
			}
		}
	}

	if len(dirs) == 0 && len(files) == 0 {
		result.WriteString("Directory is empty.\n")
	}

	return result.String(), nil
}
