package port

//go:generate mockgen -source=inbound.go -destination=../../mocks/mock_inbound.go -package=mocks

import "context"

type CreateProduct interface {
	Execute(ctx context.Context, cmd CreateProductCommand) (ProductResult, error)
}

type GetProduct interface {
	Execute(ctx context.Context, id string) (ProductResult, error)
}

type ListProducts interface {
	Execute(ctx context.Context, query ListProductsQuery) (ListProductsResult, error)
}

type UpdateProduct interface {
	Execute(ctx context.Context, cmd UpdateProductCommand) (ProductResult, error)
}

type DeleteProduct interface {
	Execute(ctx context.Context, id string) error
}

type ApplyDiscount interface {
	Execute(ctx context.Context, cmd ApplyDiscountCommand) (ProductResult, error)
}
