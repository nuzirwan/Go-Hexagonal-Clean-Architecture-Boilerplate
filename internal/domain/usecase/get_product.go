package usecase

import (
	"context"

	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

type GetProductUsecase struct {
	repo  port.ProductRepository
	cache port.ProductCache
}

func NewGetProduct(repo port.ProductRepository, cache port.ProductCache) port.GetProduct {
	return &GetProductUsecase{repo: repo, cache: cache}
}

func (u *GetProductUsecase) Execute(ctx context.Context, id string) (port.ProductResult, error) {
	// 1. Check cache
	if cached, err := u.cache.Get(ctx, id); err == nil && cached != nil {
		return toResult(cached), nil
	}

	// 2. DB fetch on cache miss
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return port.ProductResult{}, err
	}

	// 3. Async cache write
	u.cache.SetAsync(ctx, product)

	return toResult(product), nil
}
