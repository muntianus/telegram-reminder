package logger

import (
	"net/http"
	"time"
)

type loggingTransport struct{ base http.RoundTripper }

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	if req.Body != nil && req.ContentLength > 0 {
		L.Debug("http request", "method", req.Method, "url", req.URL.String(), "bytes", req.ContentLength)
	} else {
		L.Debug("http request", "method", req.Method, "url", req.URL.String())
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		L.Error("http error", "method", req.Method, "url", req.URL.String(), "err", err)
		return nil, err
	}
	L.Debug("http response", "method", req.Method, "url", req.URL.String(), "status", resp.StatusCode, "duration", time.Since(start))
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
