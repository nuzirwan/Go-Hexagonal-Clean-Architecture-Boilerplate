package shutdown

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func Serve(ctx context.Context, srv *http.Server, timeout time.Duration) error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
