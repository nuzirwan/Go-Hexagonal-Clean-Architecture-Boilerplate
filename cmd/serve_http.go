package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	inhttp "github.com/nuzirwan/go-boilerplate/internal/adapter/in/http"
	"github.com/nuzirwan/go-boilerplate/internal/adapter/in/http/handler"
	asynqadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/asynq"
	obsadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/observability"
	"github.com/nuzirwan/go-boilerplate/internal/adapter/out/postgres"
	redisadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/redis"
	"github.com/nuzirwan/go-boilerplate/internal/domain/usecase"
	"github.com/nuzirwan/go-boilerplate/pkg/health"
	"github.com/nuzirwan/go-boilerplate/pkg/shutdown"
)

var serveHTTPCmd = &cobra.Command{
	Use:   "http",
	Short: "Start HTTP server",
	RunE:  runHTTP,
}

func runHTTP(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Outbound: Postgres
	db, err := postgres.Connect(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	// Outbound: Redis
	redisClient, err := redisadapter.Connect(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer redisClient.Close()

	// Outbound: asynq
	enqueuer := asynqadapter.NewTaskEnqueuer(cfg.Redis.Addr)
	defer enqueuer.Close()

	// Wire outbound adapters
	productRepo := postgres.NewProductRepository(db)
	txManager := postgres.NewTxManager(db)
	productCache := redisadapter.NewProductCache(redisClient, 5*time.Minute)

	// Wire use cases
	createProduct := usecase.NewCreateProduct(productRepo, enqueuer, txManager)
	getProduct := usecase.NewGetProduct(productRepo, productCache)
	listProducts := usecase.NewListProducts(productRepo, getProduct)
	updateProduct := usecase.NewUpdateProduct(productRepo, txManager)
	deleteProduct := usecase.NewDeleteProduct(productRepo)
	applyDiscount := usecase.NewApplyDiscount(productRepo, enqueuer, txManager)

	// Wrap usecases with observability
	createProduct = obsadapter.WrapCreateProduct(createProduct)
	getProduct = obsadapter.WrapGetProduct(getProduct)
	listProducts = obsadapter.WrapListProducts(listProducts)
	updateProduct = obsadapter.WrapUpdateProduct(updateProduct)
	deleteProduct = obsadapter.WrapDeleteProduct(deleteProduct)
	applyDiscount = obsadapter.WrapApplyDiscount(applyDiscount)

	// Wire inbound: HTTP handler
	productHandler := handler.NewProductHandler(handler.ProductHandlerDeps{
		CreateProduct: createProduct,
		GetProduct:    getProduct,
		ListProducts:  listProducts,
		UpdateProduct: updateProduct,
		DeleteProduct: deleteProduct,
		ApplyDiscount: applyDiscount,
	})

	// Health probes
	healthHandler := health.NewHandler(map[string]health.Checker{
		"postgres": postgres.NewHealthChecker(db),
		"redis":    redisadapter.NewHealthChecker(redisClient),
	})

	// Router (wires handler + health + middleware)
	router := inhttp.NewRouter(productHandler, healthHandler)

	// HTTP Server
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	slog.Info("starting HTTP server", "port", cfg.App.Port, "env", cfg.App.Env)
	return shutdown.Serve(ctx, srv, 10*time.Second)
}
