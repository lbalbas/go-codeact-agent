package llm

import (
	"context"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func Initialize() (client *openai.Client, ctx context.Context) {
	apiKey := os.Getenv("GROQ_API_KEY")

	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable not set")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	client = openai.NewClientWithConfig(config)
	return client, context.Background()
}

const Model = "llama-3.3-70b-versatile"
