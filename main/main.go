package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	// Load the .env file before reading environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it, continuing with environment variables")
	}

	input := getPrompt()
	client, ctx := initializeClient()

	prompt := `You are an expert Senior Software Engineer and Security Auditor acting as an Automated Code Reviewer. Your task is to analyze code changes and provide a comprehensive, constructive code review.
				Use your provided native tools to read files and get git diffs. If the native tools are insufficient to understand the workspace, you can output powershell scripts wrapped in ` + "```powershell ... ```" + ` to execute local commands (like linters or formatting tools).
				Always start by requesting the git diff (using your tool or a script) if it is not provided in the prompt.
				Analyze the code for:
				- Logic flaws or potential bugs
				- Security vulnerabilities
				- Best practices and code style
				- Performance bottlenecks
				If you write a powershell script, I will execute it locally and send you the output in our loop.
				Once you have retrieved the necessary information and completed your review, provide a final structured review in Markdown format. If you can provide an automated fix, output the fixing powershell script.
				Once the task is fully completed and your final review is delivered, output the exact string [DONE] on its own line to end the session.
				Prompt: ` + input

	contents := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}
	toolsList := getToolConfig()

	for {
		fmt.Println("Calling API")
		req := openai.ChatCompletionRequest{
			Model:    "llama-3.3-70b-versatile",
			Messages: contents,
			Tools:    toolsList,
		}

		result, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			log.Fatal(err)
		}

		msg := result.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			// Append the assistant's tool call message
			contents = append(contents, msg)
			
			for _, call := range msg.ToolCalls {
				fmt.Println("Tool Call:", call.Function.Name)
				if handler, ok := tools[call.Function.Name]; ok {
					callResult := handler(call.Function.Arguments)
					// Append the tool result
					contents = append(contents, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						Content:    callResult,
						Name:       call.Function.Name,
						ToolCallID: call.ID,
					})
				} else {
					log.Fatal("Unknown tool: " + call.Function.Name)
				}
			}
		} else {
			text := msg.Content
			if strings.Contains(text, "[DONE]") {
				// Print the final review (strip the [DONE] marker) before exiting
				finalText := strings.Replace(text, "[DONE]", "", -1)
				finalText = strings.TrimSpace(finalText)
				if finalText != "" {
					fmt.Println(finalText)
				}
				break
			}

			fmt.Println(text)

			// Only attempt script execution if the response contains a powershell code block
			if strings.Contains(text, "```powershell") || strings.Contains(text, "```ps") {
				fmt.Println("Executing script")
				output := executeScript(cleanScript(text))
				fmt.Println("Output from previous script: " + output)

				// Append assistant's text msg and the simulated user output
				contents = append(contents, msg)
				contents = append(contents, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "Output from previous script: " + output,
				})
			} else {
				// No script to execute, just append and continue the conversation
				contents = append(contents, msg)
				contents = append(contents, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "Continue with the review. Remember to output [DONE] when finished.",
				})
			}
		}
		
		fmt.Println("Waiting 1 minute to avoid API rate limits...")
		time.Sleep(1 * time.Minute)
	}
	fmt.Println("DONE")
}

func getPrompt() string {
	fmt.Println("Enter your prompt:")

	var input string
	var err error

	reader := bufio.NewReader(os.Stdin)
	input, err = reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
		return ""
	}

	input = strings.TrimSpace(input)
	return input
}

func initializeClient() (client *openai.Client, ctx context.Context) {
	apiKey := os.Getenv("GROQ_API_KEY")

	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable not set")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	client = openai.NewClientWithConfig(config)
	return client, context.Background()
}

func cleanScript(script string) string {
	// Trim whitespace first
	script = strings.TrimSpace(script)

	// Check if wrapped in markdown code blocks
	if strings.HasPrefix(script, "```") && strings.HasSuffix(script, "```") {
		// Remove the first line (```powershell or ```)
		lines := strings.Split(script, "\n")
		if len(lines) > 2 {
			// Remove first and last line
			script = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	return strings.TrimSpace(script)
}

// executeScript runs the command and returns combined output (stdout + stderr)
func executeScript(script string) string {
	if script == "" {
		return "Error: No script found to execute."
	}

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.CombinedOutput()

	// If there is no output but the command succeeded, let the AI know
	if len(output) == 0 && err == nil {
		return "Script executed successfully (no output)."
	}

	return string(output)
}

func getToolConfig() []openai.Tool {
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

var tools = map[string]func(string) string{
	"get_git_diff":       handleGetGitDiff,
	"read_file_contents": handleReadFileContents,
}

func handleGetGitDiff(args string) string {
	cmd := exec.Command("git", "diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error getting git diff: %v", err)
	}
	return string(output)
}

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
