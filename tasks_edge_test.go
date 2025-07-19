package main

import (
	"os"
	"telegram-reminder/internal/bot"
	"testing"
)

func TestLoadTasks_BadYAML(t *testing.T) {
	dir := t.TempDir()
	badFile := dir + "/bad.yml"
	os.WriteFile(badFile, []byte("not: [valid: yaml"), 0644)
	t.Setenv("TASKS_FILE", badFile)
	_, err := bot.LoadTasks()
	if err == nil {
		t.Fatal("expected error for bad yaml")
	}
}

func TestLoadTasks_BadJSON(t *testing.T) {
	t.Setenv("TASKS_JSON", "{not: valid json}")
	_, err := bot.LoadTasks()
	if err == nil {
		t.Fatal("expected error for bad json")
	}
}

func TestLoadTasks_NoPerm(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/tasks.yml"
	os.WriteFile(file, []byte("- name: test\n  prompt: hi"), 0000)
	t.Setenv("TASKS_FILE", file)
	_, err := bot.LoadTasks()
	if err == nil {
		t.Fatal("expected error for no permissions")
	}
}
