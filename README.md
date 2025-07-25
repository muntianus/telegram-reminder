# Телеграм-бот Billion Roadmap

Этот проект — интеллектуальный телеграм-бот на Go для создания персонализированных дайджестов. Бот использует OpenAI для генерации контента и отправляет запланированные сообщения с актуальной информацией за сегодняшний день.

## 🚀 Новые возможности

### Специализированные дайджесты
- **Криптовалюты** (`/crypto`) - анализ рынка, тренды, метрики
- **Технологии** (`/tech`) - новости ИИ, стартапы, инновации
- **Недвижимость** (`/realestate`) - цены, лоты, инвестиции
- **Бизнес** (`/business`) - тренды, стратегии, возможности
- **Инвестиции** (`/investment`) - активы, риски, прогнозы
- **Стартапы** (`/startup`) - фандинг, продукты, тренды
- **Глобальные события** (`/global`) - геополитика, экономика

### Адаптация под Telegram
- 📅 Фокус на информацию за сегодняшний день
- 🔍 Анализ интернета по интересующим темам
- 📱 Оптимизированное форматирование для Telegram
- 🎯 Специализированные промпты для каждого типа контента

### Гибкая настройка
- Полная настройка через `TASKS_FILE` или `TASKS_JSON`
- Выбор оптимальной модели для каждого типа дайджеста
- Автоматическое расписание с московским временем

## 📅 Расписание по умолчанию

* **09:00 MSK** – цены на землю в Подмосковье
* **10:00 MSK** – утренний дайджест
* **11:00 MSK** – криптовалютный обзор
* **12:00 MSK** – технологический дайджест
* **13:00 MSK** – новые лоты ГИС-Торги
* **14:00 MSK** – дневной дайджест
* **15:00 MSK** – MVP идея
* **16:00 MSK** – вечерний крипто-обзор
* **17:00 MSK** – бизнес-дайджест
* **18:00 MSK** – инвестиционный дайджест
* **19:00 MSK** – стартап-дайджест
* **19:30 MSK** – глобальный дайджест
* **20:00 MSK** – бизнес-идея
* **20:30 MSK** – BRI дайджест (Евразия)
* **21:00 MSK** – вечерний дайджест

Бот можно развернуть на любой постоянно работающей площадке: Railway, Fly.io или VPS.

## Команды

### Основные команды
- `/chat <сообщение>` – задать боту вопрос и получить ответ от OpenAI.
- `/search <запрос>` – выполнить поиск через встроенный веб‑поиск OpenAI.
- `/ping` – проверка состояния, в ответ приходит `pong`.
- `/start` – добавить текущий чат в рассылку.
- `/whitelist` – показать список подключённых чатов.
- `/remove <id>` – убрать чат из списка.
- `/model [имя]` – показать или сменить модель генерации (по умолчанию `gpt-4.1`).
- `/lunch` – немедленно запросить идеи на обед.
- `/brief` – немедленно запросить вечерний дайджест.
- `/tasks` – вывести текущее расписание задач.
- `/task [имя]` – показать список задач или выполнить выбранную.

### 🚀 Новые команды дайджестов
- `/crypto` – криптовалютный дайджест за сегодня (рыночные метрики, on-chain анализ, деривативы)
- `/tech` – технологический дайджест за сегодня (AI, мобильные технологии, научные открытия)
- `/realestate` – дайджест недвижимости за сегодня (цены, аукционы, инвестиционные возможности)
- `/business` – бизнес-дайджест за сегодня (рыночные тренды, инновации, экономические индикаторы)
- `/investment` – инвестиционный дайджест за сегодня (акции, крипто, недвижимость, технический анализ)
- `/startup` – стартап-дайджест за сегодня (фандинг, новые продукты, глобальные тренды)
- `/global` – глобальный дайджест за сегодня (геополитика, экономика, региональные новости)

### ✨ Особенности дайджестов

- **Максимум полезной информации**: Конкретные цифры, проценты, имена, даты
- **Поисковые хештеги**: Каждый дайджест заканчивается релевантными хештегами
- **Надежные источники**: 8-10 проверенных источников для каждого типа дайджеста
- **Актуальность**: Только сегодняшние данные и события
- **Структурированность**: Четкое разделение по секциям с эмодзи

