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

	var contents []*genai.Content
	contents = append(contents, genai.NewContentFromText("You're an AI CodeAct agent, reply to the prompt solving it using exclusively powershell commands. Output ONLY a powershell script/commands wrapped in ```powershell  ... ```. I'll send you the output of your script execution in a loop so you can verify the result and output more scripts to execute if needed. Once the result or the prompt has been properly completed, wait for the output of the final script, check if it's the expected then output [DONE] to end the session.\n Prompt: "+input, genai.RoleUser))

	fmt.Println("Calling API")
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		contents,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	for !strings.Contains(result.Text(), "[DONE]") {
		fmt.Println(result.Text())
		fmt.Println("Executing script")
		output := executeScript(cleanScript(result.Text()))
		fmt.Println("Output from previous script: " + output)
		fmt.Println("Calling API with output")
		contents = append(contents, genai.NewContentFromText(result.Text(), genai.RoleModel), genai.NewContentFromText("Output from previous script: "+string(output), genai.RoleUser))
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
