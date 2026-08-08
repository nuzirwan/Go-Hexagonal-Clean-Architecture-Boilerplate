package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"

	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

// --- CreateProduct ---

type createProductDecorator struct {
	inner port.CreateProduct
}

func WrapCreateProduct(inner port.CreateProduct) port.CreateProduct {
	return &createProductDecorator{inner: inner}
}

func (d *createProductDecorator) Execute(ctx context.Context, cmd port.CreateProductCommand) (port.ProductResult, error) {
	ctx, end := StartOp(ctx, "usecase.CreateProduct",
		attribute.String("name", cmd.Name),
		attribute.Float64("price", cmd.Price),
		attribute.String("currency", cmd.Currency),
	)
	result, err := d.inner.Execute(ctx, cmd)
	end(err, attribute.String("product_id", result.ID))
	return result, err
}

// --- GetProduct ---

type getProductDecorator struct {
	inner port.GetProduct
}

func WrapGetProduct(inner port.GetProduct) port.GetProduct {
	return &getProductDecorator{inner: inner}
}

func (d *getProductDecorator) Execute(ctx context.Context, id string) (port.ProductResult, error) {
	ctx, end := StartOp(ctx, "usecase.GetProduct", attribute.String("id", id))
	result, err := d.inner.Execute(ctx, id)
	end(err)
	return result, err
}

// --- ListProducts ---

type listProductsDecorator struct {
	inner port.ListProducts
}

func WrapListProducts(inner port.ListProducts) port.ListProducts {
	return &listProductsDecorator{inner: inner}
}

func (d *listProductsDecorator) Execute(ctx context.Context, query port.ListProductsQuery) (port.ListProductsResult, error) {
	ctx, end := StartOp(ctx, "usecase.ListProducts",
		attribute.Int("limit", query.Limit),
		attribute.String("cursor", query.Cursor),
	)
	result, err := d.inner.Execute(ctx, query)
	end(err, attribute.Int("count", len(result.Products)))
	return result, err
}

// --- UpdateProduct ---

type updateProductDecorator struct {
	inner port.UpdateProduct
}

func WrapUpdateProduct(inner port.UpdateProduct) port.UpdateProduct {
	return &updateProductDecorator{inner: inner}
}

func (d *updateProductDecorator) Execute(ctx context.Context, cmd port.UpdateProductCommand) (port.ProductResult, error) {
	ctx, end := StartOp(ctx, "usecase.UpdateProduct", attribute.String("id", cmd.ID))
	result, err := d.inner.Execute(ctx, cmd)
	end(err)
	return result, err
}

// --- DeleteProduct ---

type deleteProductDecorator struct {
	inner port.DeleteProduct
}

func WrapDeleteProduct(inner port.DeleteProduct) port.DeleteProduct {
	return &deleteProductDecorator{inner: inner}
}

func (d *deleteProductDecorator) Execute(ctx context.Context, id string) error {
	ctx, end := StartOp(ctx, "usecase.DeleteProduct", attribute.String("id", id))
	err := d.inner.Execute(ctx, id)
	end(err)
	return err
}

// --- ApplyDiscount ---

type applyDiscountDecorator struct {
	inner port.ApplyDiscount
}

func WrapApplyDiscount(inner port.ApplyDiscount) port.ApplyDiscount {
	return &applyDiscountDecorator{inner: inner}
}

func (d *applyDiscountDecorator) Execute(ctx context.Context, cmd port.ApplyDiscountCommand) (port.ProductResult, error) {
	ctx, end := StartOp(ctx, "usecase.ApplyDiscount",
		attribute.String("product_id", cmd.ProductID),
		attribute.Float64("percentage", cmd.Percentage),
	)
	result, err := d.inner.Execute(ctx, cmd)
	end(err, attribute.String("product_id", result.ID))
	return result, err
}

// A is a shorthand for attribute.String with fmt.Sprint for non-string values.
// Use attribute.String/Int/Float64 directly for zero-alloc hot paths.
func A(key string, val any) attribute.KeyValue {
	return attribute.String(key, fmt.Sprint(val))
}
