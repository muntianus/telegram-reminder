package bot

import (
	tb "gopkg.in/telebot.v3"
)

// sendLong splits text into chunks that fit within Telegram's 4096 character
// limit and sends them sequentially.
func sendLong(b *tb.Bot, to tb.Recipient, text string) error {
	runes := []rune(text)
	for len(runes) > 0 {
		end := TelegramMessageLimit
		if len(runes) < end {
			end = len(runes)
		}
		if _, err := b.Send(to, string(runes[:end])); err != nil {
			return err
		}
		runes = runes[end:]
	}
	return nil
}

// replyLong sends the text as a reply in the context's chat, splitting if
// necessary.
func replyLong(c tb.Context, text string) error {
	runes := []rune(text)
	for len(runes) > 0 {
		end := TelegramMessageLimit
		if len(runes) < end {
			end = len(runes)
		}
		if err := c.Send(string(runes[:end])); err != nil {
			return err
		}
		runes = runes[end:]
	}
	return nil
}
