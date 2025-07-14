package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	botpkg "telegram-reminder/internal/bot"
)

func TestLoadTasksLogsFallback(t *testing.T) {
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(t.TempDir())

	t.Setenv("TASKS_FILE", "")
	t.Setenv("TASKS_JSON", "")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	_, err := botpkg.LoadTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "tasks.yml not found") {
		t.Errorf("log message missing: %s", buf.String())
	}
}