### 📊 Примеры хештегов

- Крипто: `#crypto #bitcoin #altcoins #defi #trading #blockchain`
- Инвестиции: `#investing #stocks #crypto #realestate #finance #wealth #portfolio`
- Технологии: `#tech #ai #startups #innovation #software #blockchain #web3`
- Недвижимость: `#realestate #property #investment #moscow #land #mortgage`
- Бизнес: `#business #entrepreneurship #strategy #growth #startups #innovation`
- Стартапы: `#startups #funding #entrepreneurship #innovation #tech #venture`
- Глобальные события: `#world #geopolitics #economy #news #global #politics`

📖 **Подробные примеры дайджестов**: См. [ENHANCED_DIGEST_EXAMPLES.md](docs/ENHANCED_DIGEST_EXAMPLES.md) для детальных примеров всех типов дайджестов с реальными данными и форматированием.

🤖 **Стратегия моделей**: См. [MODELS_STRATEGY.md](docs/MODELS_STRATEGY.md) для подробной информации о распределении моделей OpenAI (o3, o3-mini, gpt-4o-mini) по типам задач.

### Автоматические команды
Все задачи, у которых задано поле `name`, также доступны как команды `/имя`.

Команда `/model` управляет выбором модели OpenAI. Без аргументов она выводит текущую модель и список поддерживаемых идентификаторов из библиотеки `go-openai`. Полный перечень находится в [MODELS.md](docs/MODELS.md). Чтобы переключить модель, передайте её имя, например:

```bash
/model gpt-4-turbo
```

## Требования

* Go 1.24+
* Токен телеграм-бота (`TELEGRAM_TOKEN`)
* ID целевого чата (`CHAT_ID`, опционально)
* Ключ API OpenAI (`OPENAI_API_KEY`)
* Имя модели OpenAI (`OPENAI_MODEL`, опционально, по умолчанию `gpt-4.1`)
* Максимальное число токенов ответа (`OPENAI_MAX_TOKENS`, по умолчанию `600`)
* Выбор использования tools (`OPENAI_TOOL_CHOICE`, по умолчанию `auto`)
* Время идей на обед (`LUNCH_TIME`, опционально, по умолчанию `13:00`)
* Время вечернего дайджеста (`BRIEF_TIME`, опционально, по умолчанию `20:00`)
* Путь к файлу задач (`TASKS_FILE`) или JSON в `TASKS_JSON` для полной настройки расписания
* Путь к файлу whitelist (`WHITELIST_FILE`, опционально, по умолчанию `whitelist.json`)
* URL API блокчейна (`BLOCKCHAIN_API`, опционально, по умолчанию `https://api.blockchain.info/stats`)
* Включить веб-поиск (`ENABLE_WEB_SEARCH`, `true`/`false`, по умолчанию `true`)
* Уровень логирования (`LOG_LEVEL`, опционально, `debug`, `info`, `warn` или `error`) – на `debug` пишутся все события шедулера и запросы к OpenAI
* ID чата для логов (`LOG_CHAT_ID`, опционально)

## Добавление бота в каналы и группы

1. Пригласите бота в нужный канал или группу.
2. Дайте ему права отправлять сообщения (администратор в каналах).
3. Напишите команду `/start` в этом чате. Идентификатор будет сохранён в whitelist, и бот начнёт присылать сообщения.

## Запуск локально

Установите необходимые переменные окружения и запустите бот:

```sh
export TELEGRAM_TOKEN=your_telegram_token
export OPENAI_API_KEY=your_openai_key
export LUNCH_TIME=13:00
export BRIEF_TIME=20:00
export LOG_LEVEL=debug
# подробный уровень логов

# опционально задайте начальный чат
# export CHAT_ID=123456789
# файл whitelist создастся автоматически

# или загрузите задачи из файла
# export TASKS_FILE=tasks.yaml

go run ./cmd/bot
```

После запуска планировщик автоматически отправит два ежедневных сообщения в указанное время. При старте бот публикует текущую версию и перечень доступных команд, подтверждая успешный деплой. Каждый запрос к OpenAI имеет тайм‑аут 3 минуты. Перед созданием pull request запустите `gofmt -w -s` для форматирования кода и `go vet ./...` для проверки. Линтер можно установить командой `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.0`, затем выполните `golangci-lint run`.

