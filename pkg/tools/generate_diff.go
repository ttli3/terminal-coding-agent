package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

var GenerateDiffDefinition = ToolDefinition{
	Name: "generate_diff",
	Description: `Generate a diff between two versions of code.
	
This tool helps visualize the differences between an original version of code and a modified version.
It will highlight additions, deletions, and unchanged lines.
`,
	InputSchema: GenerateDiffInputSchema,
	Function:    GenerateDiff,
}

type GenerateDiffInput struct {
	OriginalCode string `json:"original_code" jsonschema_description:"The original version of the code"`
	ModifiedCode string `json:"modified_code" jsonschema_description:"The modified version of the code"`
}

var GenerateDiffInputSchema = GenerateSchema[GenerateDiffInput]()

func GenerateDiff(input json.RawMessage) (string, error) {
	generateDiffInput := GenerateDiffInput{}
	err := json.Unmarshal(input, &generateDiffInput)
	if err != nil {
		return "", err
	}

	if generateDiffInput.OriginalCode == generateDiffInput.ModifiedCode {
		return "No differences found. The original and modified code are identical.", nil
	}

	// Split the code into lines for line-by-line comparison
	originalLines := strings.Split(generateDiffInput.OriginalCode, "\n")
	modifiedLines := strings.Split(generateDiffInput.ModifiedCode, "\n")
	
	// Format the diff for better readability
	var diffResult strings.Builder
	diffResult.WriteString("Diff:\n")
	
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
	
	return diffResult.String(), nil
}
