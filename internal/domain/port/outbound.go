package port

//go:generate mockgen -source=outbound.go -destination=../../mocks/mock_outbound.go -package=mocks

import (
	"context"

	"github.com/nuzirwan/go-boilerplate/internal/domain/entity"
	"github.com/nuzirwan/go-boilerplate/internal/domain/event"
)

type ProductRepository interface {
	Save(ctx context.Context, product *entity.Product) error
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	FindAll(ctx context.Context, filter ProductFilter) ([]*entity.Product, string, error)
	FindAllIDs(ctx context.Context, filter ProductFilter) ([]string, string, error)
	Delete(ctx context.Context, id string) error
}

type ProductCache interface {
	Get(ctx context.Context, id string) (*entity.Product, error)
	SetAsync(ctx context.Context, product *entity.Product)
	Delete(ctx context.Context, id string) error
}

type EventPublisher interface {
	Publish(ctx context.Context, events ...event.Event) error
}

type TransactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type ProductFilter struct {
	Cursor string
	Limit  int
	Status string
}
