package port

type ListProductsQuery struct {
	Cursor string
	Limit  int `validate:"gte=1,lte=100"`
	Status string
}
