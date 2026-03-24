package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	// Load the .env file before reading environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it, continuing with environment variables")
	}

	//Get input
	fmt.Println("Enter your prompt:")
	var input string
	fmt.Scanln(&input)

	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	//Process input, call Gemini API
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text("You're an AI CodeAct agent, reply to the prompt solving it using exclusively bash scripts. Output ONLY a bash script wrapped in ```bash  ... ```.\n Prompt: "+input),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Text())
}
