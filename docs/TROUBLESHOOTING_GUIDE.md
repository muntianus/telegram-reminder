# 🔧 Руководство по устранению неполадок

## Содержание

- [Проблемы с запуском](#проблемы-с-запуском)
- [Ошибки OpenAI API](#ошибки-openai-api)
- [Проблемы с веб-поиском](#проблемы-с-веб-поиском)
- [Проблемы с Telegram API](#проблемы-с-telegram-api)
- [Проблемы конфигурации](#проблемы-конфигурации)
- [Проблемы производительности](#проблемы-производительности)
- [Диагностические инструменты](#диагностические-инструменты)

## Проблемы с запуском

### ❌ "missing required env vars"

**Симптомы:**
```
config load err="missing required env vars"
```

**Причина:** Не заданы обязательные переменные окружения.

**Решение:**
```bash
# Проверьте наличие обязательных переменных
echo "TELEGRAM_TOKEN: ${TELEGRAM_TOKEN:0:10}..."
echo "OPENAI_API_KEY: ${OPENAI_API_KEY:0:10}..."

# Если переменные не заданы:
export TELEGRAM_TOKEN="ваш_telegram_token"
export OPENAI_API_KEY="ваш_openai_key"
```

### ❌ "Invalid Chat ID"

**Симптомы:**
```
config load err="invalid CHAT_ID: user ID too large"
```

**Причина:** Chat ID выходит за допустимые пределы Telegram.

**Решение:**
```bash
# Проверьте диапазоны Chat ID:
# Пользователи: 1 до 2,147,483,647
# Группы/каналы: -2,147,483,648 до -1

# Получите правильный Chat ID:
# 1. Добавьте бота в чат
# 2. Отправьте команду /start
# 3. Проверьте логи бота для получения правильного ID
```

### ❌ Контейнер сразу завершается

**Симптомы:**
```bash
docker ps  # Контейнер не запущен
docker logs telegram-bot  # Показывает ошибки
```

**Диагностика:**
```bash
# Проверьте логи
docker logs telegram-bot

# Проверьте переменные окружения
docker exec telegram-bot env | grep -E "(TELEGRAM|OPENAI)"

# Проверьте права доступа к файлам
docker exec telegram-bot ls -la /app/
```

## Ошибки OpenAI API

### ❌ Пустые ответы от OpenAI

**Симптомы:**
- Команды возвращают "Извините, получен пустой ответ от OpenAI"
- В логах: `openai returned empty response`

**Причины и решения:**

1. **Превышение лимитов токенов:**
```bash
# Увеличьте лимит токенов
export OPENAI_MAX_TOKENS=1000

# Или проверьте использование в OpenAI Dashboard
curl -s "https://api.openai.com/v1/usage" \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

2. **Проблемы с моделью:**
```bash
# Переключитесь на другую модель
export OPENAI_MODEL="gpt-4o-mini"

# Проверьте доступность модели
curl -s "https://api.openai.com/v1/models" \
  -H "Authorization: Bearer $OPENAI_API_KEY" | jq '.data[].id'
```

3. **Проблемы с инструментами:**
```bash
# Отключите веб-поиск временно
export ENABLE_WEB_SEARCH=false
export OPENAI_TOOL_CHOICE=none
```

### ❌ "Invalid API Key"

**Симптомы:**
```
❌ Неверный API ключ OpenAI
💡 Проверьте OPENAI_API_KEY в настройках
```

**Решение:**
```bash
# Проверьте формат ключа
echo $OPENAI_API_KEY | grep -E "^sk-[a-zA-Z0-9-]{40,}"

# Проверьте ключ через API
curl -s "https://api.openai.com/v1/models" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -w "%{http_code}\n"

# Код 200 = ключ валиден
# Код 401 = невалидный ключ
# Код 429 = превышен лимит
```

### ❌ Превышение лимитов (Rate Limit)

**Симптомы:**
```
openai error err="rate limit exceeded"
```

**Решение:**
```bash
# Проверьте лимиты в OpenAI Dashboard
# Или добавьте retry логику

# Временное решение - увеличьте паузу между запросами
# В коде это автоматически обрабатывается
```

## Проблемы с веб-поиском

### ❌ "Веб-поиск временно недоступен"

**Симптомы:**
- Команды с поиском возвращают ошибку
- В логах: `web search failed`

**Диагностика:**
```bash
# Проверьте настройки веб-поиска
echo "ENABLE_WEB_SEARCH: $ENABLE_WEB_SEARCH"

# Проверьте логи поиска
docker logs telegram-bot 2>&1 | grep -i "web search"
```

**Решение:**
```bash
# 1. Временно отключите веб-поиск
export ENABLE_WEB_SEARCH=false

# 2. Проверьте модель (не все поддерживают веб-поиск)
export OPENAI_MODEL="gpt-4.1"  # Поддерживает веб-поиск

# 3. Проверьте tool_choice
export OPENAI_TOOL_CHOICE=auto
```

### ❌ "По вашему запросу ничего не найдено"

**Симптомы:**
- Поиск возвращает пустые результаты
- В логах: `web search returned empty results`

**Решение:**
```bash
# Переформулируйте запрос более конкретно
/search "OpenAI API documentation 2025"

# Проверьте кеш поиска
# Кеш автоматически очищается каждые 10 минут
```

## Проблемы с Telegram API

### ❌ "telegram: Not Found (404)"

**Симптомы:**
```
telegram send err="telegram: Not Found (404)"
```

**Причины и решения:**

1. **Неверный токен бота:**
```bash
# Проверьте токен
curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/getMe"
```

2. **Неверный Chat ID:**
```bash
# Получите правильный Chat ID:
# 1. Добавьте @userinfobot в чат
# 2. Отправьте команду /start
# 3. Скопируйте ID из ответа
```

3. **Бот заблокирован пользователем:**
```bash
# Пользователь должен разблокировать бота
# Или использовать другой Chat ID
```

### ❌ "telegram: Forbidden (403)"

**Симптомы:**
```
telegram send err="telegram: Forbidden (403)"
```

**Причины:**
- Бот не добавлен в группу/канал
- У бота нет прав отправлять сообщения
- Бот заблокирован администратором

**Решение:**
```bash
# 1. Добавьте бота в чат как администратора
# 2. Дайте права "Send Messages"
# 3. Отправьте /start в чате
```

## Проблемы конфигурации

### ❌ "Input too long or invalid"

**Симптомы:**
- Команды отклоняются с сообщением о длине
- В логах: `invalid payload`

**Причина:** Ввод превышает установленные лимиты.

**Лимиты безопасности:**
```
Команды (/remove, /model):     макс. 2000 символов
Поисковые запросы (/search):   2-1000 символов  
Сообщения чата (/chat):        макс. 4000 символов
```

**Решение:**
```bash
# Сократите длину ввода
/chat Короткий вопрос вместо очень длинного

# Разбейте длинный текст на части
/chat Часть 1 вопроса
/chat Часть 2 вопроса
```

### ❌ "tasks.yml not found"

**Симптомы:**
```
tasks.yml not found; using default tasks
```

**Решение:**
```bash
# Создайте файл задач
cat > tasks.yml << 'EOF'
model: gpt-4.1
tasks:
  - name: test_task
    time: "09:00"
    prompt: "Тестовая задача"
EOF

# Укажите путь к файлу
export TASKS_FILE=./tasks.yml
```

## Проблемы производительности

### ❌ Медленные ответы

**Симптомы:**
- Команды выполняются дольше 30 секунд
- Таймауты в логах

**Диагностика:**
```bash
# Проверьте время ответа OpenAI
time curl -s "https://api.openai.com/v1/chat/completions" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hi"}],
    "max_tokens": 100
  }'
```

**Решение:**
```bash
# 1. Используйте более быстрые модели
export OPENAI_MODEL="gpt-4o-mini"

# 2. Уменьшите лимит токенов
export OPENAI_MAX_TOKENS=300

# 3. Отключите веб-поиск для простых запросов
export ENABLE_WEB_SEARCH=false
```

### ❌ Высокое использование памяти

**Симптомы:**
```bash
docker stats telegram-bot  # Высокое MEM USAGE
```

**Диагностика:**
```bash
# Проверьте размер кеша
docker exec telegram-bot ps aux

# Проверьте goroutine
docker exec telegram-bot pgrep -f bot | xargs -I {} cat /proc/{}/status | grep -E "(VmRSS|Threads)"
```

**Решение:**
```bash
# Кеш автоматически ограничен 100 записями
# Если проблема продолжается, перезапустите контейнер
docker restart telegram-bot
```

## Диагностические инструменты

### 🔍 Включение отладочных логов

```bash
# Максимальная детализация
export LOG_LEVEL=debug

# Отправка логов в отдельный чат
export LOG_CHAT_ID=-1001234567890
```

### 📊 Мониторинг в реальном времени

```bash
# Логи в реальном времени
docker logs -f telegram-bot

# Фильтрация ошибок
docker logs telegram-bot 2>&1 | grep -i error

# Мониторинг ресурсов
watch 'docker stats --no-stream telegram-bot'
```

### 🧪 Тестирование компонентов

#### Тест Telegram API
```bash
curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/getMe" | jq .
```

#### Тест OpenAI API
```bash
curl -s "https://api.openai.com/v1/models" \
  -H "Authorization: Bearer $OPENAI_API_KEY" | jq '.data[0]'
```

#### Тест конфигурации
```bash
# Проверка всех переменных
env | grep -E "(TELEGRAM|OPENAI|LOG)" | sort
```

### 📋 Чек-лист диагностики

При возникновении проблем выполните:

- [ ] Проверьте логи: `docker logs telegram-bot`
- [ ] Проверьте переменные окружения
- [ ] Проверьте доступность внешних API
- [ ] Проверьте права доступа к файлам
- [ ] Проверьте сетевое подключение
- [ ] Перезапустите сервис: `docker restart telegram-bot`

### 🆘 Сбор информации для поддержки

При обращении за помощью предоставьте:

```bash
# Информация о версии
docker exec telegram-bot ./bot --version

# Конфигурация (без секретов)
env | grep -E "(TELEGRAM_TOKEN|OPENAI_API_KEY)" | sed 's/=.*/=***/'
env | grep -v -E "(TELEGRAM_TOKEN|OPENAI_API_KEY)" | grep -E "(TELEGRAM|OPENAI|LOG)"

# Последние логи (без личной информации)
docker logs --tail 50 telegram-bot 2>&1 | sed 's/chat":[0-9]*/chat":***/'

# Системная информация
docker info | grep -E "(Server Version|OS/Arch)"
```

---

**💡 Совет:** Большинство проблем решаются перезапуском бота и проверкой конфигурации. Всегда сначала проверяйте логи.

**📞 Поддержка:** Если проблема не решается, создайте issue в GitHub с подробным описанием и диагностической информацией.