package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/nuzirwan/go-boilerplate/internal/domain/entity"
	domainerror "github.com/nuzirwan/go-boilerplate/internal/domain/error"
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
	"github.com/nuzirwan/go-boilerplate/pkg/dbtx"
	"github.com/nuzirwan/go-boilerplate/pkg/resilience"
)

type ProductRepository struct {
	db    *sql.DB
	sf    singleflight.Group
	retry *resilience.Retry
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db:    db,
		retry: resilience.NewRetry(resilience.RetryConfig{MaxAttempts: 2, BaseDelay: 50 * time.Millisecond, MaxDelay: 200 * time.Millisecond}),
	}
}

func (r *ProductRepository) Save(ctx context.Context, p *entity.Product) error {
	db := dbtx.GetDBTX(ctx, r.db)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		INSERT INTO products (id, name, price, currency, stock, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = $2, price = $3, currency = $4, stock = $5, status = $6, updated_at = $8
	`, p.ID, p.Name, p.Price, p.Currency, p.Stock, p.Status, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	// Singleflight: deduplicate concurrent requests for same ID
	val, err, _ := r.sf.Do("find:"+id, func() (any, error) {
		// Retry on transient errors (connection reset, bad conn)
		var product *entity.Product
		retryErr := r.retry.Do(ctx, func(ctx context.Context) error {
			p, err := r.findByID(ctx, id)
			if err != nil {
				if isTransient(err) {
					return err // retry
				}
				product = nil
				return &nonRetryable{err} // don't retry domain/logic errors
			}
			product = p
			return nil
		})
		if retryErr != nil {
			var nr *nonRetryable
			if errors.As(retryErr, &nr) {
				return nil, nr.err
			}
			return nil, retryErr
		}
		return product, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*entity.Product), nil
}

func (r *ProductRepository) findByID(ctx context.Context, id string) (*entity.Product, error) {
	db := dbtx.GetDBTX(ctx, r.db)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var p entity.Product
	err := db.QueryRowContext(ctx, `
		SELECT id, name, price, currency, stock, status, created_at, updated_at
		FROM products WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Price, &p.Currency, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainerror.New(domainerror.ErrNotFound, "product not found")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) FindAll(ctx context.Context, filter port.ProductFilter) ([]*entity.Product, string, error) {
	db := dbtx.GetDBTX(ctx, r.db)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, name, price, currency, stock, status, created_at, updated_at FROM products WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.Status != "" {
		query += ` AND status = $` + strconv.Itoa(argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Cursor != "" {
		query += ` AND id < $` + strconv.Itoa(argIdx)
		args = append(args, filter.Cursor)
		argIdx++
	}

	query += ` ORDER BY id DESC LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, filter.Limit+1)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Currency, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, "", err
		}
		products = append(products, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(products) > filter.Limit {
		products = products[:filter.Limit]
		nextCursor = products[len(products)-1].ID
	}

	return products, nextCursor, nil
}

func (r *ProductRepository) FindAllIDs(ctx context.Context, filter port.ProductFilter) ([]string, string, error) {
	db := dbtx.GetDBTX(ctx, r.db)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id FROM products WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.Status != "" {
		query += ` AND status = $` + strconv.Itoa(argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Cursor != "" {
		query += ` AND id < $` + strconv.Itoa(argIdx)
		args = append(args, filter.Cursor)
		argIdx++
	}

	query += ` ORDER BY id DESC LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, filter.Limit+1)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, "", err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(ids) > filter.Limit {
		ids = ids[:filter.Limit]
		nextCursor = ids[len(ids)-1]
	}

	return ids, nextCursor, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	db := dbtx.GetDBTX(ctx, r.db)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domainerror.New(domainerror.ErrNotFound, "product not found")
	}
	return nil
}
