package tools

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Definitions returns the tool schemas that get sent to the LLM,
// telling it what functions it can call and what arguments they expect.
func Definitions() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_git_diff",
				Description: "Returns the git diff for the current local repository. Call this to see the unstaged/staged changes.",
				Parameters: jsonschema.Definition{
					Type:       jsonschema.Object,
					Properties: map[string]jsonschema.Definition{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file_contents",
				Description: "Reads the contents of a file at the given filepath.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"filepath": {
							Type:        jsonschema.String,
							Description: "The relative or absolute path of the file to read.",
						},
					},
					Required: []string{"filepath"},
				},
			},
		},
	}
}

// Handlers returns the dispatch map that connects tool names (the same names
// from Definitions) to the Go functions that actually execute them.
func Handlers() map[string]func(string) string {
	return map[string]func(string) string{
		"get_git_diff":       handleGetGitDiff,
		"read_file_contents": handleReadFileContents,
	}
}
