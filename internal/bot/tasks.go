package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"telegram-reminder/internal/logger"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Task представляет задачу для планировщика или команды.
type Task struct {
	Name   string `json:"name" yaml:"name"`
	Prompt string `json:"prompt" yaml:"prompt"`
	Time   string `json:"time,omitempty" yaml:"time,omitempty"`
	Cron   string `json:"cron,omitempty" yaml:"cron,omitempty"`
}

// BasePrompt — общий шаблон для задач (может быть переопределён из файла).
// LoadedTasks — глобальный список задач (используется только в тестах для подмены задач).
// TasksMu — мьютекс для потокобезопасной работы с LoadedTasks в тестах.
var (
	BasePrompt  string
	LoadedTasks []Task
	TasksMu     sync.RWMutex
)

// envDefault возвращает значение переменной окружения или дефолт.
func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// readTasksFile загружает задачи из YAML или JSON файла.
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

// LoadTasks читает задачи из конфигурации (env, файл, дефолты).
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

	logger.L.Info("tasks.yml not found; using default tasks")

	const DefaultLunchTime = "13:00"
	const DefaultBriefTime = "20:00"
	lunchTime := envDefault("LUNCH_TIME", DefaultLunchTime)
	briefTime := envDefault("BRIEF_TIME", DefaultBriefTime)
	return []Task{
		{Name: "lunch", Prompt: LunchIdeaPrompt, Time: lunchTime},
		{Name: "brief", Prompt: DailyBriefPrompt, Time: briefTime},
	}, nil
}

// applyTemplate подставляет переменные окружения и шаблоны в prompt задачи.
func applyTemplate(prompt string) string {
	vars := map[string]string{
		"base_prompt":  BasePrompt,
		"date":         time.Now().Format("2006-01-02"),
		"exchange_api": os.Getenv("EXCHANGE_API"),
		"chart_path":   os.Getenv("CHART_PATH"),
	}
	for k, v := range vars {
		prompt = strings.ReplaceAll(prompt, "{"+k+"}", v)
	}
	return prompt
}

// FormatTasks возвращает текстовое описание задач с их временем или cron.
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

// FormatTaskNames возвращает список имён задач через перевод строки.
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

// FindTask ищет задачу по имени в списке задач.
func FindTask(tasks []Task, name string) (Task, bool) {
	for _, t := range tasks {
		if t.Name == name {
			return t, true
		}
	}
	return Task{}, false
}
