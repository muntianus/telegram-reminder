# 🔍 Диагностика проблем с деплоем

## 📋 Быстрая проверка

### 1. Проверьте GitHub Actions
1. Перейдите в ваш репозиторий на GitHub
2. Откройте вкладку **Actions**
3. Найдите последний workflow **Build & Deploy**
4. Проверьте статус:
   - ✅ **Зеленая галочка** = деплой успешен
   - ❌ **Красный крест** = есть ошибки
   - 🟡 **Желтый кружок** = деплой в процессе

### 2. Проверьте GitHub Secrets
1. **Settings** → **Secrets and variables** → **Actions**
2. Убедитесь, что есть все секреты:
   ```
   ✅ TELEGRAM_TOKEN
   ✅ OPENAI_API_KEY
   ✅ OPENAI_MODEL (должно быть o3-2025-04-16)
   ✅ DOCKERHUB_USER
   ✅ DOCKERHUB_TOKEN
   ✅ VPS_SSH_KEY
   ✅ VPS_USER
   ✅ VPS_HOST
   ```

### 3. Проверьте VPS
Подключитесь к серверу по SSH и выполните:

```bash
# Проверьте статус контейнера
docker ps

# Проверьте логи
docker logs telegram-reminder -f

# Проверьте переменные окружения
docker exec telegram-reminder env | grep -E "(OPENAI|TELEGRAM)"

# Проверьте версию образа
docker images | grep telegram-reminder
```

## 🚨 Возможные проблемы

### Проблема 1: GitHub Actions не запустился
**Решение:**
- Сделайте новый push в main: `git push origin main`
- Или вручную запустите workflow в Actions

### Проблема 2: Ошибка в GitHub Actions
**Проверьте логи:**
- Откройте failed workflow
- Найдите красную строку с ошибкой
- Исправьте проблему и сделайте новый push

### Проблема 3: Контейнер не запустился
**Проверьте:**
```bash
# На VPS
docker logs telegram-reminder
# Ищите ошибки типа:
# - "missing required env vars"
# - "invalid api key"
# - "model not found"
```

### Проблема 4: Старая версия контейнера
**Решение:**
```bash
# На VPS
docker pull ваш_логин/telegram-reminder:latest
docker compose down
docker compose up -d
```

### Проблема 5: Неправильные переменные окружения
**Проверьте .env файл на VPS:**
```bash
# На VPS
cat .env
# Должно быть:
# TELEGRAM_TOKEN=ваш_токен
# OPENAI_API_KEY=ваш_ключ
# OPENAI_MODEL=o3-2025-04-16
```

## 🔧 Временное решение

Если деплой еще не завершился, можете временно запустить бота локально:

```bash
# Установите переменные окружения
export TELEGRAM_TOKEN="ваш_токен"
export OPENAI_API_KEY="ваш_ключ"
export OPENAI_MODEL="o3-2025-04-16"

# Запустите бота
go run main.go
```

## 📞 Что проверить в первую очередь

1. **GitHub Actions** - завершился ли деплой?
2. **GitHub Secrets** - настроены ли все переменные?
3. **VPS логи** - что показывает контейнер?
4. **Переменные окружения** - правильно ли переданы?

## 🎯 Ожидаемый результат

После успешного деплоя команды должны возвращать:
- ✅ Подробные сообщения об ошибках вместо "OpenAI error"
- ✅ Информативные подсказки по решению проблем
- ✅ Детальные логи для отладки 