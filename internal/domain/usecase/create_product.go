package usecase

import (
	"context"

	"github.com/nuzirwan/go-boilerplate/internal/domain/entity"
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

type CreateProductUsecase struct {
	repo      port.ProductRepository
	publisher port.EventPublisher
	txManager port.TransactionManager
}

func NewCreateProduct(repo port.ProductRepository, pub port.EventPublisher, tx port.TransactionManager) port.CreateProduct {
	return &CreateProductUsecase{repo: repo, publisher: pub, txManager: tx}
}

func (u *CreateProductUsecase) Execute(ctx context.Context, cmd port.CreateProductCommand) (port.ProductResult, error) {
	product, err := entity.NewProduct(cmd.Name, cmd.Price, cmd.Currency, cmd.Stock)
	if err != nil {
		return port.ProductResult{}, err
	}

	err = u.txManager.RunInTx(ctx, func(ctx context.Context) error {
		return u.repo.Save(ctx, product)
	})
	if err != nil {
		return port.ProductResult{}, err
	}

	if events := product.Events(); len(events) > 0 {
		_ = u.publisher.Publish(ctx, events...)
	}

	return toResult(product), nil
}
