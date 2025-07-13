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

The bot logs a message like `✅ Bot up as @mybot (ID: 123)` when it starts so you know it is running. Verbose logging of API requests is enabled by default.

Send any message to the bot in Telegram and it will reply using ChatGPT.

To forward a specific instruction to ChatGPT, use the `/task` command followed by your text:

```text
/task Rewrite my homework in bullet points
```

The bot will send the request to ChatGPT and reply with the result in the chat.

Other useful commands:

```
/start - show a welcome message
/help  - list available commands
/ping  - check if the bot is responsive
```


## Running on a VPS

1. Install Go on your server (for Ubuntu: `sudo apt-get install golang`).
2. Clone this repository on the VPS.
3. Set the `TELEGRAM_TOKEN` and `OPENAI_API_KEY` environment variables.
4. Build the binary:

```sh
go build -o bot
```

5. Run the bot with `./bot` or configure a process manager like `systemd`.

Example `systemd` service file:

```ini
[Unit]
Description=Telegram ChatGPT Bot
After=network.target

[Service]
WorkingDirectory=/path/to/telegram-reminder
Environment=TELEGRAM_TOKEN=your_token
Environment=OPENAI_API_KEY=your_api_key
ExecStart=/path/to/telegram-reminder/bot
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```sh
sudo systemctl enable --now telegram-bot.service
```

