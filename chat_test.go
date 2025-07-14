package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type fakeChatClient struct{ text string }

func (f fakeChatClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: f.text}}}}, nil
}

func TestChatRepliesToGroup(t *testing.T) {
	client := fakeChatClient{text: "hi"}

	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`)); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer srv.Close()

	bot, err := tb.NewBot(tb.Settings{Token: "T", URL: srv.URL, Client: srv.Client(), Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/chat", func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), openAITimeout)
		defer cancel()

		text, err := userCompletion(ctx, client, q)
		if err != nil {
			return c.Send("OpenAI error")
		}

		to := tb.Recipient(c.Sender())
		if !c.Message().Private() {
			to = c.Chat()
		}
		_, err = c.Bot().Send(to, text)
		return err
	})

	group := &tb.Chat{ID: -99, Type: tb.ChatGroup}
	msg := &tb.Message{Chat: group, Sender: &tb.User{ID: 1}, Payload: "hello"}

	ctx := bot.NewContext(tb.Update{Message: msg})
	if err := bot.Trigger("/chat", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["chat_id"] != "-99" {
		t.Errorf("sent to wrong chat: %v", got["chat_id"])
	}
}
