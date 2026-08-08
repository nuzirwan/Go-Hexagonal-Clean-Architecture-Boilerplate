package resilience

import (
	"time"

	"github.com/sony/gobreaker"
)

type CircuitBreakerConfig struct {
	Name        string
	MaxRequests uint32        // max requests allowed in half-open state
	Interval    time.Duration // cyclic period of closed state to clear counts
	Timeout     time.Duration // period of open state before becoming half-open
}

func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
	}
}

func NewCircuitBreaker(cfg CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	})
}
