package resilience

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

type Retry struct {
	config RetryConfig
}

func NewRetry(cfg RetryConfig) *Retry {
	return &Retry{config: cfg}
}

// Do executes fn with exponential backoff and jitter.
// It stops retrying if ctx is cancelled or max attempts reached.
func (r *Retry) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error
	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		if attempt == r.config.MaxAttempts-1 {
			break
		}

		// Exponential backoff with jitter
		delay := r.backoff(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastErr
}

func (r *Retry) backoff(attempt int) time.Duration {
	delay := float64(r.config.BaseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}
	// Add jitter: 0.5x to 1.5x
	jitter := 0.5 + rand.Float64()
	return time.Duration(delay * jitter)
}
