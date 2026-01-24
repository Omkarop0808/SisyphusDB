package main

import (
	"fmt"
	"net/http"
	"time"

	"KV-Store/pkg/metrics"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (ww *responseWriterWrapper) WriteHeader(code int) {
	ww.statusCode = code
	ww.ResponseWriter.WriteHeader(code)
}

func withMetrics(handler http.HandlerFunc, method, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: 200}

		handler(ww, r)

		duration := time.Since(start).Seconds()
		metrics.HttpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
		metrics.HttpRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprintf("%d", ww.statusCode)).Inc()
	}
}
