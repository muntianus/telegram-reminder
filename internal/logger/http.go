package logger

import (
	"net/http"
	"time"
)

type loggingTransport struct{ base http.RoundTripper }

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// HTTP request/response debug logging removed to prevent Telegram spam
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		L.Error("http error", "method", req.Method, "url", req.URL.String(), "err", err)
		return nil, err
	}
	return resp, nil
}

// NewTransport wraps the given RoundTripper with logging at debug level.
func NewTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return loggingTransport{base: base}
}

// NewHTTPClient returns an http.Client with the logging transport and timeout.
func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, Transport: NewTransport(http.DefaultTransport)}
}
