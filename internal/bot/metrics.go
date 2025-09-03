package bot

import (
	"fmt"
	"runtime"
	"time"
)

// BotMetrics содержит метрики работы бота
type BotMetrics struct {
	StartTime       time.Time
	CommandsTotal   int64
	MessagesTotal   int64
	ErrorsTotal     int64
	ActiveChats     int
	LastError       string
	LastErrorTime   time.Time
	OpenAIRequests  int64
	MemoryUsage     uint64
}

var metrics = &BotMetrics{
	StartTime: time.Now(),
}

// IncrementCommands увеличивает счетчик команд
func IncrementCommands() {
	metrics.CommandsTotal++
}

// IncrementMessages увеличивает счетчик сообщений  
func IncrementMessages() {
	metrics.MessagesTotal++
}

// IncrementErrors увеличивает счетчик ошибок
func IncrementErrors(err string) {
	metrics.ErrorsTotal++
	metrics.LastError = err
	metrics.LastErrorTime = time.Now()
}

// IncrementOpenAI увеличивает счетчик запросов к OpenAI
func IncrementOpenAI() {
	metrics.OpenAIRequests++
}

// UpdateActiveChats обновляет количество активных чатов
func UpdateActiveChats(count int) {
	metrics.ActiveChats = count
}

// GetMetrics возвращает текущие метрики в читаемом формате
func GetMetrics() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	uptime := time.Since(metrics.StartTime)
	
	return fmt.Sprintf(`📊 *Метрики бота*

⏰ Время работы: %v
📩 Всего команд: %d
💬 Всего сообщений: %d
🔥 Запросы OpenAI: %d
👥 Активных чатов: %d
❌ Ошибок: %d
🧠 Память: %.1f MB

%s`, 
		formatDuration(uptime),
		metrics.CommandsTotal,
		metrics.MessagesTotal,
		metrics.OpenAIRequests,
		metrics.ActiveChats,
		metrics.ErrorsTotal,
		float64(m.Alloc)/1024/1024,
		formatLastError(),
	)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

func formatLastError() string {
	if metrics.LastError == "" {
		return "🟢 Ошибок нет"
	}
	
	return fmt.Sprintf("🔴 Последняя: %s (%v назад)", 
		metrics.LastError, 
		formatDuration(time.Since(metrics.LastErrorTime)))
}