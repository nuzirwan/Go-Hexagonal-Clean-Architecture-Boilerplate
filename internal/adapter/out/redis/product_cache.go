package redis

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	goredis "github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"

	"github.com/nuzirwan/go-boilerplate/internal/domain/entity"
	"github.com/nuzirwan/go-boilerplate/pkg/resilience"
)

type ProductCache struct {
	client  *goredis.Client
	ttl     time.Duration
	breaker *gobreaker.CircuitBreaker[*entity.Product]
	timeout time.Duration
}

func NewProductCache(client *goredis.Client, ttl time.Duration) *ProductCache {
	return &ProductCache{
		client:  client,
		ttl:     ttl,
		breaker: resilience.NewCircuitBreaker[*entity.Product](resilience.DefaultCircuitBreakerConfig("redis-product-cache")),
		timeout: 500 * time.Millisecond,
	}
}

func (c *ProductCache) Get(ctx context.Context, id string) (*entity.Product, error) {
	product, err := c.breaker.Execute(func() (*entity.Product, error) {
		ctx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		data, err := c.client.Get(ctx, "product:"+id).Bytes()
		if errors.Is(err, goredis.Nil) {
			return nil, nil // cache miss — not a failure
		}
		if err != nil {
			return nil, err // real error — circuit breaker counts this
		}

		var p entity.Product
		if err := sonic.Unmarshal(data, &p); err != nil {
			return nil, nil // corrupted data — treat as miss, don't trip breaker
		}
		return &p, nil
	})

	if err != nil {
		// Circuit open or timeout — treat as cache miss, don't fail the request
		slog.Warn("redis cache degraded",
			"error", err.Error(),
			"key", "product:"+id,
			"circuit_state", c.breaker.State().String(),
		)
		return nil, nil
	}

	return product, nil
}

// SetAsync writes to cache in a background goroutine — never blocks the response.
func (c *ProductCache) SetAsync(ctx context.Context, p *entity.Product) {
	data, err := sonic.Marshal(p)
	if err != nil {
		return
	}
	go c.client.Set(context.Background(), "product:"+p.ID, data, c.ttl)
}

func (c *ProductCache) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.client.Del(ctx, "product:"+id).Err()
}
