package usecase

import (
	"github.com/nuzirwan/go-boilerplate/internal/domain/entity"
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

func toResult(p *entity.Product) port.ProductResult {
	return port.ProductResult{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Currency:  p.Currency,
		Stock:     p.Stock,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
