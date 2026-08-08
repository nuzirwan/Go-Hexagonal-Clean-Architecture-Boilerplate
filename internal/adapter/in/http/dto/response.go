package dto

import (
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

type ProductResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Stock     int     `json:"stock"`
	Status    string  `json:"status"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

func ToProductResponse(r port.ProductResult) ProductResponse {
	return ProductResponse{
		ID:        r.ID,
		Name:      r.Name,
		Price:     r.Price,
		Currency:  r.Currency,
		Stock:     r.Stock,
		Status:    r.Status,
		CreatedAt: r.CreatedAt.UnixMilli(),
		UpdatedAt: r.UpdatedAt.UnixMilli(),
	}
}

func ToProductListResponse(results []port.ProductResult) []ProductResponse {
	out := make([]ProductResponse, len(results))
	for i, r := range results {
		out[i] = ToProductResponse(r)
	}
	return out
}
