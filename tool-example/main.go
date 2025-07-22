package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Сделай поиск цены BTC",
		},
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: messages,
		Tools: []openai.Tool{
			{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name: "search",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]interface{}{
								"type": "string",
							},
						},
						"required": []string{"query"},
					},
				},
			},
		},
		ToolChoice: "auto",
	})
	if err != nil {
		panic(err)
	}

	if len(resp.Choices) > 0 && len(resp.Choices[0].Message.ToolCalls) > 0 {
		toolCall := resp.Choices[0].Message.ToolCalls[0]

		toolResult := callSearch(toolCall.Function.Arguments)

		toolResponse := openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: toolCall.ID,
			Content:    toolResult,
		}

		messages = append(messages, resp.Choices[0].Message)
		messages = append(messages, toolResponse)

		finalResp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    openai.GPT4o,
			Messages: messages,
		})
		if err != nil {
			panic(err)
		}

		fmt.Println(finalResp.Choices[0].Message.Content)
	}
}

func callSearch(args string) string {
	var p struct {
		Query string `json:"query"`
	}
	json.Unmarshal([]byte(args), &p)
	if p.Query == "btc price" || p.Query == "цена btc" {
		return "BTC сейчас стоит $62,000"
	}
	return "ничего не нашёл"
}
