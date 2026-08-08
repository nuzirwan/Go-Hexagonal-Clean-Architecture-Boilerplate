package usecase

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

type ListProductsUsecase struct {
	repo       port.ProductRepository
	getProduct port.GetProduct
}

func NewListProducts(repo port.ProductRepository, getProduct port.GetProduct) port.ListProducts {
	return &ListProductsUsecase{repo: repo, getProduct: getProduct}
}

func (u *ListProductsUsecase) Execute(ctx context.Context, query port.ListProductsQuery) (port.ListProductsResult, error) {
	// Phase 1: Get ordered IDs (index-only scan, cheap)
	ids, nextCursor, err := u.repo.FindAllIDs(ctx, port.ProductFilter{
		Cursor: query.Cursor,
		Limit:  query.Limit,
		Status: query.Status,
	})
	if err != nil {
		return port.ListProductsResult{}, err
	}

	if len(ids) == 0 {
		return port.ListProductsResult{}, nil
	}

	// Phase 2: Resolve each ID concurrently (reuses GetProduct with cache)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(20)

	results := make([]port.ProductResult, len(ids))
	for i, id := range ids {
		g.Go(func() error {
			result, err := u.getProduct.Execute(ctx, id)
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return port.ListProductsResult{}, err
	}

	return port.ListProductsResult{
		Products:   results,
		NextCursor: nextCursor,
		HasNext:    nextCursor != "",
	}, nil
}
