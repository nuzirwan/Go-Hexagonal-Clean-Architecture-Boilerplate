package event

import "time"

type ProductCreated struct {
	Base
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
}

func NewProductCreated(id, name string, price float64) ProductCreated {
	return ProductCreated{
		Base:      Base{EventType: "product:created", OccurredOn: time.Now()},
		ProductID: id,
		Name:      name,
		Price:     price,
	}
}

type ProductDiscounted struct {
	Base
	ProductID  string  `json:"product_id"`
	Percentage float64 `json:"percentage"`
	NewPrice   float64 `json:"new_price"`
}

func NewProductDiscounted(id string, pct, newPrice float64) ProductDiscounted {
	return ProductDiscounted{
		Base:       Base{EventType: "product:discounted", OccurredOn: time.Now()},
		ProductID:  id,
		Percentage: pct,
		NewPrice:   newPrice,
	}
}
