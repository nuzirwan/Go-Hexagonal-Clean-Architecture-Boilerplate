package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	obsadapter "github.com/nuzirwan/go-boilerplate/internal/adapter/out/observability"
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

// Benchmarks with observability decorators wrapping usecases (no-op tracer/meter).

func BenchmarkGetProduct_WithObs(b *testing.B) {
	h := newBenchHandlerWithObs()
	req := httptest.NewRequest("GET", "/products/01HZTEST000000000000000000", nil)
	req.SetPathValue("id", "01HZTEST000000000000000000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Get(w, req)
	}
}

func BenchmarkCreateProduct_WithObs(b *testing.B) {
	h := newBenchHandlerWithObs()
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

func BenchmarkDeleteProduct_WithObs(b *testing.B) {
	h := newBenchHandlerWithObs()
	req := httptest.NewRequest("DELETE", "/products/01HZTEST000000000000000000", nil)
	req.SetPathValue("id", "01HZTEST000000000000000000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Delete(w, req)
	}
}

func BenchmarkListProducts_WithObs(b *testing.B) {
	h := newBenchHandlerWithObs()
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

func newBenchHandlerWithObs() *ProductHandler {
	repo := &stubRepo{}
	pub := &stubPublisher{}
	tx := &stubTxManager{}

	var createProduct port.CreateProduct = newCreateProductUC(repo, pub, tx)
	var getProduct port.GetProduct = newGetProductUC(repo)
	var listProducts port.ListProducts = newListProductsUC(repo)
	var updateProduct port.UpdateProduct = newUpdateProductUC(repo, tx)
	var deleteProduct port.DeleteProduct = newDeleteProductUC(repo)
	var applyDiscount port.ApplyDiscount = newApplyDiscountUC(repo, pub, tx)

	createProduct = obsadapter.WrapCreateProduct(createProduct)
	getProduct = obsadapter.WrapGetProduct(getProduct)
	listProducts = obsadapter.WrapListProducts(listProducts)
	updateProduct = obsadapter.WrapUpdateProduct(updateProduct)
	deleteProduct = obsadapter.WrapDeleteProduct(deleteProduct)
	applyDiscount = obsadapter.WrapApplyDiscount(applyDiscount)

	return NewProductHandler(ProductHandlerDeps{
		CreateProduct: createProduct,
		GetProduct:    getProduct,
		ListProducts:  listProducts,
		UpdateProduct: updateProduct,
		DeleteProduct: deleteProduct,
		ApplyDiscount: applyDiscount,
	})
}
