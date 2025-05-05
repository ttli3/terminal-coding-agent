package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
