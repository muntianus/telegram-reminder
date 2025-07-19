package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type fakeClientHandlers struct{ resp string }

func (f fakeClientHandlers) CreateChatCompletion(ctx context.Context, req interface{}) (interface{}, error) {
	return f.resp, nil
}

// testCtx теперь реализует Chat()
type testCtxHandlers struct {
	called  bool
	msg     interface{}
	chatID  int64
	payload string
	bot     *tb.Bot
}

func (c *testCtxHandlers) Send(what interface{}, opts ...interface{}) error {
	c.called = true
	c.msg = what
	return nil
}

func (c *testCtxHandlers) SendAlbum(_ tb.Album, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) Chat() *tb.Chat {
	return &tb.Chat{ID: c.chatID}
}

func (c *testCtxHandlers) Message() *tb.Message {
	return &tb.Message{Payload: c.payload}
}

func (c *testCtxHandlers) Accept(...string) error { return nil }

func (c *testCtxHandlers) Answer(_ *tb.QueryResponse) error { return nil }

func (c *testCtxHandlers) Args() []string { return nil }

func (c *testCtxHandlers) Boost() *tb.BoostUpdated { return nil }

func (c *testCtxHandlers) BoostRemoved() *tb.BoostRemoved { return nil }

func (c *testCtxHandlers) Bot() *tb.Bot { return c.bot }

func (c *testCtxHandlers) Callback() *tb.Callback { return nil }

func (c *testCtxHandlers) ChatJoinRequest() *tb.ChatJoinRequest { return nil }

func (c *testCtxHandlers) ChatMember() *tb.ChatMemberUpdate { return nil }

func (c *testCtxHandlers) Data() string { return "" }

func (c *testCtxHandlers) Delete() error { return nil }

func (c *testCtxHandlers) DeleteAfter(_ time.Duration) *time.Timer { return nil }

func (c *testCtxHandlers) Edit(_ interface{}, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) EditCaption(_ string, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) EditOrReply(_ interface{}, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) EditOrSend(_ interface{}, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) Entities() tb.Entities { return nil }

func (c *testCtxHandlers) Forward(_ tb.Editable, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) ForwardTo(_ tb.Recipient, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) Get(_ string) interface{} { return nil }

func (c *testCtxHandlers) InlineResult() *tb.InlineResult { return nil }

func (c *testCtxHandlers) Notify(_ tb.ChatAction) error { return nil }

func (c *testCtxHandlers) Poll() *tb.Poll { return nil }

func (c *testCtxHandlers) PollAnswer() *tb.PollAnswer { return nil }

func (c *testCtxHandlers) PreCheckoutQuery() *tb.PreCheckoutQuery { return nil }

func (c *testCtxHandlers) Query() *tb.Query { return nil }

func (c *testCtxHandlers) Recipient() tb.Recipient { return nil }

func (c *testCtxHandlers) Reply(_ interface{}, _ ...interface{}) error { return nil }

func (c *testCtxHandlers) Respond(...*tb.CallbackResponse) error { return nil }

func (c *testCtxHandlers) RespondAlert(_ string) error { return nil }

func (c *testCtxHandlers) RespondText(_ string) error { return nil }

func (c *testCtxHandlers) Sender() *tb.User { return &tb.User{ID: 1} }

func (c *testCtxHandlers) Set(string, interface{}) {}

func (c *testCtxHandlers) Ship(...interface{}) error { return nil }

func (c *testCtxHandlers) ShippingQuery() *tb.ShippingQuery { return nil }

func (c *testCtxHandlers) Text() string { return "" }

func (c *testCtxHandlers) Update() tb.Update { return tb.Update{} }

type Topic = tb.Topic

func (c *testCtxHandlers) Topic() *tb.Topic { return nil }

type Migration struct{}

func (c *testCtxHandlers) Migration() (int64, int64) { return 0, 0 }

type fakeOpenAIError struct{}

func (f fakeOpenAIError) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{}, errors.New("fail")
}

type fakeOpenAISuccess struct{ resp string }

func (f fakeOpenAISuccess) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: f.resp}}}}, nil
}

func TestHandlers_Lunch_Brief_Chat_CompletionOnly(t *testing.T) {
	client := fakeOpenAISuccess{resp: "ok"}

	// lunch
	resp, err := SystemCompletion(context.Background(), client, LunchIdeaPrompt)
	if err != nil || resp != "ok" {
		t.Errorf("lunch: got %v, err %v", resp, err)
	}

	// brief
	resp, err = SystemCompletion(context.Background(), client, DailyBriefPrompt)
	if err != nil || resp != "ok" {
		t.Errorf("brief: got %v, err %v", resp, err)
	}

	// chat
	resp, err = UserCompletion(context.Background(), client, "hello")
	if err != nil || resp != "ok" {
		t.Errorf("chat: got %v, err %v", resp, err)
	}
}

func TestHandlers_Lunch_Brief_Chat_CompletionError(t *testing.T) {
	client := fakeOpenAIError{}

	// lunch
	resp, err := SystemCompletion(context.Background(), client, LunchIdeaPrompt)
	if err == nil {
		t.Errorf("lunch: expected error, got %v", resp)
	}

	// brief
	resp, err = SystemCompletion(context.Background(), client, DailyBriefPrompt)
	if err == nil {
		t.Errorf("brief: expected error, got %v", resp)
	}

	// chat
	resp, err = UserCompletion(context.Background(), client, "hello")
	if err == nil {
		t.Errorf("chat: expected error, got %v", resp)
	}
}
