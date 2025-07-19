# 🔧 Настройка GitHub Actions для модели o3-2025-04-16

## 📋 Необходимые GitHub Secrets

Для правильной работы бота с моделью `o3-2025-04-16` в GitHub Actions нужно настроить следующие секреты:

### 🔑 Обязательные секреты

1. **`TELEGRAM_TOKEN`** - токен вашего Telegram бота
   - Получите у @BotFather в Telegram
   - Формат: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`

2. **`OPENAI_API_KEY`** - ключ API OpenAI
   - Получите на https://platform.openai.com/api-keys
   - Формат: `sk-...`

3. **`OPENAI_MODEL`** - модель OpenAI (НОВОЕ!)
   - Значение: `o3-2025-04-16`
   - Это самая новая и мощная модель OpenAI

### 🐳 Docker секреты

4. **`DOCKERHUB_USER`** - имя пользователя Docker Hub
5. **`DOCKERHUB_TOKEN`** - токен доступа Docker Hub

### 🖥️ VPS секреты

6. **`VPS_SSH_KEY`** - приватный SSH ключ для подключения к серверу
7. **`VPS_USER`** - имя пользователя на VPS
8. **`VPS_HOST`** - IP адрес или домен VPS

### 📱 Опциональные секреты

9. **`CHAT_ID`** - ID чата для отправки сообщений (опционально)

## ⚙️ Настройка в GitHub

### Шаг 1: Перейдите в настройки репозитория
1. Откройте ваш репозиторий на GitHub
2. Перейдите в **Settings** → **Secrets and variables** → **Actions**

### Шаг 2: Добавьте секреты
Нажмите **New repository secret** и добавьте каждый секрет:

```
Name: TELEGRAM_TOKEN
Value: ваш_токен_телеграм

Name: OPENAI_API_KEY  
Value: ваш_ключ_openai

Name: OPENAI_MODEL
Value: o3-2025-04-16

Name: DOCKERHUB_USER
Value: ваш_логин_dockerhub

Name: DOCKERHUB_TOKEN
Value: ваш_токен_dockerhub

Name: VPS_SSH_KEY
Value: -----BEGIN OPENSSH PRIVATE KEY-----
        ваш_приватный_ключ
        -----END OPENSSH PRIVATE KEY-----

Name: VPS_USER
Value: root

Name: VPS_HOST
Value: ваш_ip_или_домен

Name: CHAT_ID (опционально)
Value: 123456789
```

## 🚀 Запуск деплоя

После настройки всех секретов:

1. **Сделайте push в ветку `main`**
2. **GitHub Actions автоматически запустит деплой**
3. **Проверьте логи в Actions** → **Build & Deploy**

## 🔍 Проверка работы

После успешного деплоя:

1. **Подключитесь к VPS по SSH**
2. **Проверьте статус контейнера:**
   ```bash
   docker ps
   docker logs telegram-reminder
   ```

3. **Протестируйте команды в Telegram:**
   - `/model` - должна показать `o3-2025-04-16`
   - `/crypto` - должен работать с o3-2025-04-16
   - `/realestate` - должен работать с o3-2025-04-16
   - `/business` - должен работать с o3-2025-04-16

## 🛠️ Устранение неполадок

### Ошибка "OpenAI error"
1. **Проверьте `OPENAI_API_KEY`** - должен быть действительным
2. **Проверьте `OPENAI_MODEL`** - должно быть `o3-2025-04-16`
3. **Проверьте баланс OpenAI** - должно быть достаточно кредитов
4. **Проверьте доступ к o3** - убедитесь, что ваш аккаунт поддерживает o3

### Ошибка деплоя
1. **Проверьте все секреты** - все обязательные секреты должны быть настроены
2. **Проверьте SSH ключ** - должен быть в правильном формате
3. **Проверьте права доступа** - VPS_USER должен иметь права на Docker

### Проверка логов
```bash
# На VPS
docker logs telegram-reminder -f

# В GitHub Actions
Actions → Build & Deploy → Deploy to VPS → View logs
```

## 📊 Мониторинг

### Проверка переменных окружения
```bash
# На VPS
docker exec telegram-reminder env | grep -E "(OPENAI|TELEGRAM)"
```

### Проверка модели
```bash
# В Telegram боте
/model
```

## 🎯 Результат

После правильной настройки:
- ✅ Бот будет использовать модель `o3-2025-04-16`
- ✅ Все дайджесты будут генерироваться с максимальным качеством
- ✅ Автоматический деплой при каждом push в main
- ✅ Логирование и мониторинг

## 📞 Поддержка

Если возникли проблемы:
1. Проверьте логи GitHub Actions
2. Проверьте логи контейнера на VPS
3. Убедитесь, что все секреты настроены правильно
4. Проверьте доступность модели o3-2025-04-16 в вашем аккаунте OpenAI 