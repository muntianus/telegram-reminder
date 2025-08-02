package bot

import (
	"telegram-reminder/internal/logger"

	tb "gopkg.in/telebot.v3"
)

// sendLong splits text into chunks that fit within Telegram's 4096 character
// limit and sends them sequentially.
func sendLong(b *tb.Bot, to tb.Recipient, text string) error {
	if text == "" {
		logger.L.Warn("empty text in sendLong")
		_, err := b.Send(to, "❌ Получен пустой ответ")
		return err
	}

	runes := []rune(text)
	for len(runes) > 0 {
		end := TelegramMessageLimit
		if len(runes) < end {
			end = len(runes)
		}
		if _, err := b.Send(to, string(runes[:end]), tb.ModeHTML); err != nil {
			return err
		}
		runes = runes[end:]
	}
	return nil
}

// replyLong sends the text as a reply in the context's chat, splitting if
// necessary.
func replyLong(c tb.Context, text string) error {
	if text == "" {
		logger.L.Warn("empty text in replyLong")
		return c.Send("❌ Получен пустой ответ")
	}

	runes := []rune(text)
	for len(runes) > 0 {
		end := TelegramMessageLimit
		if len(runes) < end {
			end = len(runes)
		}
		if err := c.Send(string(runes[:end]), tb.ModeHTML); err != nil {
			return err
		}
		runes = runes[end:]
	}
	return nil
}
