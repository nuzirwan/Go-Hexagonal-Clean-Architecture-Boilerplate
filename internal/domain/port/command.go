package port

type CreateProductCommand struct {
	Name     string  `validate:"required,min=1,max=100"`
	Price    float64 `validate:"required,gt=0"`
	Currency string  `validate:"required,len=3"`
	Stock    int     `validate:"gte=0"`
}

type UpdateProductCommand struct {
	ID       string  `validate:"required"`
	Name     string  `validate:"required,min=1,max=100"`
	Price    float64 `validate:"required,gt=0"`
	Currency string  `validate:"required,len=3"`
	Stock    int     `validate:"gte=0"`
}

type ApplyDiscountCommand struct {
	ProductID  string  `validate:"required"`
	Percentage float64 `validate:"required,gt=0,lte=50"`
}
