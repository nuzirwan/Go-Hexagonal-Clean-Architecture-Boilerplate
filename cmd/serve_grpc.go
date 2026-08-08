package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	grpchandler "github.com/nuzirwan/go-boilerplate/internal/adapter/in/grpc/handler"
	"github.com/nuzirwan/go-boilerplate/internal/adapter/in/grpc/interceptor"
	asynqadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/asynq"
	obsadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/observability"
	"github.com/nuzirwan/go-boilerplate/internal/adapter/out/postgres"
	redisadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/redis"
	"github.com/nuzirwan/go-boilerplate/internal/domain/usecase"
)

var serveGRPCCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Start gRPC server",
	RunE:  runGRPC,
}

func init() {
	serveCmd.AddCommand(serveGRPCCmd)
}

func runGRPC(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Outbound: Postgres
	db, err := postgres.Connect(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	// Outbound: asynq enqueuer
	enqueuer := asynqadapter.NewTaskEnqueuer(cfg.Redis.Addr)
	defer enqueuer.Close()

	// Wire
	productRepo := postgres.NewProductRepository(db)
	txManager := postgres.NewTxManager(db)

	// Need redis for cache in GetProduct
	redisClient, err := redisadapter.Connect(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer redisClient.Close()
	productCache := redisadapter.NewProductCache(redisClient, 5*time.Minute)

	createProduct := usecase.NewCreateProduct(productRepo, enqueuer, txManager)
	getProduct := usecase.NewGetProduct(productRepo, productCache)
	applyDiscount := usecase.NewApplyDiscount(productRepo, enqueuer, txManager)

	// Wrap usecases with observability
	createProduct = obsadapter.WrapCreateProduct(createProduct)
	getProduct = obsadapter.WrapGetProduct(getProduct)
	applyDiscount = obsadapter.WrapApplyDiscount(applyDiscount)

	// gRPC handler
	_ = grpchandler.NewProductGRPCHandler(grpchandler.ProductGRPCHandlerDeps{
		CreateProduct: createProduct,
		GetProduct:    getProduct,
		ApplyDiscount: applyDiscount,
	})

	// gRPC server
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryRecovery(),
		),
	)

	// TODO: Register proto service after code generation:
	// productv1.RegisterProductServiceServer(srv, productGRPCHandler)

	// Listener
	port := cfg.App.Port + 1 // gRPC on port+1 by default
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	go func() {
		slog.Info("starting gRPC server", "port", port)
		if err := srv.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gRPC server...")
	srv.GracefulStop()
	return nil
}