## Настройка задач

Бот читает дополнительные задачи из YAML-файла, путь к которому задаётся в переменной `TASKS_FILE`. Каждая задача должна содержать поле `time` в формате `HH:MM` и поле `prompt` с текстом сообщения. Если добавить поле `name`, задачу можно вызвать вручную командой `/имя`. Поле `model` позволяет указать модель OpenAI для конкретной задачи. Также можно задать `model` на верхнем уровне файла — это установит модель по умолчанию для всех задач.

Поддерживаемые переменные окружения и ключи:

- `TELEGRAM_TOKEN` – токен телеграм-бота
- `CHAT_ID` – числовой ID чата назначения (опционально)
- `LOG_CHAT_ID` – ID чата для отправки логов (опционально)
- `OPENAI_API_KEY` – ключ API OpenAI
- `OPENAI_MODEL` – имя модели OpenAI (опционально, по умолчанию `gpt-4.1`)
- `OPENAI_MAX_TOKENS` – максимальное число токенов ответа (по умолчанию `600`)
- `OPENAI_TOOL_CHOICE` – режим tool calling (`auto` или `none`)
- `LUNCH_TIME` – время для идей на обед
- `BRIEF_TIME` – время вечернего дайджеста
- `TASKS_FILE` – путь к YAML-файлу с пользовательскими заданиями
- `WHITELIST_FILE` – путь к файлу со списком чатов (по умолчанию `whitelist.json`)
- `BLOCKCHAIN_API` – URL API блокчейна для команды `/blockchain`
- `ENABLE_WEB_SEARCH` – включить веб-поиск (`true`/`false`, по умолчанию `true`)
- `LOG_LEVEL` – уровень логирования (`debug`, `info`, `warn` или `error`)

Пример `.env` и `tasks.yml`:

```ini
TELEGRAM_TOKEN=123456:ABC-DEF
OPENAI_API_KEY=sk-xxxxxxxx
OPENAI_MODEL=gpt-4.1
OPENAI_MAX_TOKENS=600
OPENAI_TOOL_CHOICE=auto
LUNCH_TIME=12:00
BRIEF_TIME=18:00
TASKS_FILE=tasks.yml
WHITELIST_FILE=whitelist.json
BLOCKCHAIN_API=https://api.blockchain.info/stats
ENABLE_WEB_SEARCH=true
LOG_CHAT_ID=123456789
```

```yaml
- model: gpt-4.1
- base_prompt: &base |
    Ты говоришь кратко, дерзко, панибратски.
    Заполни блоки:
    ⚡ Микродействие (одно простое действие на сегодня)
    🧠 Тема дня (мини‑инсайт/мысль)
    💰 Что залутать (актив/идея)
    🏞️ Земля на присмотр (лоты в южном Подмосковье: Бутово, Щербинка, Подольск, Воскресенск)
    🪙 Альт дня (актуальная монета, линк CoinGecko)
    🚀 Пушка с ProductHunt (ссылка)
    Форматируй одним сообщением, без лишней воды.

tasks:
  - name: land_price
    time: "10:00"
    model: gpt-4.1
    prompt: |
      {base_prompt}
      Задача: Найди цену сотки на {date} и переведи в USD по курсу из {exchange_api}.
  - name: micro_noon
    time: "12:00"
    prompt: *base
  - name: crypto_am
    time: "13:00"
    prompt: |
      Найди метрики крипторынка
...
```

После указания переменных запустите бота командой:

```sh
go run ./cmd/bot
```

## Развёртывание на сервере

1. Установите Go на сервер.
2. Клонируйте этот репозиторий.
3. Задайте переменные окружения, перечисленные выше.
4. Соберите бинарник командой `go build -o bot`.
5. Запустите `./bot` или настройте менеджер процессов, например `systemd`.

## Docker

Соберите образ для текущей платформы:

```sh
docker build --env-file .env -t telegram-bot .
```

Для сборки и публикации мультиархитектурного образа (linux/amd64 и linux/arm64) с помощью `buildx`:

```sh
docker buildx build --platform linux/amd64,linux/arm64 \
  --env-file .env \
  -t your_dockerhub_user/telegram-bot:latest --push .
```

