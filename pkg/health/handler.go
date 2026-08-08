package health

import (
	"context"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
)

type Handler struct {
	checkers map[string]Checker
}

func NewHandler(checkers map[string]Checker) *Handler {
	return &Handler{checkers: checkers}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Liveness)
	mux.HandleFunc("GET /readyz", h.Readiness)
}

func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	buf, _ := sonic.Marshal(map[string]string{"status": "ok"})
	w.Write(buf)
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	results := make(map[string]string)
	allHealthy := true

	for name, checker := range h.checkers {
		if err := checker.Check(ctx); err != nil {
			results[name] = err.Error()
			allHealthy = false
		} else {
			results[name] = "ok"
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	statusStr := "ok"
	if !allHealthy {
		statusStr = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	buf, _ := sonic.Marshal(map[string]any{
		"status": statusStr,
		"checks": results,
	})
	w.Write(buf)
}
