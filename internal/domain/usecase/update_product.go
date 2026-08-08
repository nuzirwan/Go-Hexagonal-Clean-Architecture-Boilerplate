package usecase

import (
	"context"

	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

type UpdateProductUsecase struct {
	repo      port.ProductRepository
	txManager port.TransactionManager
}

func NewUpdateProduct(repo port.ProductRepository, tx port.TransactionManager) port.UpdateProduct {
	return &UpdateProductUsecase{repo: repo, txManager: tx}
}

func (u *UpdateProductUsecase) Execute(ctx context.Context, cmd port.UpdateProductCommand) (port.ProductResult, error) {
	product, err := u.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return port.ProductResult{}, err
	}

	if err := product.Update(cmd.Name, cmd.Price, cmd.Currency, cmd.Stock); err != nil {
		return port.ProductResult{}, err
	}

	err = u.txManager.RunInTx(ctx, func(ctx context.Context) error {
		return u.repo.Save(ctx, product)
	})
	if err != nil {
		return port.ProductResult{}, err
	}

	return toResult(product), nil
}
