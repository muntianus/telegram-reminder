package main

import (
	"log"

	"telegram-reminder/internal/bot"
	"telegram-reminder/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := bot.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
