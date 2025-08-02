# ⚙️ Подробное руководство по конфигурации

## Содержание

- [Переменные окружения](#переменные-окружения)
- [Файловая конфигурация](#файловая-конфигурация)
- [Примеры конфигурации](#примеры-конфигурации)
- [Валидация настроек](#валидация-настроек)
- [Производственная конфигурация](#производственная-конфигурация)

## Переменные окружения

### 🔑 Обязательные переменные

| Переменная | Описание | Пример | Валидация |
|------------|----------|---------|-----------|
| `TELEGRAM_TOKEN` | Токен Telegram бота | `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11` | Формат: `число:строка` |
| `OPENAI_API_KEY` | Ключ OpenAI API | `sk-1234567890abcdef...` | Начинается с `sk-` |

### 📋 Основные настройки

| Переменная | Описание | По умолчанию | Валидация |
|------------|----------|--------------|-----------|
| `CHAT_ID` | ID целевого чата | не задан | Диапазон Telegram ID |
| `LOG_CHAT_ID` | ID чата для логов | не задан | Диапазон Telegram ID |
| `OPENAI_MODEL` | Модель OpenAI | `gpt-4.1` | Список поддерживаемых |
| `OPENAI_MAX_TOKENS` | Максимум токенов | `600` | 1-32768 |
| `LOG_LEVEL` | Уровень логирования | `info` | debug/info/warn/error |

### 🔧 Расширенные настройки

| Переменная | Описание | По умолчанию | Примечания |
|------------|----------|--------------|------------|
| `OPENAI_TOOL_CHOICE` | Режим tool calling | `auto` | auto/none/required |
| `OPENAI_SERVICE_TIER` | Уровень сервиса | не задан | default/flex |
| `OPENAI_REASONING_EFFORT` | Усилие рассуждения | не задан | low/medium/high |
| `ENABLE_WEB_SEARCH` | Включить веб-поиск | `true` | true/false |
| `BLOCKCHAIN_API` | URL API блокчейна | `https://api.blockchain.info/stats` | Валидный URL |

### ⏰ Временные настройки

| Переменная | Описание | По умолчанию | Формат |
|------------|----------|--------------|--------|
| `LUNCH_TIME` | Время идей для обеда | `13:00` | HH:MM |
| `BRIEF_TIME` | Время вечернего дайджеста | `20:00` | HH:MM |

### 📁 Файловые настройки

| Переменная | Описание | По умолчанию | Примечания |
|------------|----------|--------------|------------|
| `TASKS_FILE` | Путь к файлу задач | не задан | YAML файл |
| `TASKS_JSON` | JSON строка с задачами | не задан | Приоритет над TASKS_FILE |
| `WHITELIST_FILE` | Путь к whitelist | `whitelist.json` | JSON файл |

## Файловая конфигурация

### 📋 Конфигурация задач (tasks.yml)

#### Базовая структура
```yaml
# Глобальные настройки
model: gpt-4.1
base_prompt: &base |
  Твой базовый промпт здесь.
  Может содержать переменные: {date}, {time}

# Список задач
tasks:
  - name: task_name        # Имя команды (/task_name)
    time: "09:00"         # Время выполнения (HH:MM)
    model: gpt-4.1        # Модель для этой задачи (опционально)
    prompt: |             # Промпт для задачи
      Твой промпт здесь
  
  - name: another_task
    time: "14:00"
    prompt: *base         # Использование базового промпта
```

#### Переменные в промптах
```yaml
tasks:
  - name: daily_report
    time: "09:00"  
    prompt: |
      Создай отчет за {date}.
      Текущее время: {time}
      API курса валют: {exchange_api}
      Путь к графику: {chart_path}
```

Доступные переменные:
- `{date}` - текущая дата
- `{time}` - текущее время  
- `{exchange_api}` - из `EXCHANGE_API`
- `{chart_path}` - из `CHART_PATH`

### 📝 Альтернативная конфигурация (JSON)

#### Через переменную окружения
```bash
export TASKS_JSON='[
  {
    "name": "morning_brief",
    "time": "09:00",
    "model": "gpt-4.1",
    "prompt": "Создай утренний дайджест"
  }
]'
```

#### Через файл
```json
[
  {
    "name": "morning_brief", 
    "time": "09:00",
    "model": "gpt-4.1",
    "prompt": "Создай утренний дайджест"
  },
  {
    "name": "evening_brief",
    "time": "20:00", 
    "prompt": "Создай вечерний дайджест"
  }
]
```

### 📋 Whitelist (whitelist.json)

```json
[
  123456789,
  -1001234567890,
  987654321
]
```

## Примеры конфигурации

### 🏠 Локальная разработка

#### .env файл
```env
# Обязательные
TELEGRAM_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
OPENAI_API_KEY=sk-1234567890abcdef...

# Для тестирования
CHAT_ID=123456789
LOG_LEVEL=debug
LOG_CHAT_ID=123456789

# OpenAI настройки
OPENAI_MODEL=gpt-4o-mini
OPENAI_MAX_TOKENS=300
ENABLE_WEB_SEARCH=false

# Временные настройки
LUNCH_TIME=12:00
BRIEF_TIME=18:00

# Файлы
TASKS_FILE=./tasks.yml
WHITELIST_FILE=./whitelist.json
```

#### Запуск
```bash
source .env
go run ./cmd/bot
```

### 🐳 Docker разработка

#### docker-compose.yml
```yaml
version: '3.8'

services:
  telegram-bot:
    build: .
    environment:
      - TELEGRAM_TOKEN=${TELEGRAM_TOKEN}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - CHAT_ID=${CHAT_ID}
      - LOG_LEVEL=info
      - OPENAI_MODEL=gpt-4.1
      - OPENAI_MAX_TOKENS=600
    volumes:
      - ./tasks.yml:/app/tasks.yml:ro
      - ./data:/app/data
    restart: unless-stopped
```

### 🏭 Производственная среда

#### Kubernetes ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: telegram-bot-config
data:
  OPENAI_MODEL: "gpt-4.1"
  OPENAI_MAX_TOKENS: "600"
  LOG_LEVEL: "info"
  ENABLE_WEB_SEARCH: "true"
  BLOCKCHAIN_API: "https://api.blockchain.info/stats"
  LUNCH_TIME: "13:00"
  BRIEF_TIME: "20:00"
```

#### Kubernetes Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: telegram-bot-secrets
type: Opaque
stringData:
  TELEGRAM_TOKEN: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
  OPENAI_API_KEY: "sk-1234567890abcdef..."
  CHAT_ID: "123456789"
  LOG_CHAT_ID: "-1001234567890"
```

#### Docker Swarm
```yaml
version: '3.8'

services:
  telegram-bot:
    image: telegram-bot:latest
    deploy:
      replicas: 1
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    environment:
      - OPENAI_MODEL=gpt-4.1
      - LOG_LEVEL=info
    secrets:
      - telegram_token
      - openai_api_key
    configs:
      - source: tasks_config
        target: /app/tasks.yml

secrets:
  telegram_token:
    external: true
  openai_api_key:
    external: true

configs:
  tasks_config:
    file: ./tasks.yml
```

## Валидация настроек

### 🔍 Автоматическая валидация

При запуске бот автоматически проверяет:

#### Обязательные переменные
```go
if telegramToken == "" || openaiKey == "" {
    return cfg, fmt.Errorf("missing required env vars")
}
```

#### Chat ID диапазоны
```go
// Пользователи: 1 до 2,147,483,647
// Группы/каналы: -2,147,483,648 до -1
if chatID > 0 && chatID > 2147483647 {
    return fmt.Errorf("user ID too large")
}
```

#### Численные значения
```go
maxTokens := 600
if maxTokensStr != "" {
    if v, err := strconv.Atoi(maxTokensStr); err == nil {
        maxTokens = v
    }
}
```

### ✅ Ручная проверка конфигурации

#### Скрипт валидации
```bash
#!/bin/bash
# validate_config.sh

echo "🔍 Проверка конфигурации..."

# Проверка обязательных переменных
if [[ -z "$TELEGRAM_TOKEN" ]]; then
    echo "❌ TELEGRAM_TOKEN не задан"
    exit 1
fi

if [[ -z "$OPENAI_API_KEY" ]]; then
    echo "❌ OPENAI_API_KEY не задан"
    exit 1
fi

# Проверка формата TELEGRAM_TOKEN
if [[ ! "$TELEGRAM_TOKEN" =~ ^[0-9]+:.+ ]]; then
    echo "❌ Неверный формат TELEGRAM_TOKEN"
    exit 1
fi

# Проверка формата OPENAI_API_KEY
if [[ ! "$OPENAI_API_KEY" =~ ^sk-.+ ]]; then
    echo "❌ Неверный формат OPENAI_API_KEY"
    exit 1
fi

# Проверка CHAT_ID
if [[ -n "$CHAT_ID" ]]; then
    if [[ ! "$CHAT_ID" =~ ^-?[0-9]+$ ]]; then
        echo "❌ CHAT_ID должен быть числом"
        exit 1
    fi
    
    if [[ "$CHAT_ID" -eq 0 ]]; then
        echo "❌ CHAT_ID не может быть 0"
        exit 1
    fi
fi

# Проверка файлов
if [[ -n "$TASKS_FILE" && ! -f "$TASKS_FILE" ]]; then
    echo "⚠️  Файл задач $TASKS_FILE не найден"
fi

if [[ -n "$WHITELIST_FILE" && ! -f "$WHITELIST_FILE" ]]; then
    echo "⚠️  Файл whitelist $WHITELIST_FILE не найден"
fi

echo "✅ Конфигурация валидна"
```

### 🧪 Тестирование конфигурации

#### Dry-run режим
```bash
# Проверка конфигурации без запуска
go run ./cmd/bot --dry-run
```

#### Проверка подключений
```bash
# Тест Telegram API
curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/getMe"

# Тест OpenAI API
curl -s "https://api.openai.com/v1/models" \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

## Производственная конфигурация

### 🔐 Безопасность

#### Ротация секретов
```bash
# Скрипт ротации API ключей
#!/bin/bash
NEW_OPENAI_KEY="sk-новый-ключ"
OLD_OPENAI_KEY="$OPENAI_API_KEY"

# Обновление в секретах
kubectl patch secret telegram-bot-secrets \
  --patch '{"stringData":{"OPENAI_API_KEY":"'$NEW_OPENAI_KEY'"}}'

# Перезапуск подов
kubectl rollout restart deployment telegram-bot

# Деактивация старого ключа (после проверки)
# curl -X DELETE "https://api.openai.com/v1/keys/$OLD_OPENAI_KEY"
```

#### Мониторинг конфигурации
```yaml
# Prometheus мониторинг
- name: config_validation
  rules:
  - alert: InvalidChatID
    expr: telegram_bot_config_errors{type="chat_id"} > 0
    for: 1m
    annotations:
      summary: "Неверный Chat ID в конфигурации"
      
  - alert: MissingRequiredEnv
    expr: telegram_bot_config_errors{type="required_env"} > 0
    for: 0s
    annotations:
      summary: "Отсутствуют обязательные переменные окружения"
```

### 📊 Мониторинг настроек

#### Логирование конфигурации
```go
logger.L.Info("Bot configuration loaded",
    "model", cfg.OpenAIModel,
    "max_tokens", cfg.OpenAIMaxTokens,
    "web_search_enabled", cfg.EnableWebSearch,
    "chat_id_set", cfg.ChatID != 0,
)
```

#### Healthcheck endpoint
```go
func healthCheck(cfg Config) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        status := map[string]interface{}{
            "status": "ok",
            "config": map[string]interface{}{
                "model": cfg.OpenAIModel,
                "web_search": cfg.EnableWebSearch,
                "chat_configured": cfg.ChatID != 0,
            },
        }
        json.NewEncoder(w).Encode(status)
    }
}
```

---

**💡 Совет:** Всегда тестируйте изменения конфигурации в изолированной среде перед применением в продакшене.

**📞 Поддержка:** При проблемах с конфигурацией проверьте логи бота с уровнем `debug` для подробной диагностики.