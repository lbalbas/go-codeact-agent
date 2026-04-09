package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
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

	var contents []*genai.Content
	contents = append(contents, genai.NewContentFromText(prompt, genai.RoleUser))
	config := getToolConfig()
	fmt.Println("Calling API")
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		contents,
		config,
	)

	if err != nil {
		log.Fatal(err)
	}

	for {
		if len(result.FunctionCalls()) > 0 {
			fmt.Println(result.FunctionCalls())
			for _, call := range result.FunctionCalls() {
				if handler, ok := tools[call.Name]; ok {
					callResult := handler(call)
					contents = append(contents, genai.NewContentFromFunctionResponse(call.Name, map[string]any{"content": callResult}, genai.RoleUser))
					result, err = client.Models.GenerateContent(
						ctx,
						"gemini-3-flash-preview",
						contents,
						nil,
					)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					log.Fatal("Unknown tool: " + call.Name)
				}
			}
		} else {
			text := result.Text()
			if strings.Contains(text, "[DONE]") {
				break
			}

			fmt.Println(text)
			fmt.Println("Executing script")
			output := executeScript(cleanScript(text))
			fmt.Println("Output from previous script: " + output)
			fmt.Println("Calling API with output")
			contents = append(contents, genai.NewContentFromText(text, genai.RoleModel), genai.NewContentFromText("Output from previous script: "+string(output), genai.RoleUser))
			result, err = client.Models.GenerateContent(
				ctx,
				"gemini-3-flash-preview",
				contents,
				nil,
			)
			if err != nil {
				log.Fatal(err)
			}
		}
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

func initializeClient() (client *genai.Client, ctx context.Context) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	ctx = context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}
	return client, ctx
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

func getToolConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						// No parameters needed, we just run git diff
						Name:        "get_git_diff",
						Description: "Returns the git diff for the current local repository. Call this to see the unstaged/staged changes.",
					},
					{
						// Parameters needed for reading specific files
						Name:        "read_file_contents",
						Description: "Reads the contents of a file at the given filepath.",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"filepath": {
									Type:        genai.TypeString,
									Description: "The relative or absolute path of the file to read.",
								},
							},
							Required: []string{"filepath"},
						},
					},
				},
			},
		},
	}
}

var tools = map[string]func(*genai.FunctionCall) string{
	"get_git_diff":       handleGetGitDiff,
	"read_file_contents": handleReadFileContents,
}

func handleGetGitDiff(call *genai.FunctionCall) string {
	cmd := exec.Command("git", "diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error getting git diff: %v", err)
	}
	return string(output)
}

func handleReadFileContents(call *genai.FunctionCall) string {
	filepath := call.Args["filepath"].(string)
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	return string(bytes)
}
