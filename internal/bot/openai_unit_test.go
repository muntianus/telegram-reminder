package bot

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

type fakeClient struct{ resp string }

func (f fakeClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: f.resp}}}}, nil
}

func TestChatCompletion(t *testing.T) {
	ModelMu.Lock()
	CurrentModel = "gpt-4o"
	ModelMu.Unlock()
	client := fakeClient{resp: "answer"}
	msg := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "hi"}
	resp, err := ChatCompletion(context.Background(), client, []openai.ChatCompletionMessage{msg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "answer" {
		t.Errorf("unexpected resp: %v", resp)
	}
}

func TestSystemCompletion(t *testing.T) {
	client := fakeClient{resp: "sys"}
	resp, err := SystemCompletion(context.Background(), client, "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "sys" {
		t.Errorf("unexpected resp: %v", resp)
	}
}

func TestUserCompletion(t *testing.T) {
	client := fakeClient{resp: "user"}
	resp, err := UserCompletion(context.Background(), client, "msg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "user" {
		t.Errorf("unexpected resp: %v", resp)
	}
}
