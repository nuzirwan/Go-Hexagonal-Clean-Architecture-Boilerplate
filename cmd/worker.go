package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"

	"github.com/nuzirwan/go-boilerplate/internal/adapter/in/worker"
	asynqadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/asynq"
	"github.com/nuzirwan/go-boilerplate/internal/adapter/out/postgres"
	"github.com/nuzirwan/go-boilerplate/internal/domain/usecase"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start asynq task worker",
	RunE:  runWorker,
}

func init() {
	rootCmd.AddCommand(workerCmd)
}

func runWorker(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Outbound: Postgres (for repository)
	db, err := postgres.Connect(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	// Outbound: asynq enqueuer (for use case dependency — events from use cases)
	enqueuer := asynqadapter.NewTaskEnqueuer(cfg.Redis.Addr)
	defer enqueuer.Close()

	// Wire
	productRepo := postgres.NewProductRepository(db)
	txManager := postgres.NewTxManager(db)
	applyDiscount := usecase.NewApplyDiscount(productRepo, enqueuer, txManager)

	// Task handler
	taskHandler := worker.NewProductTaskHandler(applyDiscount)

	// asynq server (manages its own Redis connection)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	taskHandler.Register(mux)

	go func() {
		slog.Info("starting asynq worker", "concurrency", 10)
		if err := srv.Run(mux); err != nil {
			slog.Error("asynq worker failed", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down asynq worker...")
	srv.Shutdown()
	return nil
}
