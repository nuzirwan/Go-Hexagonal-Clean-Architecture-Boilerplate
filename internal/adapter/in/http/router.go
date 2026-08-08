package inhttp

import (
	"net/http"

	"github.com/nuzirwan/go-boilerplate/internal/adapter/in/http/handler"
	"github.com/nuzirwan/go-boilerplate/internal/adapter/in/http/middleware"
	"github.com/nuzirwan/go-boilerplate/pkg/health"
)

func NewRouter(productHandler *handler.ProductHandler, healthHandler *health.Handler) http.Handler {
	mux := http.NewServeMux()

	// Health probes
	healthHandler.RegisterRoutes(mux)

	// Product routes
	mux.HandleFunc("POST /products", productHandler.Create)
	mux.HandleFunc("GET /products/{id}", productHandler.Get)
	mux.HandleFunc("GET /products", productHandler.List)
	mux.HandleFunc("PUT /products/{id}", productHandler.Update)
	mux.HandleFunc("DELETE /products/{id}", productHandler.Delete)
	mux.HandleFunc("POST /products/{id}/discount", productHandler.ApplyDiscount)

	// Middleware chain (outermost first)
	var root http.Handler = mux
	root = middleware.BodyLimit(1 << 20)(root) // 1MB
	root = middleware.Logger(root)
	root = middleware.RequestID(root)
	root = middleware.Recover(root)
	root = middleware.CORS(root)

	return root
}
