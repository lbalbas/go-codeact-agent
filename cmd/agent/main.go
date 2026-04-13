package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"

	"github.com/lbalbas/go-codeact-agent/internal/agent"
	"github.com/lbalbas/go-codeact-agent/internal/llm"
)

func main() {
	// Load the .env file before reading environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it, continuing with environment variables")
	}

	input := getPrompt()
	client, ctx := llm.Initialize()

	contents := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: agent.SystemPrompt + input,
		},
	}

	agent.RunLoop(ctx, client, contents)

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
