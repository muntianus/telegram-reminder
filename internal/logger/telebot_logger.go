package logger

import (
	tb "gopkg.in/telebot.v3"
)

// TelebotMiddleware logs every incoming update at debug level and reports handler errors.
func TelebotMiddleware() tb.MiddlewareFunc {
	return func(next tb.HandlerFunc) tb.HandlerFunc {
		return func(c tb.Context) error {
			if m := c.Message(); m != nil {
				L.Debug("update message", "chat", m.Chat.ID, "text", m.Text)
			} else if cb := c.Callback(); cb != nil {
				L.Debug("update callback", "chat", cb.Message.Chat.ID, "data", cb.Data)
			} else {
				up := c.Update()
				L.Debug("update", "id", up.ID)
			}
			err := next(c)
			if err != nil {
				L.Error("handler error", "err", err)
			}
			return err
		}
	}
}
