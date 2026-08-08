package handler

// NOTE: This handler uses generated proto types. Generate with:
//   protoc --go_out=. --go-grpc_out=. api/proto/product/v1/product.proto
// Then uncomment and adjust imports.

import (
	"context"
	"strings"
	"time"

	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductGRPCHandler implements the gRPC ProductService.
// Uncomment proto imports and embed UnimplementedProductServiceServer after code generation.
type ProductGRPCHandler struct {
	// productv1.UnimplementedProductServiceServer
	createProduct port.CreateProduct
	getProduct    port.GetProduct
	applyDiscount port.ApplyDiscount
}

type ProductGRPCHandlerDeps struct {
	CreateProduct port.CreateProduct
	GetProduct    port.GetProduct
	ApplyDiscount port.ApplyDiscount
}

func NewProductGRPCHandler(deps ProductGRPCHandlerDeps) *ProductGRPCHandler {
	return &ProductGRPCHandler{
		createProduct: deps.CreateProduct,
		getProduct:    deps.GetProduct,
		applyDiscount: deps.ApplyDiscount,
	}
}

// CreateProduct handles gRPC CreateProduct call.
// After proto generation, this will implement productv1.ProductServiceServer.
func (h *ProductGRPCHandler) CreateProduct(ctx context.Context, name string, price float64, currency string, stock int) (port.ProductResult, error) {
	result, err := h.createProduct.Execute(ctx, port.CreateProductCommand{
		Name:     name,
		Price:    price,
		Currency: currency,
		Stock:    stock,
	})
	if err != nil {
		return port.ProductResult{}, mapDomainToGRPCError(err)
	}
	return result, nil
}

// GetProduct handles gRPC GetProduct call.
func (h *ProductGRPCHandler) GetProduct(ctx context.Context, id string) (port.ProductResult, error) {
	result, err := h.getProduct.Execute(ctx, id)
	if err != nil {
		return port.ProductResult{}, mapDomainToGRPCError(err)
	}
	return result, nil
}

// ApplyDiscount handles gRPC ApplyDiscount call.
func (h *ProductGRPCHandler) ApplyDiscount(ctx context.Context, productID string, percentage float64) (port.ProductResult, error) {
	result, err := h.applyDiscount.Execute(ctx, port.ApplyDiscountCommand{
		ProductID:  productID,
		Percentage: percentage,
	})
	if err != nil {
		return port.ProductResult{}, mapDomainToGRPCError(err)
	}
	return result, nil
}

func mapDomainToGRPCError(err error) error {
	if err == nil {
		return nil
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, "NOT_FOUND") {
		return status.Error(codes.NotFound, errMsg)
	}
	if strings.Contains(errMsg, "VALIDATION_ERROR") {
		return status.Error(codes.InvalidArgument, errMsg)
	}
	if strings.Contains(errMsg, "CONFLICT") {
		return status.Error(codes.AlreadyExists, errMsg)
	}
	if strings.Contains(errMsg, "BUSINESS_RULE") {
		return status.Error(codes.FailedPrecondition, errMsg)
	}
	return status.Error(codes.Internal, "internal error")
}

// Placeholder to avoid unused import.
var _ = time.RFC3339
