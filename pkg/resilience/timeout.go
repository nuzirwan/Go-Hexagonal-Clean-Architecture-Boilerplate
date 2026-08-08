package resilience

import (
	"context"
	"time"
)

// WithTimeout creates a context with the given timeout.
// Caller must call the returned cancel function.
func WithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}
