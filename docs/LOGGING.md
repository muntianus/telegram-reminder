# Система логирования

Проект использует структурированную систему логирования на основе `log/slog` с человекочитаемым форматированием и модульной архитектурой.

## Основные возможности

### 1. Структурированное логирование
- Контекстно-зависимые логгеры для разных модулей
- Операционное логирование с отслеживанием жизненного цикла
- Метрики производительности и API вызовов
- Безопасное логирование с защитой от утечки данных

### 2. Человекочитаемые форматы
- **Pretty Format**: Цветной вывод с улучшенной читаемостью
- **JSON Format**: Структурированный JSON для парсинга
- **Text Format**: Компактный текстовый формат

### 3. Модульные логгеры
Каждый модуль имеет собственный логгер с настраиваемым уровнем:
- `bot` - основная логика бота
- `api` - API вызовы
- `task` - выполнение задач
- `telegram` - операции с Telegram
- `security` - события безопасности
- `digest` - операции с дайджестами
- `handler` - обработчики запросов
- `openai` - операции с OpenAI

## Конфигурация

### Переменные окружения

```bash
# Основные настройки
LOG_LEVEL=info                    # debug, info, warn, error
LOG_FORMAT=pretty                 # pretty, json, text
LOG_COLORS=true                   # включить цвета в pretty формате

# Уровни для отдельных модулей
LOG_MODULE_LEVELS="bot=debug,api=info,security=warn"

# Telegram логирование (для критических событий)
LOG_TELEGRAM_ENABLE=true
LOG_TELEGRAM_CHAT=123456789
```

### Примеры форматов

#### Pretty Format (рекомендуется для разработки)
```
15:30:45 [INFO ] [bot] User action {user_id=123456789, action=chat, query_length=25}
15:30:45 [DEBUG] [handler] Operation step {operation=chat_completion, step=calling_openai_api, user_id=123456789}
15:30:46 [INFO ] [openai] API call successful {service=openai, endpoint=chat_completion, duration=1.2s}
15:30:46 [INFO ] [handler] Chat completion successful {operation=chat_completion, status=success, duration=1.3s, response_length=156}
```

#### JSON Format (для продакшена)
```json
{"timestamp":"2025-08-03T15:30:45Z","level":"INFO","module":"bot","msg":"User action","user_id":123456789,"action":"chat","query_length":25}
{"timestamp":"2025-08-03T15:30:46Z","level":"INFO","module":"openai","msg":"API call successful","service":"openai","endpoint":"chat_completion","duration":"1.2s"}
```

## Использование в коде

### Базовое логирование

```go
import "telegram-reminder/internal/logger"

// Получение модульного логгера
handlerLogger := logger.GetHandlerLogger()
openaiLogger := logger.GetOpenAILogger()

// Простое логирование
handlerLogger.Info("Processing user request", "user_id", userID)
handlerLogger.Error("Request failed", "error", err)
```

### Операционное логирование

```go
// Создание операции
op := handlerLogger.Operation("user_registration")
op.WithContext("user_id", userID)
op.WithContext("email", email)

// Логирование шагов
op.Step("validating_input", "email_valid", true)
op.Step("checking_database")

// Завершение операции
if err != nil {
    op.Failure("Registration failed", err)
} else {
    op.Success("User registered successfully", "account_id", accountID)
}
```

### Специализированные методы

```go
// API вызовы
openaiLogger.APICall("openai", "chat_completion", true, duration, nil)

// Пользовательские действия
handlerLogger.UserAction(userID, "send_message", map[string]interface{}{
    "message_length": len(message),
    "chat_type": "private",
})

// События безопасности
securityLogger := logger.GetSecurityLogger()
securityLogger.SecurityEvent("failed_auth", userID, map[string]interface{}{
    "attempts": 3,
    "ip": "192.168.1.1",
})

// Выполнение задач
taskLogger := logger.GetTaskLogger()
taskLogger.TaskExecution("daily_reminder", true, duration, nil)

// HTTP запросы
handlerLogger.HTTPRequest("POST", "/api/chat", 200, duration)

// Метрики производительности
handlerLogger.Performance("database_query", duration, map[string]interface{}{
    "query_type": "SELECT",
    "rows_affected": 10,
})
```

### Контекстные логгеры

```go
// Создание логгера с постоянным контекстом
userLogger := handlerLogger.With("user_id", userID, "session_id", sessionID)
userLogger.Info("Starting user session")
userLogger.Debug("Loading user preferences")
```

## Рекомендации по использованию

### 1. Выбор уровней логирования

- **DEBUG**: Детальная отладочная информация (шаги алгоритмов, значения переменных)
- **INFO**: Важные события (пользовательские действия, успешные операции)
- **WARN**: Предупреждения (неоптимальные условия, fallback логика)
- **ERROR**: Ошибки (неудачные операции, исключения)

### 2. Структурирование данных

```go
// ✅ Хорошо - структурированные данные
logger.Info("User login successful", 
    "user_id", userID,
    "login_method", "oauth",
    "duration", duration,
)

// ❌ Плохо - конкатенация строк
logger.Info(fmt.Sprintf("User %d logged in via %s in %s", userID, method, duration))
```

### 3. Безопасность

```go
// ✅ Хорошо - безопасные данные
logger.Info("Password change", "user_id", userID, "success", true)

// ❌ Плохо - утечка чувствительных данных
logger.Debug("User credentials", "password", password) // НИКОГДА!
```

### 4. Производительность

```go
// ✅ Хорошо - проверка уровня для дорогих операций
if logger.Enabled(context.Background(), slog.LevelDebug) {
    expensiveData := computeExpensiveDebugInfo()
    logger.Debug("Debug info", "data", expensiveData)
}

// ✅ Хорошо - ленивое вычисление
logger.Debug("Debug info", "data", func() string {
    return computeExpensiveDebugInfo()
})
```

## Мониторинг и алерты

### Критические события
При включенном Telegram логировании следующие события отправляются в Telegram:
- Ошибки аутентификации
- Сбои API
- Критические системные ошибки
- События безопасности

### Метрики
Система автоматически собирает:
- Время выполнения операций
- Частоту ошибок API
- Количество пользовательских действий
- Производительность задач

## Отладка

### Включение детального логирования
```bash
# Для всего приложения
LOG_LEVEL=debug

# Для конкретных модулей
LOG_MODULE_LEVELS="openai=debug,bot=info"

# Красивый формат для разработки
LOG_FORMAT=pretty
LOG_COLORS=true
```

### Анализ логов
```bash
# Поиск ошибок
grep "ERROR" logs/app.log

# Фильтрация по модулю
grep "\[openai\]" logs/app.log

# JSON логи с jq
cat logs/app.log | jq 'select(.level=="ERROR")'
```

## Миграция с старой системы

Старые вызовы `logger.L.Debug()` продолжают работать благодаря слою совместимости:

```go
// Старый способ (все еще работает)
logger.L.Debug("message", "key", value)

// Новый способ (рекомендуется)
moduleLogger := logger.GetBotLogger()
moduleLogger.Debug("message", "key", value)
```

Рекомендуется постепенно мигрировать на новую систему для получения всех преимуществ структурированного логирования.