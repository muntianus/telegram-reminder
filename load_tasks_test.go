package main

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	botpkg "telegram-reminder/internal/bot"
	"telegram-reminder/internal/logger"
)

func TestLoadTasksLogsFallback(t *testing.T) {
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(t.TempDir())

	t.Setenv("TASKS_FILE", "")
	t.Setenv("TASKS_JSON", "")

	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, nil))
	logger.SetLogger(l)
	t.Cleanup(func() { logger.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))) })

	_, err := botpkg.LoadTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "tasks.yml not found") {
		t.Errorf("log message missing: %s", buf.String())
	}
}
