package usecase

import (
	"context"

	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

type ApplyDiscountUsecase struct {
	repo      port.ProductRepository
	publisher port.EventPublisher
	txManager port.TransactionManager
}

func NewApplyDiscount(repo port.ProductRepository, pub port.EventPublisher, tx port.TransactionManager) port.ApplyDiscount {
	return &ApplyDiscountUsecase{repo: repo, publisher: pub, txManager: tx}
}

func (u *ApplyDiscountUsecase) Execute(ctx context.Context, cmd port.ApplyDiscountCommand) (port.ProductResult, error) {
	product, err := u.repo.FindByID(ctx, cmd.ProductID)
	if err != nil {
		return port.ProductResult{}, err
	}

	if err := product.ApplyDiscount(cmd.Percentage); err != nil {
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
