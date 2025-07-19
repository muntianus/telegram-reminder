package bot

import (
	"testing"
)

func TestFormatTasks(t *testing.T) {
	tasks := []Task{
		{Name: "a", Time: "10:00"},
		{Name: "b", Cron: "0 5 * * *"},
		{},
	}
	out := FormatTasks(tasks)
	if out == "" || out == "no tasks" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormatTaskNames(t *testing.T) {
	tasks := []Task{{Name: "a"}, {Name: "b"}, {}}
	out := FormatTaskNames(tasks)
	if out != "a\nb" {
		t.Errorf("unexpected output: %q", out)
	}
	if FormatTaskNames(nil) != "no tasks" {
		t.Errorf("expected 'no tasks' for nil slice")
	}
}

func TestFindTask(t *testing.T) {
	tasks := []Task{{Name: "a"}, {Name: "b"}}
	foundTask, ok := FindTask(tasks, "b")
	if !ok || foundTask.Name != "b" {
		t.Errorf("not found or wrong task: %+v", foundTask)
	}
	_, ok = FindTask(tasks, "c")
	if ok {
		t.Errorf("should not find non-existent task")
	}
}
