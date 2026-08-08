package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type HealthChecker struct {
	client *goredis.Client
}

func NewHealthChecker(client *goredis.Client) *HealthChecker {
	return &HealthChecker{client: client}
}

func (h *HealthChecker) Check(ctx context.Context) error {
	return h.client.Ping(ctx).Err()
}
