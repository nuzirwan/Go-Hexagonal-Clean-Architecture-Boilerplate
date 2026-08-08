package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/nuzirwan/go-boilerplate/pkg/httputil"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(sr, r)

		slog.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sr.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", httputil.GetRequestID(r.Context()),
		)
	})
}
