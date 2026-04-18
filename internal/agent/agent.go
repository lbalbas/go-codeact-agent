package agent

import (
	"context"
	"fmt"
	"log"
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
		fmt.Println("Calling API")
		req := openai.ChatCompletionRequest{
			Model:    llm.Model,
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

			fmt.Println(text)

			// Only attempt script execution if the response contains a powershell code block
			if strings.Contains(text, "```powershell") || strings.Contains(text, "```ps") || strings.Contains(text, "```ps1") {
				fmt.Println("Executing script")
				scripts := script.ExtractScripts(text)
				scriptBody := strings.Join(scripts, "\n")
				output := script.Execute(scriptBody, 30)
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
}
