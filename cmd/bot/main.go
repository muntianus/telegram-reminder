package main

import (
	"os"

	"telegram-reminder/internal/bot"
	"telegram-reminder/internal/config"
	"telegram-reminder/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.L.Error("config load", "err", err)
		os.Exit(1)
	}
	if err := bot.Run(cfg); err != nil {
		logger.L.Error("bot run", "err", err)
		os.Exit(1)
	}
}
