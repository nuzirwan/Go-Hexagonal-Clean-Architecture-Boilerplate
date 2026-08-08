package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"

	"github.com/nuzirwan/go-boilerplate/internal/adapter/in/http/dto"
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
	"github.com/nuzirwan/go-boilerplate/pkg/httputil"
)

type ProductHandler struct {
	createProduct port.CreateProduct
	getProduct    port.GetProduct
	listProducts  port.ListProducts
	updateProduct port.UpdateProduct
	deleteProduct port.DeleteProduct
	applyDiscount port.ApplyDiscount
	validate      *validator.Validate
}

type ProductHandlerDeps struct {
	CreateProduct port.CreateProduct
	GetProduct    port.GetProduct
	ListProducts  port.ListProducts
	UpdateProduct port.UpdateProduct
	DeleteProduct port.DeleteProduct
	ApplyDiscount port.ApplyDiscount
}

func NewProductHandler(deps ProductHandlerDeps) *ProductHandler {
	return &ProductHandler{
		createProduct: deps.CreateProduct,
		getProduct:    deps.GetProduct,
		listProducts:  deps.ListProducts,
		updateProduct: deps.UpdateProduct,
		deleteProduct: deps.DeleteProduct,
		applyDiscount: deps.ApplyDiscount,
		validate:      validator.New(),
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var cmd port.CreateProductCommand
	if err := decodeJSON(r, &cmd); err != nil {
		httputil.Error(w, r, http.StatusBadRequest, "PARSE_ERROR", "invalid request body")
		return
	}
	if err := h.validate.Struct(cmd); err != nil {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	result, err := h.createProduct.Execute(r.Context(), cmd)
	if err != nil {
		httputil.DomainError(w, r, err)
		return
	}
	httputil.Success(w, r, http.StatusCreated, dto.ToProductResponse(result))
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "id is required")
		return
	}
	result, err := h.getProduct.Execute(r.Context(), id)
	if err != nil {
		httputil.DomainError(w, r, err)
		return
	}
	httputil.Success(w, r, http.StatusOK, dto.ToProductResponse(result))
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := port.ListProductsQuery{
		Cursor: r.URL.Query().Get("cursor"),
		Limit:  limit,
		Status: r.URL.Query().Get("status"),
	}
	if err := h.validate.Struct(query); err != nil {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	result, err := h.listProducts.Execute(r.Context(), query)
	if err != nil {
		httputil.DomainError(w, r, err)
		return
	}
	httputil.SuccessList(w, r, dto.ToProductListResponse(result.Products), httputil.Pagination{
		Cursor:  result.NextCursor,
		Limit:   limit,
		HasNext: result.HasNext,
	})
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "id is required")
		return
	}

	var cmd port.UpdateProductCommand
	if err := decodeJSON(r, &cmd); err != nil {
		httputil.Error(w, r, http.StatusBadRequest, "PARSE_ERROR", "invalid request body")
		return
	}
	cmd.ID = id

	if err := h.validate.Struct(cmd); err != nil {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	result, err := h.updateProduct.Execute(r.Context(), cmd)
	if err != nil {
		httputil.DomainError(w, r, err)
		return
	}
	httputil.Success(w, r, http.StatusOK, dto.ToProductResponse(result))
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "id is required")
		return
	}
	if err := h.deleteProduct.Execute(r.Context(), id); err != nil {
		httputil.DomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) ApplyDiscount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "id is required")
		return
	}

	var cmd port.ApplyDiscountCommand
	if err := decodeJSON(r, &cmd); err != nil {
		httputil.Error(w, r, http.StatusBadRequest, "PARSE_ERROR", "invalid request body")
		return
	}
	cmd.ProductID = id

	if err := h.validate.Struct(cmd); err != nil {
		httputil.Error(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	result, err := h.applyDiscount.Execute(r.Context(), cmd)
	if err != nil {
		httputil.DomainError(w, r, err)
		return
	}
	httputil.Success(w, r, http.StatusOK, dto.ToProductResponse(result))
}

func decodeJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return sonic.Unmarshal(body, dst)
}
