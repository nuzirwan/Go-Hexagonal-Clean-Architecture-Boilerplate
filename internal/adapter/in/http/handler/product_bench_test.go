package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nuzirwan/go-boilerplate/internal/domain/entity"
	"github.com/nuzirwan/go-boilerplate/internal/domain/event"
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

// --- Stubs (no mock framework, no DB, pure in-memory) ---

type stubRepo struct{}

func (s *stubRepo) Save(_ context.Context, _ *entity.Product) error { return nil }
func (s *stubRepo) FindByID(_ context.Context, _ string) (*entity.Product, error) {
	return &entity.Product{
		ID: "01HZTEST000000000000000000", Name: "Widget",
		Price: 99.99, Currency: "IDR", Stock: 10, Status: "active",
	}, nil
}
func (s *stubRepo) FindAll(_ context.Context, _ port.ProductFilter) ([]*entity.Product, string, error) {
	products := make([]*entity.Product, 0, 10)
	for i := 0; i < 10; i++ {
		products = append(products, &entity.Product{
			ID:   "01HZTEST00000000000000000" + string(rune('0'+i)),
			Name: "Widget", Price: 99.99, Currency: "IDR", Stock: 10, Status: "active",
		})
	}
	return products, "", nil
}
func (s *stubRepo) Delete(_ context.Context, _ string) error { return nil }

type stubPublisher struct{}

func (s *stubPublisher) Publish(_ context.Context, _ ...event.Event) error { return nil }

type stubTxManager struct{}

func (s *stubTxManager) RunInTx(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}

// --- Benchmarks ---

func BenchmarkListProducts(b *testing.B) {
	h := newBenchHandler()
	req := httptest.NewRequest("GET", "/products?limit=10&status=active", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.List(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("got %d", w.Code)
		}
	}
}

func BenchmarkGetProduct(b *testing.B) {
	h := newBenchHandler()
	req := httptest.NewRequest("GET", "/products/01HZTEST000000000000000000", nil)
	req.SetPathValue("id", "01HZTEST000000000000000000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Get(w, req)
	}
}

func BenchmarkCreateProduct(b *testing.B) {
	h := newBenchHandler()
	body := `{"name":"Widget","price":99.99,"currency":"IDR","stock":10}`

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Create(w, req)
	}
}

func BenchmarkUpdateProduct(b *testing.B) {
	h := newBenchHandler()
	body := `{"name":"Widget Pro","price":149.99,"currency":"IDR","stock":20}`

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("PUT", "/products/01HZTEST000000000000000000", strings.NewReader(body))
		req.SetPathValue("id", "01HZTEST000000000000000000")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Update(w, req)
	}
}

func BenchmarkDeleteProduct(b *testing.B) {
	h := newBenchHandler()
	req := httptest.NewRequest("DELETE", "/products/01HZTEST000000000000000000", nil)
	req.SetPathValue("id", "01HZTEST000000000000000000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Delete(w, req)
	}
}

func BenchmarkApplyDiscount(b *testing.B) {
	h := newBenchHandler()
	body := `{"percentage":15}`

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/products/01HZTEST000000000000000000/discount", strings.NewReader(body))
		req.SetPathValue("id", "01HZTEST000000000000000000")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ApplyDiscount(w, req)
	}
}

func newBenchHandler() *ProductHandler {
	repo := &stubRepo{}
	pub := &stubPublisher{}
	tx := &stubTxManager{}
	return NewProductHandler(ProductHandlerDeps{
		CreateProduct: newCreateProductUC(repo, pub, tx),
		GetProduct:    newGetProductUC(repo),
		ListProducts:  newListProductsUC(repo),
		UpdateProduct: newUpdateProductUC(repo, tx),
		DeleteProduct: newDeleteProductUC(repo),
		ApplyDiscount: newApplyDiscountUC(repo, pub, tx),
	})
}

// --- Use case wiring helpers (thin wrappers to avoid importing usecase pkg in test) ---

type listProductsUC struct{ repo *stubRepo }

func newListProductsUC(repo *stubRepo) *listProductsUC { return &listProductsUC{repo: repo} }
func (u *listProductsUC) Execute(ctx context.Context, query port.ListProductsQuery) (port.ListProductsResult, error) {
	products, cursor, err := u.repo.FindAll(ctx, port.ProductFilter{Cursor: query.Cursor, Limit: query.Limit, Status: query.Status})
	if err != nil {
		return port.ListProductsResult{}, err
	}
	results := make([]port.ProductResult, 0, len(products))
	for _, p := range products {
		results = append(results, port.ProductResult{ID: p.ID, Name: p.Name, Price: p.Price, Currency: p.Currency, Stock: p.Stock, Status: p.Status})
	}
	return port.ListProductsResult{Products: results, NextCursor: cursor, HasNext: cursor != ""}, nil
}

type getProductUC struct{ repo *stubRepo }

func newGetProductUC(repo *stubRepo) *getProductUC { return &getProductUC{repo: repo} }
func (u *getProductUC) Execute(ctx context.Context, id string) (port.ProductResult, error) {
	p, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return port.ProductResult{}, err
	}
	return port.ProductResult{ID: p.ID, Name: p.Name, Price: p.Price, Currency: p.Currency, Stock: p.Stock, Status: p.Status}, nil
}

type createProductUC struct{}

func newCreateProductUC(_ *stubRepo, _ *stubPublisher, _ *stubTxManager) *createProductUC {
	return &createProductUC{}
}
func (u *createProductUC) Execute(_ context.Context, _ port.CreateProductCommand) (port.ProductResult, error) {
	return port.ProductResult{}, nil
}

type updateProductUC struct{}

func newUpdateProductUC(_ *stubRepo, _ *stubTxManager) *updateProductUC { return &updateProductUC{} }
func (u *updateProductUC) Execute(_ context.Context, _ port.UpdateProductCommand) (port.ProductResult, error) {
	return port.ProductResult{}, nil
}

type deleteProductUC struct{}

func newDeleteProductUC(_ *stubRepo) *deleteProductUC                { return &deleteProductUC{} }
func (u *deleteProductUC) Execute(_ context.Context, _ string) error { return nil }

type applyDiscountUC struct{}

func newApplyDiscountUC(_ *stubRepo, _ *stubPublisher, _ *stubTxManager) *applyDiscountUC {
	return &applyDiscountUC{}
}
func (u *applyDiscountUC) Execute(_ context.Context, _ port.ApplyDiscountCommand) (port.ProductResult, error) {
	return port.ProductResult{}, nil
}
