package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Task struct is defined in bot.go

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// readTasksFile loads tasks from a YAML or JSON file.
func readTasksFile(fn string) ([]Task, string, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, "", err
	}
	tasks := []Task{}
	ext := strings.ToLower(filepath.Ext(fn))
	var tf struct {
		BasePrompt string `json:"base_prompt" yaml:"base_prompt"`
		Tasks      []Task `json:"tasks" yaml:"tasks"`
	}
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &tf); err == nil && len(tf.Tasks) > 0 {
			return tf.Tasks, tf.BasePrompt, nil
		}
		if err := yaml.Unmarshal(data, &tasks); err != nil {
			return nil, "", err
		}
	} else {
		if err := json.Unmarshal(data, &tf); err == nil && len(tf.Tasks) > 0 {
			return tf.Tasks, tf.BasePrompt, nil
		}
		if err := json.Unmarshal(data, &tasks); err != nil {
			return nil, "", err
		}
	}
	return tasks, "", nil
}

// LoadTasks reads task configuration from TASKS_FILE or TASKS_JSON. If neither
// is provided, it falls back to tasks.yml or the legacy LUNCH_TIME and
// BRIEF_TIME environment variables.
func LoadTasks() ([]Task, error) {
	if fn := os.Getenv("TASKS_FILE"); fn != "" {
		tasks, bp, err := readTasksFile(fn)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		if bp != "" {
			BasePrompt = bp
		}
		return tasks, nil
	}

	if txt := os.Getenv("TASKS_JSON"); txt != "" {
		tasks := []Task{}
		if err := json.Unmarshal([]byte(txt), &tasks); err != nil {
			return nil, err
		}
		return tasks, nil
	}

	for _, fn := range []string{"tasks.yml", "tasks.yaml"} {
		if _, err := os.Stat(fn); err == nil {
			tasks, bp, err := readTasksFile(fn)
			if err != nil {
				return nil, err
			}
			if bp != "" {
				BasePrompt = bp
			}
			return tasks, nil
		}
	}

	log.Print("tasks.yml not found; using default tasks")

	lunchTime := envDefault("LUNCH_TIME", DefaultLunchTime)
	briefTime := envDefault("BRIEF_TIME", DefaultBriefTime)
	return []Task{
		{Name: "lunch", Prompt: LunchIdeaPrompt, Time: lunchTime},
		{Name: "brief", Prompt: DailyBriefPrompt, Time: briefTime},
	}, nil
}

// FormatTasks returns a text summary of tasks with their time or cron expression.
func FormatTasks(tasks []Task) string {
	if len(tasks) == 0 {
		return "no tasks"
	}
	var b strings.Builder
	for i, t := range tasks {
		when := t.Cron
		if when == "" {
			when = t.Time
			if when == "" {
				when = "00:00"
			}
		}
		name := t.Name
		if name == "" {
			name = fmt.Sprintf("task %d", i+1)
		}
		fmt.Fprintf(&b, "%s - %s\n", when, name)
	}
	return strings.TrimSpace(b.String())
}

// FormatTaskNames returns a newline separated list of task names.
func FormatTaskNames(tasks []Task) string {
	names := []string{}
	for _, t := range tasks {
		if t.Name != "" {
			names = append(names, t.Name)
		}
	}
	if len(names) == 0 {
		return "no tasks"
	}
	return strings.Join(names, "\n")
}

// FindTask returns the task with the given name, if any.
func FindTask(tasks []Task, name string) (Task, bool) {
	for _, t := range tasks {
		if t.Name == name {
			return t, true
		}
	}
	return Task{}, false
}
