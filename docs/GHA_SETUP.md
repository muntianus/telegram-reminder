# 🔧 Настройка GitHub Actions для модели gpt-4.1

## 📋 Требования

Для правильной работы бота с моделью `gpt-4.1` в GitHub Actions нужно настроить следующие секреты:

### 1. GitHub Secrets

Перейдите в ваш репозиторий:
**Settings** → **Secrets and variables** → **Actions**

Добавьте следующие секреты:

#### 🔑 OpenAI API
- **Name**: `OPENAI_API_KEY`
- **Value**: `sk-...` (ваш API ключ OpenAI)
- **Note**: Модель gpt-4.1 установлена по умолчанию в коде

#### 🤖 Telegram Bot
- **Name**: `TELEGRAM_BOT_TOKEN`
- **Value**: `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz` (токен от @BotFather)

#### 🐳 Docker Hub
- **Name**: `DOCKERHUB_USER`
- **Value**: `ваш_логин_dockerhub`

- **Name**: `DOCKERHUB_TOKEN`
- **Value**: `ваш_токен_dockerhub`

#### 🖥️ VPS (опционально)
- **Name**: `VPS_SSH_KEY`
- **Value**: `-----BEGIN OPENSSH PRIVATE KEY-----...`

- **Name**: `VPS_USER`
- **Value**: `root` (или ваш пользователь)

- **Name**: `VPS_HOST`
- **Value**: `ваш_сервер.com` (IP или домен)

### 2. Переменные окружения

В файле `.github/workflows/deploy.yml` настроены только необходимые переменные:

```yaml
env:
  "OPENAI_API_KEY": ${{ secrets.OPENAI_API_KEY }}
  "TELEGRAM_BOT_TOKEN": ${{ secrets.TELEGRAM_BOT_TOKEN }}
```

**Примечание**: Модель `gpt-4.1` установлена по умолчанию в коде приложения.

## 🚀 Запуск деплоя

### Автоматический деплой
1. Сделайте push в ветку `main`
2. GitHub Actions автоматически запустится
3. Проверьте статус в **Actions** → **Build & Deploy**

### Ручной запуск
1. Перейдите в **Actions**
2. Выберите workflow **Build & Deploy**
3. Нажмите **Run workflow**
4. Выберите ветку `main`
5. Нажмите **Run workflow**

## 📊 Мониторинг процесса

### Этапы деплоя:
1. **Checkout** - клонирование репозитория
2. **Set up Go** - установка Go 1.21
3. **Build** - сборка приложения
4. **Login to Docker Hub** - авторизация в Docker Hub
5. **Build and push Docker image** - сборка и загрузка образа
6. **Deploy to VPS** - деплой на сервер (если настроен)

### Ожидаемое время:
- **Сборка**: 2-3 минуты
- **Загрузка образа**: 1-2 минуты
- **Деплой**: 30-60 секунд

## ✅ Проверка успешности

### 1. GitHub Actions
- ✅ Все этапы зеленые
- ✅ Нет ошибок в логах
- ✅ Образ загружен в Docker Hub

### 2. Docker Hub
- ✅ Образ `ваш_логин/telegram-reminder:latest` обновлен
- ✅ Размер образа ~50-100MB

### 3. VPS (если настроен)
- ✅ Контейнер запущен: `docker ps | grep telegram-reminder`
- ✅ Логи без ошибок: `docker logs telegram-reminder-bot`
- ✅ Переменные окружения: `docker exec telegram-reminder-bot env | grep -E "(OPENAI|TELEGRAM)"`

### 4. Тестирование бота
- ✅ `/ping` - отвечает "pong"
- ✅ `/model` - должна показать `gpt-4.1`
- ✅ `/crypto` - должен работать с gpt-4.1
- ✅ `/realestate` - должен работать с gpt-4.1
- ✅ `/business` - должен работать с gpt-4.1

## 🔧 Устранение проблем

### Ошибка сборки
```
❌ Build failed
```
**Решение:**
1. Проверьте синтаксис Go кода
2. Убедитесь, что все зависимости в `go.mod`
3. Проверьте права доступа к репозиторию

### Ошибка Docker Hub
```
❌ Login failed
```
**Решение:**
1. Проверьте `DOCKERHUB_USER` и `DOCKERHUB_TOKEN`
2. Убедитесь, что токен не истек
3. Проверьте права доступа к репозиторию

### Ошибка VPS
```
❌ SSH connection failed
```
**Решение:**
1. Проверьте `VPS_SSH_KEY`, `VPS_USER`, `VPS_HOST`
2. Убедитесь, что SSH ключ добавлен на сервер
3. Проверьте доступность сервера

### Ошибка OpenAI
```
❌ Model gpt-4.1 not found
```
**Решение:**
1. Проверьте `OPENAI_API_KEY`
2. Убедитесь, что модель gpt-4.1 доступна в вашем аккаунте
3. Проверьте баланс на аккаунте OpenAI

## 📝 Логи и отладка

### Просмотр логов GitHub Actions
1. Откройте failed workflow
2. Нажмите на красный этап
3. Разверните детали ошибки

### Просмотр логов VPS
```bash
# Логи контейнера
docker logs telegram-reminder-bot

# Логи в реальном времени
docker logs -f telegram-reminder-bot

# Переменные окружения
docker exec telegram-reminder-bot env | grep -E "(OPENAI|TELEGRAM)"
```

## 🎯 Результат

После успешного деплоя:

1. **Бот будет использовать модель gpt-4.1**
2. **Веб-поиск будет работать** для получения актуальной информации
3. **Все команды будут отвечать** с свежими данными из интернета
4. **Автоматические задачи** будут выполняться по расписанию

## 🔄 Обновление

Для обновления бота:
1. Внесите изменения в код
2. Сделайте `git push origin main`
3. GitHub Actions автоматически пересоберет и задеплоит

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи GitHub Actions
2. Проверьте логи VPS
3. Проверьте переменные окружения
4. Убедитесь в доступности модели gpt-4.1 в вашем аккаунте OpenAI