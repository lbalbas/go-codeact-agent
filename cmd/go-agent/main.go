package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"

	"github.com/lbalbas/go-codeact-agent/internal/agent"
	"github.com/lbalbas/go-codeact-agent/internal/llm"
)

func main() {
	// Try loading .env from current directory
	_ = godotenv.Load()

	// Fallback to home directory if local .env doesn't satisfy requirements
	if os.Getenv("GROQ_API_KEY") == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			_ = godotenv.Load(home + "/.go-agent.env")
		}
	}

	if os.Getenv("GROQ_API_KEY") == "" {
		log.Println("Note: GROQ_API_KEY not found in .env or ~/.go-agent.env. Ensure it is set in your environment variables.")
	}

	input, err := getPrompt()
	if err != nil {
		log.Fatal(err)
		return
	}
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

func getPrompt() (string, error) {
	fmt.Println("Enter your prompt:")

	var input string
	var err error

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err = reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if err != nil {
			if err == io.EOF {
				if input == "" {
					return "", io.EOF
				}
			} else {
				return "", err
			}
		}

		if input == "" {
			fmt.Println("Prompt cannot be empty, please enter a valid prompt:")
			continue
		}

		return input, nil
	}
}
