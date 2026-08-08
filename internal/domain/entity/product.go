package entity

import (
	"time"

	domainerror "github.com/nuzirwan/go-boilerplate/internal/domain/error"
	"github.com/nuzirwan/go-boilerplate/internal/domain/event"
)

type Product struct {
	ID        string
	Name      string
	Price     float64
	Currency  string
	Stock     int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	events    []event.Event
}

func NewProduct(name string, price float64, currency string, stock int) (*Product, error) {
	if price <= 0 {
		return nil, domainerror.NewValidation("price", "price must be positive")
	}

	status := "active"
	if stock == 0 {
		status = "out_of_stock"
	}

	p := &Product{
		ID:        NewID(),
		Name:      name,
		Price:     price,
		Currency:  currency,
		Stock:     stock,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		events:    make([]event.Event, 0, 4),
	}

	p.events = append(p.events, event.NewProductCreated(p.ID, p.Name, p.Price))
	return p, nil
}

func (p *Product) Update(name string, price float64, currency string, stock int) error {
	if price <= 0 {
		return domainerror.NewValidation("price", "price must be positive")
	}

	p.Name = name
	p.Price = price
	p.Currency = currency
	p.Stock = stock
	p.UpdatedAt = time.Now()

	if p.Stock == 0 {
		p.Status = "out_of_stock"
	} else {
		p.Status = "active"
	}

	return nil
}

func (p *Product) ApplyDiscount(percentage float64) error {
	if percentage > 50 {
		return domainerror.New(domainerror.ErrBusinessRule, "discount exceeds maximum of 50%")
	}
	if percentage <= 0 {
		return domainerror.New(domainerror.ErrBusinessRule, "discount must be positive")
	}

	p.Price = p.Price * (1 - percentage/100)
	p.UpdatedAt = time.Now()
	p.events = append(p.events, event.NewProductDiscounted(p.ID, percentage, p.Price))
	return nil
}

func (p *Product) CanBeSold() bool {
	return p.Status == "active" && p.Stock > 0
}

func (p *Product) Events() []event.Event { return p.events }
func (p *Product) ClearEvents()          { p.events = p.events[:0] }
