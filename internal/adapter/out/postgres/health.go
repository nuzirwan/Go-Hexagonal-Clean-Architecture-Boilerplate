package postgres

import (
	"context"
	"database/sql"
)

type HealthChecker struct {
	db *sql.DB
}

func NewHealthChecker(db *sql.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

func (h *HealthChecker) Check(ctx context.Context) error {
	return h.db.PingContext(ctx)
}
