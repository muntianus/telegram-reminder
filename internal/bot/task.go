package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"telegram-reminder/internal/logger"

	yaml "gopkg.in/yaml.v3"
)

// Task struct is defined in bot.go

// envDefault returns the environment variable value or a default if not set.
// This is a utility function for providing fallback values for configuration.
//
// Parameters:
//   - key: Environment variable name
//   - def: Default value to return if environment variable is not set
//
// Returns:
//   - string: Environment variable value or default value
func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// readTasksFile loads tasks from a YAML or JSON file.
// Supports both individual task arrays and structured files with base_prompt.
//
// Parameters:
//   - fn: File path to read tasks from
//
// Returns:
//   - []Task: Array of loaded tasks
//   - string: Base prompt if found in file
//   - error: Any error that occurred during file reading or parsing
func readTasksFile(fn string) ([]Task, string, string, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, "", "", err
	}
	logger.L.Debug("read tasks file", "file", fn)
	tasks := []Task{}
	ext := strings.ToLower(filepath.Ext(fn))
	var tf struct {
		BasePrompt string `json:"base_prompt" yaml:"base_prompt"`
		Model      string `json:"model" yaml:"model"`
		Tasks      []Task `json:"tasks" yaml:"tasks"`
	}
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &tf); err == nil && len(tf.Tasks) > 0 {
			for i := range tf.Tasks {
				if tf.Tasks[i].Model == "" {
					tf.Tasks[i].Model = tf.Model
				}
			}
			return tf.Tasks, tf.BasePrompt, tf.Model, nil
		}
		if err := yaml.Unmarshal(data, &tasks); err != nil {
			return nil, "", "", err
		}
	} else {
		if err := json.Unmarshal(data, &tf); err == nil && len(tf.Tasks) > 0 {
			for i := range tf.Tasks {
				if tf.Tasks[i].Model == "" {
					tf.Tasks[i].Model = tf.Model
				}
			}
			return tf.Tasks, tf.BasePrompt, tf.Model, nil
		}
		if err := json.Unmarshal(data, &tasks); err != nil {
			return nil, "", "", err
		}
	}
	return tasks, "", "", nil
}

// LoadTasks reads task configuration from multiple sources in order of priority:
// 1. TASKS_FILE environment variable (YAML/JSON file)
// 2. TASKS_JSON environment variable (JSON string)
// 3. tasks.yml or tasks.yaml files in current directory
// 4. Default tasks using LUNCH_TIME and BRIEF_TIME environment variables
//
// Returns:
//   - []Task: Array of loaded tasks
//   - error: Any error that occurred during loading
func LoadTasks() ([]Task, error) {
	if fn := os.Getenv("TASKS_FILE"); fn != "" {
		logger.L.Debug("load tasks from file", "file", fn)
		tasks, bp, m, err := readTasksFile(fn)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		if bp != "" {
			BasePrompt = bp
		}
		if m != "" {
			ModelMu.Lock()
			CurrentModel = m
			ModelMu.Unlock()
		}
		return tasks, nil
	}

	if txt := os.Getenv("TASKS_JSON"); txt != "" {
		logger.L.Debug("load tasks from json env")
		tasks := []Task{}
		if err := json.Unmarshal([]byte(txt), &tasks); err != nil {
			return nil, err
		}
		return tasks, nil
	}

	for _, fn := range []string{"tasks.yml", "tasks.yaml"} {
		if _, err := os.Stat(fn); err == nil {
			logger.L.Debug("load tasks from local", "file", fn)
			tasks, bp, m, err := readTasksFile(fn)
			if err != nil {
				return nil, err
			}
			if bp != "" {
				BasePrompt = bp
			}
			if m != "" {
				ModelMu.Lock()
				CurrentModel = m
				ModelMu.Unlock()
			}
			return tasks, nil
		}
	}

	logger.L.Info("tasks.yml not found; using default tasks")
	logger.L.Debug("using default times", "lunch", envDefault("LUNCH_TIME", DefaultLunchTime), "brief", envDefault("BRIEF_TIME", DefaultBriefTime))

	lunchTime := envDefault("LUNCH_TIME", DefaultLunchTime)
	briefTime := envDefault("BRIEF_TIME", DefaultBriefTime)
	return []Task{
		{Name: "lunch", Prompt: LunchIdeaPrompt, Time: lunchTime},
		{Name: "brief", Prompt: DailyBriefPrompt, Time: briefTime},
	}, nil
}

// FormatTasks returns a text summary of tasks with their time or cron expression.
// Each task is formatted as "time - name" on a separate line.
//
// Parameters:
//   - tasks: Array of tasks to format
//
// Returns:
//   - string: Formatted task list or "no tasks" if empty
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
// Only tasks with non-empty names are included.
//
// Parameters:
//   - tasks: Array of tasks to extract names from
//
// Returns:
//   - string: Newline-separated task names or "no tasks" if empty
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
// Performs case-sensitive name matching.
//
// Parameters:
//   - tasks: Array of tasks to search in
//   - name: Task name to find
//
// Returns:
//   - Task: Found task or empty Task struct
//   - bool: True if task was found, false otherwise
func FindTask(tasks []Task, name string) (Task, bool) {
	for _, t := range tasks {
		if t.Name == name {
			return t, true
		}
	}
	return Task{}, false
}
