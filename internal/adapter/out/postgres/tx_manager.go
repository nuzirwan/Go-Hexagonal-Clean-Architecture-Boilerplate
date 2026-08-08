package postgres

import (
	"context"
	"database/sql"

	"github.com/nuzirwan/go-boilerplate/pkg/dbtx"
)

type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// RunInTx executes fn within a database transaction.
// IMPORTANT: The provided ctx carries the transaction.
// Do NOT pass this ctx to goroutines — sql.Tx is not safe for concurrent use.
func (m *TxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := dbtx.WithTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
