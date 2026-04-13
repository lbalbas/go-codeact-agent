package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

// handleReadFileContents reads a file at the path specified in the JSON args.
func handleReadFileContents(args string) string {
	var input struct {
		Filepath string `json:"filepath"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return fmt.Sprintf("Error parsing arguments: %v", err)
	}

	bytes, err := os.ReadFile(input.Filepath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	return string(bytes)
}
