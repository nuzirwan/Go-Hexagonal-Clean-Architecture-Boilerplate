package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/nuzirwan/go-boilerplate/pkg/httputil"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
				)
				httputil.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
