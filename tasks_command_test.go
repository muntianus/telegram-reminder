package main

import (
	"testing"

	tb "gopkg.in/telebot.v3"
	botpkg "telegram-reminder/internal/bot"
)

type listCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

func (l *listCtx) Send(what interface{}, opts ...interface{}) error {
	l.called = true
	l.msg = what
	return nil
}

func TestTasksCommand(t *testing.T) {
	botpkg.TasksMu.Lock()
	botpkg.LoadedTasks = []botpkg.Task{{Name: "a", Time: "10:00"}, {Name: "b", Cron: "0 5 * * *"}}
	botpkg.TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	b.Handle("/tasks", func(c tb.Context) error {
		botpkg.TasksMu.RLock()
		tasks := append([]botpkg.Task(nil), botpkg.LoadedTasks...)
		botpkg.TasksMu.RUnlock()
		return c.Send(botpkg.FormatTasks(tasks))
	})

	ctx := &listCtx{}
	if err := b.Trigger("/tasks", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	want := "10:00 - a\n0 5 * * * - b"
	if ctx.msg != want {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}