Запуск контейнера:

```sh
docker run -e TELEGRAM_TOKEN=your_token -e OPENAI_API_KEY=your_api_key \
  -e LUNCH_TIME=13:00 -e BRIEF_TIME=20:00 telegram-bot
```

Бота также можно запустить через `docker-compose`. Скопируйте `.env.example` в `.env`, заполните значения и запустите сервис. Переменная `DOCKERHUB_USER` в `.env` определяет аккаунт Docker Hub, используемый в `docker-compose.yml`:

```sh
cp .env.example .env
docker-compose up -d
```

## Развёртывание через GitHub Actions

В репозитории есть workflow GitHub Actions, который автоматически собирает и деплоит Docker-образ при каждом пуше в ветку `main`. Используется Docker Buildx для создания мультиархитектурного образа. Процесс включает следующие шаги:

1. Клонирование репозитория и компиляция Go-бинарника.
2. Сборка и публикация мультиплатформенного образа (linux/amd64 и linux/arm64) в Docker Hub.
3. Подключение к VPS по SSH и перезапуск контейнера через `docker-compose`.

Не забудьте настроить необходимые секреты в настройках репозитория:

* `DOCKERHUB_USER` и `DOCKERHUB_TOKEN` для публикации образа.
* `VPS_SSH_KEY`, `VPS_USER` и `VPS_HOST` для подключения к серверу.
* `TELEGRAM_TOKEN`, `OPENAI_API_KEY` и при необходимости `CHAT_ID` и `OPENAI_MODEL` для заполнения `.env` во время деплоя.

После завершения деплоя зайдите на VPS и выполните `docker ps`, чтобы убедиться, что контейнер запущен.

## Решение проблем

Контейнер завершит работу сразу, если не заданы обязательные переменные окружения:

* `TELEGRAM_TOKEN`
* `OPENAI_API_KEY`

Убедитесь, что эти значения переданы через окружение или указаны в файле `.env`. Корректный пример `.env`:

```ini
TELEGRAM_TOKEN=123456:ABC-DEF
OPENAI_API_KEY=sk-xxxxxxxx
OPENAI_MODEL=gpt-4.1
OPENAI_MAX_TOKENS=600
OPENAI_TOOL_CHOICE=auto
LUNCH_TIME=13:00
BRIEF_TIME=20:00
# CHAT_ID=123456789
# WHITELIST_FILE=whitelist.json
# BLOCKCHAIN_API=https://api.blockchain.info/stats
```

Если контейнер не стартует, проверьте логи командой `docker logs <container>`, чтобы увидеть сообщения об ошибках.

Если бот пишет в логах `tasks.yml not found; using default tasks`, значит файл `tasks.yml` не найден и переменная `TASKS_FILE` не задана. В этом режиме используются стандартные задачи, поэтому команда `/task <имя>` для неопределённых задач вернёт «unknown task». Убедитесь, что `tasks.yml` лежит в рабочей директории или укажите путь к файлу через `TASKS_FILE`.

На Ubuntu VPS логи сервиса, запущенного через `systemd`, можно смотреть командой `journalctl -u telegram-reminder -f`. Если бот работает в Docker, используйте `docker logs -f telegram-reminder`.

## Использование built-in tools

Модели серии `gpt-4.1` поддерживают вызов встроенных инструментов через API `responses`.
Пример использования с веб‑поиском:

```javascript
import OpenAI from "openai";
const client = new OpenAI();

const response = await client.responses.create({
    model: "gpt-4.1",
    tools: [{ type: "web_search_preview" }],
    input: "What was a positive news story from today?",
});

console.log(response.output_text);
```

```python
from openai import OpenAI
client = OpenAI()

response = client.responses.create(
    model="gpt-4.1",
    tools=[{"type": "web_search_preview"}],
    input="What was a positive news story from today?"
)

print(response.output_text)
```

```bash
curl "https://api.openai.com/v1/responses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -d '{
        "model": "gpt-4.1",
        "tools": [{"type": "web_search_preview"}],
        "input": "what was a positive news story from today?"
    }'
```

Инструмент `web_search_preview` позволяет модели получать свежие данные из интернета.

## Лицензия

Проект распространяется на условиях [MIT License](LICENSE).
