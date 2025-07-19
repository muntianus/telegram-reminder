package main

import (
	"os"
	"path/filepath"
	"testing"

	botpkg "telegram-reminder/internal/bot"
)

func TestFormatTasksEmpty(t *testing.T) {
	tasks := []botpkg.Task{}
	result := botpkg.FormatTasks(tasks)
	if result != "no tasks" {
		t.Errorf("expected 'no tasks', got %q", result)
	}
}

func TestFormatTasksWithEmptyNames(t *testing.T) {
	tasks := []botpkg.Task{
		{Name: "", Time: "10:00"},
		{Name: "task1", Time: "11:00"},
		{Name: "", Cron: "0 12 * * *"},
	}
	result := botpkg.FormatTasks(tasks)
	expected := "10:00 - task 1\n11:00 - task1\n0 12 * * * - task 3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatTasksWithDefaultTime(t *testing.T) {
	tasks := []botpkg.Task{
		{Name: "task1", Time: ""},
		{Name: "task2", Cron: ""},
	}
	result := botpkg.FormatTasks(tasks)
	expected := "00:00 - task1\n00:00 - task2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatTaskNamesEmpty(t *testing.T) {
	tasks := []botpkg.Task{}
	result := botpkg.FormatTaskNames(tasks)
	if result != "no tasks" {
		t.Errorf("expected 'no tasks', got %q", result)
	}
}

func TestFormatTaskNamesWithEmptyNames(t *testing.T) {
	tasks := []botpkg.Task{
		{Name: "", Time: "10:00"},
		{Name: "task1", Time: "11:00"},
		{Name: "", Cron: "0 12 * * *"},
		{Name: "task2", Time: "13:00"},
	}
	result := botpkg.FormatTaskNames(tasks)
	expected := "task1\ntask2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFindTaskNotFound(t *testing.T) {
	tasks := []botpkg.Task{
		{Name: "task1", Time: "10:00"},
		{Name: "task2", Time: "11:00"},
	}
	task, found := botpkg.FindTask(tasks, "nonexistent")
	if found {
		t.Error("expected not found")
	}
	if task.Name != "" {
		t.Errorf("expected empty task, got %q", task.Name)
	}
}

func TestFindTaskCaseSensitive(t *testing.T) {
	tasks := []botpkg.Task{
		{Name: "Task1", Time: "10:00"},
		{Name: "task2", Time: "11:00"},
	}
	task, found := botpkg.FindTask(tasks, "task1")
	if found {
		t.Error("expected not found due to case sensitivity")
	}
	if task.Name != "" {
		t.Errorf("expected empty task, got %q", task.Name)
	}
}

func TestFindTaskEmptyName(t *testing.T) {
	tasks := []botpkg.Task{
		{Name: "", Time: "10:00"},
		{Name: "task1", Time: "11:00"},
	}
	task, found := botpkg.FindTask(tasks, "")
	if !found {
		t.Error("expected to find task with empty name")
	}
	if task.Time != "10:00" {
		t.Errorf("expected task with Time '10:00', got %q", task.Time)
	}
}

func TestLoadTasksWithInvalidJSON(t *testing.T) {
	t.Setenv("TASKS_JSON", `[{"name": "task1", "prompt": "test"}`) // Missing closing bracket

	_, err := botpkg.LoadTasks()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadTasksWithEmptyJSON(t *testing.T) {
	t.Setenv("TASKS_JSON", "[]")

	tasks, err := botpkg.LoadTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty tasks, got %d", len(tasks))
	}
}

func TestLoadTasksWithInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	invalidYAML := filepath.Join(dir, "invalid.yml")

	err := os.WriteFile(invalidYAML, []byte("invalid: yaml: content: [:"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	t.Setenv("TASKS_FILE", invalidYAML)

	_, err = botpkg.LoadTasks()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadTasksWithNonExistentFile(t *testing.T) {
	t.Setenv("TASKS_FILE", "/nonexistent/file.yml")

	_, err := botpkg.LoadTasks()
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}
