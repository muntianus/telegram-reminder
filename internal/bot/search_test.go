package bot

import (
	"context"
	"errors"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

type toolCallClient struct {
	t      *testing.T
	called int
}

func (c *toolCallClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	if c.called == 0 {
		c.called++
		return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{
			Message: openai.ChatCompletionMessage{
				ToolCalls: []openai.ToolCall{{
					ID:   "1",
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      "web_search",
						Arguments: `{"query":"test"}`,
					},
				}},
			},
		}}}, nil
	}
	if len(req.Messages) == 0 {
		c.t.Fatalf("no messages in second request")
	}
	msg := req.Messages[len(req.Messages)-1]
	if msg.Content == "" {
		return openai.ChatCompletionResponse{}, errors.New("empty content")
	}
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{
		Message: openai.ChatCompletionMessage{Content: "ok"},
	}}}, nil
}

func TestChatCompletionEmptySearchResult(t *testing.T) {
	origSearch := searchFunc
	searchFunc = func(ctx context.Context, q string) (string, error) { return "", nil }
	defer func() { searchFunc = origSearch }()
	origWeb := EnableWebSearch
	EnableWebSearch = true
	defer func() { EnableWebSearch = origWeb }()

	client := &toolCallClient{t: t}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}}
	resp, err := ChatCompletion(ctx, client, msgs, "gpt-4.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected response %q", resp)
	}
}
