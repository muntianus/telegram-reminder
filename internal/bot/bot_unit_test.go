package bot

import (
	"context"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type fakeOpenAI struct{ resp string }

func (f fakeOpenAI) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: f.resp}}}}, nil
}

type testCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

func (t *testCtx) Send(what interface{}, opts ...interface{}) error {
	t.called = true
	t.msg = what
	return nil
}

func TestRegisterTaskCommands_Unit(t *testing.T) {
	TasksMu.Lock()
	LoadedTasks = []Task{{Name: "foo", Prompt: "bar"}}
	TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}
	client := fakeOpenAI{resp: "ok"}
	RegisterTaskCommands(b, client)

	ctx := &testCtx{}
	if err := b.Trigger("/foo", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "ok" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestScheduleDailyMessages_Unit(t *testing.T) {
	TasksMu.Lock()
	LoadedTasks = []Task{{Name: "foo", Prompt: "bar", Time: "00:00"}}
	TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}
	client := fakeOpenAI{resp: "ok"}

	s := gocron.NewScheduler(time.UTC)
	ScheduleDailyMessages(s, client, b, 0)

	// Проверяем, что задача зарегистрирована
	jobs := s.Jobs()
	if len(jobs) == 0 {
		t.Fatal("no jobs scheduled")
	}
}
