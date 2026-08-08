package port

import "time"

type ProductResult struct {
	ID        string
	Name      string
	Price     float64
	Currency  string
	Stock     int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListProductsResult struct {
	Products   []ProductResult
	NextCursor string
	HasNext    bool
}
