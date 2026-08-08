package usecase

import (
	"context"

	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

type DeleteProductUsecase struct {
	repo port.ProductRepository
}

func NewDeleteProduct(repo port.ProductRepository) port.DeleteProduct {
	return &DeleteProductUsecase{repo: repo}
}

func (u *DeleteProductUsecase) Execute(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}
