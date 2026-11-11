package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if rw.code == 0 {
		rw.code = statusCode
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(wrapper, r)
		// some endpoints don't write any reply
		if wrapper.code == 0 {
			wrapper.code = 200
		}
		label := fmt.Sprintf(
			"{method=%q, url=%q, status=%q}", r.Method, r.URL.Path, http.StatusText(wrapper.code))
		metrics.GetOrCreateHistogram("http_request_duration_seconds" + label).UpdateDuration(start)
	})
}
