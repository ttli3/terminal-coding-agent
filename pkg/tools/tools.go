package tools

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

// ToolDefinition defines a tool that can be used by the agent
type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

// commonLine represents a line that appears in both the original and modified code
type commonLine struct {
	originalIndex int
	modifiedIndex int
}

// longestCommonSubsequence finds the longest common subsequence of lines between the original and modified code
func longestCommonSubsequence(originalLines, modifiedLines []string) []commonLine {
	// Create a 2D table to store the length of LCS
	m, n := len(originalLines), len(modifiedLines)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	
	// Fill the dp table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if originalLines[i-1] == modifiedLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	
	// Backtrack to find the common lines
	var result []commonLine
	i, j := m, n
	for i > 0 && j > 0 {
		if originalLines[i-1] == modifiedLines[j-1] {
			result = append([]commonLine{{originalIndex: i-1, modifiedIndex: j-1}}, result...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	
	return result
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GenerateSchema generates a JSON schema for the given type
func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		DoNotReference: true,
	}
	var t T
	schema := reflector.Reflect(t)
	schemaBytes, _ := json.Marshal(schema)
	var schemaMap map[string]interface{}
	_ = json.Unmarshal(schemaBytes, &schemaMap)
	
	// Convert to the expected format
	return anthropic.ToolInputSchemaParam{
		Type:       "object",
		Properties: schemaMap["properties"].(map[string]interface{}),
	}
}

// GetAllTools returns all the tool definitions
func GetAllTools() []ToolDefinition {
	return []ToolDefinition{
		ReadFileDefinition,
		ListFilesDefinition, 
		EditFileDefinition, 
		RunCommandDefinition, 
		GenerateDiffDefinition,
	}
}
