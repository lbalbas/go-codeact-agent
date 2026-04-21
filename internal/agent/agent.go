package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/lbalbas/go-codeact-agent/internal/llm"
	"github.com/lbalbas/go-codeact-agent/internal/script"
	"github.com/lbalbas/go-codeact-agent/internal/tools"
)

func RunLoop(ctx context.Context, client *openai.Client, contents []openai.ChatCompletionMessage) {
	toolsList := tools.Definitions()
	handlers := tools.Handlers()

	for {
		req := openai.ChatCompletionRequest{
			Model:    llm.Model,
			Messages: contents,
			Tools:    toolsList,
		}

		slog.Info("Calling API", "model", req.Model)

		result, err := callWithRetry(ctx, client, req)
		if err != nil {
			log.Fatal("API error: ", err)
		}

		msg := result.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			// Append the assistant's tool call message
			contents = append(contents, msg)

			for _, call := range msg.ToolCalls {
				slog.Info("Executing tool", "tool_name", call.Function.Name)
				if handler, ok := handlers[call.Function.Name]; ok {
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

			// Only attempt script execution if the response contains a powershell code block
			if strings.Contains(text, "```powershell") || strings.Contains(text, "```ps") || strings.Contains(text, "```ps1") {
				slog.Debug("Executing extracted powershell script")
				scripts := script.ExtractScripts(text)
				scriptBody := strings.Join(scripts, "\n")
				if scriptBody == "" {
					contents = append(contents, msg)
					contents = append(contents, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: "Error: Extracted PowerShell script block was empty. Please provide a valid script.",
					})
				} else {
					output := script.Execute(scriptBody, 30)
					slog.Info("Output from previous script", "output", output)

					// Append assistant's text msg and the simulated user output
					contents = append(contents, msg)
					contents = append(contents, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: "Output from previous script: " + output,
					})
				}
			} else if strings.Contains(text, "```") {
				contents = append(contents, msg)
				contents = append(contents, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "Error: I noticed you attempted to run commands, but you used the wrong block format. You MUST use strictly ```powershell ... ``` to execute commands in this environment.",
				})
			} else {
				// No script to execute, just append and continue the conversation
				contents = append(contents, msg)
				contents = append(contents, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "Continue with the review. Remember, if you need to execute commands, output them strictly in a ```powershell block. And remember to output [DONE] when finished.",
				})
			}
		}

		slog.Info("Waiting 1 minute to avoid API rate limits...")
		time.Sleep(1 * time.Minute)
	}
}

func callWithRetry(ctx context.Context, client *openai.Client, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	maxRetries := 5
	waitDelay := 20 * time.Second
	for i := 0; i < maxRetries; i++ {
		result, err := client.CreateChatCompletion(ctx, req)
		if err == nil {
			return &result, nil
		}
		// Check if the error is a Rate Limit (429)
		var apiErr *openai.APIError
		if errors.As(err, &apiErr) && apiErr.HTTPStatusCode == 429 {
			slog.Warn("Rate limit hit, retrying...", "attempt", i+1, "wait", waitDelay)
			time.Sleep(waitDelay)
			waitDelay *= 2 // Exponential backoff: 2s, 4s, 8s...
			continue
		}
		// If it's some other error, don't retry
		return nil, err
	}
	return nil, fmt.Errorf("max retries exceeded")
}
