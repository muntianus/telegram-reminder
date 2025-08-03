package logger

import (
	tb "gopkg.in/telebot.v3"
)

// TelebotMiddleware reports handler errors only, without debug spam.
func TelebotMiddleware() tb.MiddlewareFunc {
	return func(next tb.HandlerFunc) tb.HandlerFunc {
		return func(c tb.Context) error {
			err := next(c)
			if err != nil {
				L.Error("handler error", "err", err)
			}
			return err
		}
	}
}
