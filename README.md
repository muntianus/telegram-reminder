# Telegram ChatGPT Bot

This project is a simple Telegram bot written in Go. It forwards any text message it receives to ChatGPT (via the OpenAI API) and sends back the response.

## Prerequisites

- Go 1.20+
- Telegram bot token
- OpenAI API key

## Setup

1. Set environment variables:

```sh
export TELEGRAM_TOKEN=your_telegram_token
export OPENAI_API_KEY=your_openai_key
```

2. Run the bot:

```sh
go run main.go
```

Send any message to the bot in Telegram and it will reply using ChatGPT.

