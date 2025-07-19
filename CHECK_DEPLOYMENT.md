# 🔍 Проверка деплоя Telegram Reminder Bot

## 📋 Чек-лист проверки

### 1. GitHub Actions
- [ ] Workflow `deploy.yml` запустился успешно
- [ ] Docker образ собрался без ошибок
- [ ] Образ загружен в Docker Hub
- [ ] Деплой на сервер прошел успешно

### 2. Переменные окружения
- [ ] `OPENAI_API_KEY` установлен
- [ ] `TELEGRAM_BOT_TOKEN` установлен  
- [ ] `CHAT_ID` (опционально, для тестовых сообщений)
- [ ] Модель gpt-4o установлена по умолчанию в коде

### 3. Логи приложения
- [ ] Бот запустился без ошибок
- [ ] Подключение к Telegram API успешно
- [ ] Подключение к OpenAI API успешно
- [ ] Нет ошибок в логах

### 4. Тестирование команд
- [ ] `/ping` - отвечает "pong"
- [ ] `/start` - добавляет чат в whitelist
- [ ] `/model` - показывает текущую модель (gpt-4o)
- [ ] `/crypto` - генерирует крипто-дайджест
- [ ] `/tech` - генерирует тех-дайджест
- [ ] `/realestate` - генерирует дайджест недвижимости
- [ ] `/business` - генерирует бизнес-дайджест

## 🚀 Команды для проверки

### Проверка статуса
```bash
# Проверка логов
docker logs telegram-reminder-bot

# Проверка переменных окружения
docker exec telegram-reminder-bot env | grep -E "(OPENAI|TELEGRAM)"

# Проверка процесса
docker ps | grep telegram-reminder
```

### Тестирование команд
```
/ping
/start
/model
/crypto
/tech
/realestate
/business
```

## 🔧 Устранение проблем

### Ошибка OpenAI API
```
❌ Ошибка OpenAI: invalid_api_key
```
**Решение:** Проверьте `OPENAI_API_KEY` в секретах GitHub

### Ошибка Telegram API
```
❌ Ошибка Telegram: unauthorized
```
**Решение:** Проверьте `TELEGRAM_BOT_TOKEN` в секретах GitHub

### Ошибка модели
```
❌ Модель gpt-4o недоступна
```
**Решение:** Проверьте доступность модели в аккаунте OpenAI

### Бот не отвечает
```
❌ Бот не отвечает на команды
```
**Решение:** 
1. Проверьте логи: `docker logs telegram-reminder-bot`
2. Проверьте статус: `docker ps | grep telegram-reminder`
3. Перезапустите: `docker restart telegram-reminder-bot`

## 📊 Мониторинг

### Логи в реальном времени
```bash
docker logs -f telegram-reminder-bot
```

### Статистика использования
```bash
# Количество запросов к OpenAI
docker logs telegram-reminder-bot | grep "openai request" | wc -l

# Количество сообщений в Telegram
docker logs telegram-reminder-bot | grep "telegram send" | wc -l
```

## 🎯 Ожидаемое поведение

### При запуске
```
✅ Бот запущен
✅ Подключение к Telegram API
✅ Подключение к OpenAI API
✅ Модель: gpt-4o
✅ Запланированные задачи активны
```

### При команде /crypto
```
 КРИПТО-ДАЙДЖЕСТ НА СЕГОДНЯ (2025-01-27)

🔍 АКТУАЛЬНАЯ ИНФОРМАЦИЯ ИЗ ИНТЕРНЕТА:
📊 Поиск: bitcoin news today
• Bitcoin Surges 5% After ETF Approval News
  Источник: https://cointelegraph.com/news/bitcoin-surges-5-percent-after-etf-approval

 РЫНОЧНЫЕ МЕТРИКИ:
- Bitcoin вырос на 5% после новости об одобрении ETF
- [Источник: Cointelegraph](https://cointelegraph.com/news/bitcoin-surges-5-percent-after-etf-approval)
```

## 🔄 Перезапуск

### Полный перезапуск
```bash
# Остановка
docker stop telegram-reminder-bot
docker rm telegram-reminder-bot

# Запуск
docker run -d \
  --name telegram-reminder-bot \
  --restart unless-stopped \
  -e "TELEGRAM_TOKEN=${TELEGRAM_TOKEN}" \
  -e "OPENAI_API_KEY=${OPENAI_API_KEY}" \
  -e "OPENAI_MODEL=gpt-4o" \
  -e "CHAT_ID=${CHAT_ID}" \
  ${DOCKERHUB_USER}/telegram-reminder:latest
```

### Быстрый перезапуск
```bash
docker restart telegram-reminder-bot
```

## 📝 Примечания

1. **Модель gpt-4o** - основная модель с веб-поиском
2. **Время ожидания** - 40 секунд для OpenAI API
3. **Логирование** - все запросы логируются
4. **Автоперезапуск** - контейнер перезапускается при сбоях

## 🆘 Поддержка

При возникновении проблем:
1. Проверьте логи: `docker logs telegram-reminder-bot`
2. Проверьте переменные окружения
3. Проверьте доступность API (OpenAI, Telegram)
4. Проверьте доступность модели gpt-4o в вашем аккаунте OpenAI 