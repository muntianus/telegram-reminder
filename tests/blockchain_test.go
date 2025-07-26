package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tb "gopkg.in/telebot.v3"
)

type bcCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

func (b *bcCtx) Send(what interface{}, _ ...interface{}) error {
	b.called = true
	b.msg = what
	return nil
}

func TestBlockchainHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"market_price_usd": 10.5,
			"n_tx":             7,
			"hash_rate":        0.9,
		})
	}))
	defer srv.Close()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	b.Handle("/blockchain", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		if err != nil {
			return c.Send("blockchain error")
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return c.Send("blockchain error")
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()
		var st struct {
			MarketPriceUSD float64 `json:"market_price_usd"`
			NTx            int64   `json:"n_tx"`
			HashRate       float64 `json:"hash_rate"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
			return c.Send("blockchain error")
		}
		msg := fmt.Sprintf("BTC price: $%.2f\nTransactions: %d\nHash rate: %.2f", st.MarketPriceUSD, st.NTx, st.HashRate)
		return c.Send(msg)
	})

	ctx := &bcCtx{}
	if err := b.Trigger("/blockchain", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	want := "BTC price: $10.50\nTransactions: 7\nHash rate: 0.90"
	if ctx.msg != want {
		t.Errorf("unexpected msg: %v", ctx.msg)
	}
}
